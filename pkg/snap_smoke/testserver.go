//go:build smoke

package snap_smoke

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/pkg/testrailapi"
)

// ---------------------------------------------------------------------------
// FakeTestRail — in-memory TestRail API v2 server for smoke tests.
// Routes are compiled from pkg/testrailapi APIPath catalog, ensuring the mock
// matches the official endpoint structure.
// ---------------------------------------------------------------------------

// routeHandler is a handler for a matched API route.
type routeHandler func(w http.ResponseWriter, r *http.Request, params map[string]string)

// compiledRoute is an APIPath URI compiled into a regex for the path part,
// plus optional query-param templates extracted from "&key={param}" segments.
type compiledRoute struct {
	method         string
	pathPattern    *regexp.Regexp
	queryTemplates map[string]string // query key → param name (e.g. "suite_id" → "suite_id")
	handler        routeHandler
	uri            string // original URI for diagnostics
}

// FakeTestRail is an in-memory TestRail API v2 server backed by httptest.Server.
// Endpoint routing is derived from [testrailapi.APIPath] definitions.
type FakeTestRail struct {
	Server *httptest.Server

	mu         sync.Mutex
	cases      map[int64]*data.Case
	sections   map[int64]*data.Section
	nextCaseID int64
	nextSectID int64

	routes []compiledRoute
}

// templateParamRe matches {param_name} in APIPath URI templates.
var templateParamRe = regexp.MustCompile(`\{(\w+)\}`)

// NewFakeTestRail starts a new in-memory TestRail server with routes compiled
// from pkg/testrailapi endpoint catalog.
// Call .Close() when done (usually via t.Cleanup).
func NewFakeTestRail() *FakeTestRail {
	f := &FakeTestRail{
		cases:      make(map[int64]*data.Case),
		sections:   make(map[int64]*data.Section),
		nextCaseID: 1000,
		nextSectID: 100,
	}

	// Build handler lookup keyed by action name (e.g. "get_case").
	handlers := map[string]routeHandler{
		"get_case":     f.handleGetCase,
		"add_case":     f.handleAddCase,
		"update_case":  f.handleUpdateCase,
		"delete_case":  f.handleDeleteCase,
		"get_sections": f.handleGetSections,
		"add_section":  f.handleAddSection,
	}

	// Compile routes from the official API catalog.
	api := testrailapi.New()
	for _, p := range api.Paths() {
		route := compileAPIPath(p)
		if route == nil {
			continue
		}
		// Attach handler if we support this action.
		action := extractAction(p.URI)
		if h, ok := handlers[action]; ok {
			route.handler = h
		}
		f.routes = append(f.routes, *route)
	}

	// Longer patterns first → more specific routes match before shorter ones.
	sort.Slice(f.routes, func(i, j int) bool {
		return len(f.routes[i].pathPattern.String()) > len(f.routes[j].pathPattern.String())
	})

	f.Server = httptest.NewServer(http.HandlerFunc(f.handler))
	return f
}

// Close shuts down the underlying httptest.Server.
func (f *FakeTestRail) Close() {
	f.Server.Close()
}

// URL returns the server base URL (for client construction).
func (f *FakeTestRail) URL() string {
	return f.Server.URL
}

// SeedSection pre-populates a section so tests can reference it immediately.
func (f *FakeTestRail) SeedSection(id int64, name string, suiteID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sections[id] = &data.Section{
		ID:      id,
		Name:    name,
		SuiteID: suiteID,
	}
	if id >= f.nextSectID {
		f.nextSectID = id + 1
	}
}

// ---------------------------------------------------------------------------
// route compilation from testrailapi.APIPath
// ---------------------------------------------------------------------------

// compileAPIPath converts an APIPath into a compiled route.
// The URI is split into path (regex) and query-param templates (key lookup).
func compileAPIPath(p testrailapi.APIPath) *compiledRoute {
	const prefix = "index.php?/api/v2/"
	if !strings.HasPrefix(p.URI, prefix) {
		return nil
	}
	action := p.URI[len(prefix):]

	// Split path from query template segments at first "&".
	pathPart := action
	var queryParts []string
	if idx := strings.Index(action, "&"); idx != -1 {
		pathPart = action[:idx]
		queryParts = strings.Split(action[idx+1:], "&")
	}

	// Compile the path part into a regex with named capture groups.
	pattern := templateParamRe.ReplaceAllStringFunc(pathPart, func(match string) string {
		name := match[1 : len(match)-1]
		return fmt.Sprintf(`(?P<%s>[^/&]+)`, name)
	})

	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return nil
	}

	// Parse query template segments: "suite_id={suite_id}" → key="suite_id", param="suite_id".
	qt := make(map[string]string)
	for _, part := range queryParts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			m := templateParamRe.FindStringSubmatch(kv[1])
			if len(m) == 2 {
				qt[kv[0]] = m[1]
			}
		}
	}

	return &compiledRoute{
		method:         p.Method,
		pathPattern:    re,
		queryTemplates: qt,
		uri:            p.URI,
	}
}

// extractAction returns the action name (e.g. "get_case") from a full URI.
func extractAction(uri string) string {
	const prefix = "index.php?/api/v2/"
	rest := strings.TrimPrefix(uri, prefix)
	if idx := strings.Index(rest, "/"); idx != -1 {
		return rest[:idx]
	}
	if idx := strings.Index(rest, "&"); idx != -1 {
		return rest[:idx]
	}
	return rest
}

// ---------------------------------------------------------------------------
// request routing
// ---------------------------------------------------------------------------

// handler routes incoming requests via compiled routes from testrailapi.APIPath.
//
// TestRail URL format: /index.php?/api/v2/{action}/{id}&extra_params
// After "?" the remainder goes into RawQuery.
func (f *FakeTestRail) handler(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.RawQuery
	const prefix = "/api/v2/"
	if !strings.HasPrefix(raw, prefix) {
		writeError(w, http.StatusNotFound, "unknown endpoint")
		return
	}
	rest := raw[len(prefix):]

	// Split endpoint path from additional query params.
	endpointPath := rest
	queryStr := ""
	if idx := strings.Index(rest, "&"); idx != -1 {
		endpointPath = rest[:idx]
		queryStr = rest[idx+1:]
	}

	// Parse query params into a map for template matching.
	queryMap := make(map[string]string)
	if queryStr != "" {
		for _, part := range strings.Split(queryStr, "&") {
			if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
				queryMap[kv[0]] = kv[1]
			}
		}
	}

	for _, route := range f.routes {
		if route.method != r.Method {
			continue
		}
		match := route.pathPattern.FindStringSubmatch(endpointPath)
		if match == nil {
			continue
		}

		// Check required query templates are present.
		allMatch := true
		for key := range route.queryTemplates {
			if _, ok := queryMap[key]; !ok {
				allMatch = false
				break
			}
		}
		if !allMatch {
			continue
		}

		if route.handler == nil {
			writeError(w, http.StatusNotImplemented,
				fmt.Sprintf("endpoint %s %s recognized but not implemented in mock", route.method, route.uri))
			return
		}

		// Extract named params from path regex.
		params := map[string]string{"_raw_query": rest}
		for i, name := range route.pathPattern.SubexpNames() {
			if i != 0 && name != "" && i < len(match) {
				params[name] = match[i]
			}
		}
		// Extract query template params.
		for key, paramName := range route.queryTemplates {
			params[paramName] = queryMap[key]
		}

		route.handler(w, r, params)
		return
	}

	writeError(w, http.StatusNotFound, fmt.Sprintf("no matching route for %s %s", r.Method, rest))
}

// ---------------------------------------------------------------------------
// case handlers
// ---------------------------------------------------------------------------

func (f *FakeTestRail) handleGetCase(w http.ResponseWriter, _ *http.Request, params map[string]string) {
	id, err := strconv.ParseInt(params["case_id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid case_id")
		return
	}
	f.mu.Lock()
	c, ok := f.cases[id]
	f.mu.Unlock()
	if !ok {
		writeError(w, http.StatusBadRequest, "Field :case_id is not a valid test case.")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (f *FakeTestRail) handleAddCase(w http.ResponseWriter, r *http.Request, params map[string]string) {
	sectionID, err := strconv.ParseInt(params["section_id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid section_id")
		return
	}

	var req data.AddCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	c := &data.Case{
		ID:         f.nextCaseID,
		Title:      req.Title,
		SectionID:  sectionID,
		TypeID:     req.TypeID,
		PriorityID: req.PriorityID,
		Estimate:   req.Estimate,
		Refs:       req.Refs,
		CreatedBy:  1,
		CreatedOn:  time.Now().Unix(),
		UpdatedBy:  1,
		UpdatedOn:  time.Now().Unix(),
	}
	f.nextCaseID++
	f.cases[c.ID] = c
	writeJSON(w, http.StatusOK, c)
}

func (f *FakeTestRail) handleUpdateCase(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id, err := strconv.ParseInt(params["case_id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid case_id")
		return
	}

	var req data.UpdateCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	c, ok := f.cases[id]
	if !ok {
		writeError(w, http.StatusBadRequest, "Field :case_id is not a valid test case.")
		return
	}

	if req.Title != nil {
		c.Title = *req.Title
	}
	if req.PriorityID != nil {
		c.PriorityID = *req.PriorityID
	}
	if req.TypeID != nil {
		c.TypeID = *req.TypeID
	}
	if req.Estimate != nil {
		c.Estimate = *req.Estimate
	}
	if req.Refs != nil {
		c.Refs = *req.Refs
	}
	if req.CustomPreconds != nil {
		c.CustomPreconds = *req.CustomPreconds
	}
	if req.CustomSteps != nil {
		c.CustomSteps = *req.CustomSteps
	}
	if req.CustomExpected != nil {
		c.CustomExpected = *req.CustomExpected
	}
	if req.SectionID != nil {
		c.SectionID = *req.SectionID
	}
	if req.MilestoneID != nil {
		c.MilestoneID = *req.MilestoneID
	}
	if req.TemplateID != nil {
		c.TemplateID = *req.TemplateID
	}
	c.UpdatedOn = time.Now().Unix()

	writeJSON(w, http.StatusOK, c)
}

func (f *FakeTestRail) handleDeleteCase(w http.ResponseWriter, _ *http.Request, params map[string]string) {
	id, err := strconv.ParseInt(params["case_id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid case_id")
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.cases[id]; !ok {
		writeError(w, http.StatusBadRequest, "Field :case_id is not a valid test case.")
		return
	}
	delete(f.cases, id)
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------
// section handlers
// ---------------------------------------------------------------------------

func (f *FakeTestRail) handleGetSections(w http.ResponseWriter, _ *http.Request, params map[string]string) {
	// suite_id is extracted by router from query template.
	var filterSuiteID int64
	if v, ok := params["suite_id"]; ok {
		filterSuiteID, _ = strconv.ParseInt(v, 10, 64)
	}

	// Parse offset/limit from extra query params for pagination.
	offset, limit := parsePagination(params["_raw_query"])

	f.mu.Lock()
	defer f.mu.Unlock()

	var all []data.Section
	for _, s := range f.sections {
		if filterSuiteID != 0 && s.SuiteID != filterSuiteID {
			continue
		}
		all = append(all, *s)
	}

	writePaginatedJSON(w, http.StatusOK, "sections", all, offset, limit)
}

func (f *FakeTestRail) handleAddSection(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var req data.AddSectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	s := &data.Section{
		ID:          f.nextSectID,
		Name:        req.Name,
		Description: req.Description,
		SuiteID:     req.SuiteID,
		ParentID:    req.ParentID,
	}
	f.nextSectID++
	f.sections[s.ID] = s
	writeJSON(w, http.StatusOK, s)
}

// ---------------------------------------------------------------------------
// JSON helpers
// ---------------------------------------------------------------------------

// defaultPageLimit mirrors TestRail's default pagination limit.
const defaultPageLimit = 250

// parsePagination extracts offset and limit from the raw query string.
func parsePagination(rawQuery string) (offset, limit int) {
	limit = defaultPageLimit
	for _, part := range strings.Split(rawQuery, "&") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "offset":
			offset, _ = strconv.Atoi(kv[1])
		case "limit":
			limit, _ = strconv.Atoi(kv[1])
		}
	}
	if limit <= 0 {
		limit = defaultPageLimit
	}
	return offset, limit
}

// writePaginatedJSON writes a TestRail-style paginated wrapper response:
//
//	{"offset":0, "limit":250, "size":N, "_links":{...}, "<itemsField>":[...]}
func writePaginatedJSON(w http.ResponseWriter, code int, itemsField string, items interface{}, offset, limit int) {
	// Compute the page slice from the full items list.
	type slicer interface{ Len() int }
	var total int
	var page interface{}

	switch v := items.(type) {
	case []data.Section:
		total = len(v)
		end := offset + limit
		if end > total {
			end = total
		}
		start := offset
		if start > total {
			start = total
		}
		page = v[start:end]
	default:
		// Fallback: no pagination slicing.
		page = items
	}

	env := map[string]interface{}{
		"offset":     offset,
		"limit":      limit,
		"size":       total,
		itemsField:   page,
		"_links": map[string]interface{}{
			"next": nil,
			"prev": nil,
		},
	}

	if offset+limit < total {
		env["_links"] = map[string]interface{}{
			"next": fmt.Sprintf("/api/v2/get_%s&offset=%d&limit=%d", itemsField, offset+limit, limit),
			"prev": nil,
		}
	}

	writeJSON(w, code, env)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
