package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
)

// decodeCasesResponse decodes a get_cases response that can be either:
// - Paginated wrapper (TestRail 6.7+): {"offset":0, "limit":250, "size":N, "cases":[...]}
// - Flat array (older TestRail): [case1, case2, ...]
//
// Returns the slice of cases regardless of format.
func decodeCasesResponse(body []byte) ([]data.Case, error) {
	cases, _, err := decodeCasesResponseWithSize(body)
	return cases, err
}

// decodeCasesResponseWithSize decodes a get_cases response and returns (cases, totalSize, error).
// totalSize is the "size" field from the paginated wrapper (-1 if flat array or unavailable).
func decodeCasesResponseWithSize(body []byte) ([]data.Case, int64, error) {
	if len(body) == 0 {
		return nil, -1, nil
	}

	// Detect format by first non-whitespace byte
	for _, b := range body {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			// Paginated wrapper: {"offset":0, "limit":250, "size":N, "cases":[...]}
			var paginated data.PaginatedCasesResponse
			if err := json.Unmarshal(body, &paginated); err != nil {
				return nil, -1, fmt.Errorf("decode paginated response: %w", err)
			}
			totalSize := paginated.Size
			if totalSize == 0 && len(paginated.Cases) > 0 {
				totalSize = -1 // size field not set, mark unknown
			}
			return paginated.Cases, totalSize, nil
		case '[':
			// Flat array: [case1, case2, ...]
			var cases []data.Case
			if err := json.Unmarshal(body, &cases); err != nil {
				return nil, -1, fmt.Errorf("decode flat response: %w", err)
			}
			return cases, -1, nil
		default:
			return nil, -1, fmt.Errorf("unexpected response format (starts with %q)", string([]byte{b}))
		}
	}

	return nil, -1, nil
}

// GetCases fetches **all** cases for a project (with pagination).
// suiteID and sectionID are optional (0 = not used).
// Returns the full list of cases.
func (c *HTTPClient) GetCases(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
	return c.GetCasesWithProgress(ctx, projectID, suiteID, sectionID, nil)
}

// GetCasesPage fetches a single page of cases at the given offset/limit.
// Useful for targeted retries of failed pages without re-fetching everything.
func (c *HTTPClient) GetCasesPage(ctx context.Context, projectID, suiteID int64, offset, limit int) (data.GetCasesResponse, error) {
	endpoint := fmt.Sprintf("get_cases/%d", projectID)
	query := map[string]string{
		"suite_id": fmt.Sprintf("%d", suiteID),
		"offset":   fmt.Sprintf("%d", offset),
		"limit":    fmt.Sprintf("%d", limit),
	}

	resp, err := c.Get(ctx, endpoint, query)
	if err != nil {
		return nil, fmt.Errorf("request error GetCasesPage project=%d suite=%d offset=%d limit=%d: %w",
			projectID, suiteID, offset, limit, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response body read error GetCasesPage project=%d suite=%d offset=%d limit=%d: %w",
			projectID, suiteID, offset, limit, err)
	}

	page, err := decodeCasesResponse(body)
	if err != nil {
		return nil, fmt.Errorf("decode error cases page project=%d suite=%d offset=%d limit=%d: %w",
			projectID, suiteID, offset, limit, err)
	}

	return data.GetCasesResponse(page), nil
}

// GetCasesWithProgress fetches **all** cases for a project with progress tracking.
// monitor is called after each page (every 250 cases).
func (c *HTTPClient) GetCasesWithProgress(ctx context.Context, projectID, suiteID, sectionID int64, monitor ProgressMonitor) (data.GetCasesResponse, error) {
	var all data.GetCasesResponse
	offset := int64(0)
	limit := int64(250)

	for {
		endpoint := fmt.Sprintf("get_cases/%d", projectID)
		query := map[string]string{
			"offset": fmt.Sprintf("%d", offset),
			"limit":  fmt.Sprintf("%d", limit),
		}
		if suiteID != 0 {
			query["suite_id"] = fmt.Sprintf("%d", suiteID)
		}
		if sectionID != 0 {
			query["section_id"] = fmt.Sprintf("%d", sectionID)
		}

		resp, err := c.Get(ctx, endpoint, query)
		if err != nil {
			return nil, fmt.Errorf("request error GetCases for project %d: %w", projectID, err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("response body read error (offset=%d): %w", offset, readErr)
		}

		page, decErr := decodeCasesResponse(body)
		if decErr != nil {
			return nil, fmt.Errorf("decode error cases page (offset=%d): %w", offset, decErr)
		}

		all = append(all, page...)

		// Update progress after each page
		if monitor != nil {
			monitor.Increment()
		}

		// Break if we got fewer items than limit, or if page is empty (safety check)
		if len(page) == 0 || len(page) < int(limit) {
			break
		}

		offset += limit
	}

	return all, nil
}

// GetCase fetches a single case by ID.
func (c *HTTPClient) GetCase(ctx context.Context, caseID int64) (*data.Case, error) {
	endpoint := fmt.Sprintf("get_case/%d", caseID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	var kase data.Case
	if err := json.NewDecoder(resp.Body).Decode(&kase); err != nil {
		return nil, fmt.Errorf("decode error case %d: %w", caseID, err)
	}

	return &kase, nil
}

// GetHistoryForCase fetches the change history for a case.
func (c *HTTPClient) GetHistoryForCase(ctx context.Context, caseID int64) (*data.GetHistoryForCaseResponse, error) {
	endpoint := fmt.Sprintf("get_history_for_case/%d", caseID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetHistoryForCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	var result data.GetHistoryForCaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error history case %d: %w", caseID, err)
	}

	return &result, nil
}

// AddCase creates a new test case in a section.
// Requires sectionID and Title in the request.
func (c *HTTPClient) AddCase(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("add_case/%d", sectionID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddCase in section %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	var result data.Case
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error created case: %w", err)
	}

	return &result, nil
}

// UpdateCase updates an existing test case.
// Supports partial updates.
func (c *HTTPClient) UpdateCase(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("update_case/%d", caseID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	var result data.Case
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error updated case %d: %w", caseID, err)
	}

	return &result, nil
}

// UpdateCases performs a bulk update of cases in a suite.
func (c *HTTPClient) UpdateCases(ctx context.Context, suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("update_cases/%d", suiteID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateCases in suite %d: %w", suiteID, err)
	}
	defer resp.Body.Close()

	var result data.GetCasesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error response bulk update: %w", err)
	}

	return &result, nil
}

// DeleteCase deletes a case by ID.
// This action is irreversible.
func (c *HTTPClient) DeleteCase(ctx context.Context, caseID int64) error {
	endpoint := fmt.Sprintf("delete_case/%d", caseID)
	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("request error DeleteCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	return nil
}

// DeleteCases performs a bulk deletion of cases in a suite.
func (c *HTTPClient) DeleteCases(ctx context.Context, suiteID int64, req *data.DeleteCasesRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("delete_cases/%d", suiteID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return fmt.Errorf("request error DeleteCases in suite %d: %w", suiteID, err)
	}
	defer resp.Body.Close()

	return nil
}

// GetCaseTypes fetches all available case types.
func (c *HTTPClient) GetCaseTypes(ctx context.Context) (data.GetCaseTypesResponse, error) {
	endpoint := "get_case_types"
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetCaseTypes: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetCaseTypesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error case types: %w", err)
	}

	return result, nil
}

// GetCaseFields fetches all available case fields.
func (c *HTTPClient) GetCaseFields(ctx context.Context) (data.GetCaseFieldsResponse, error) {
	endpoint := "get_case_fields"
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetCaseFields: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetCaseFieldsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error case fields: %w", err)
	}

	return result, nil
}

// AddCaseField creates a new case field.
func (c *HTTPClient) AddCaseField(ctx context.Context, req *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := "add_case_field"
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddCaseField: %w", err)
	}
	defer resp.Body.Close()

	var result data.AddCaseFieldResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error created field case: %w", err)
	}

	return &result, nil
}

// DiffCasesData compares cases from two projects by the specified field.
// Returns a DiffCasesResponse with the differences.
// Uses parallel loading for speed.
func (c *HTTPClient) DiffCasesData(ctx context.Context, pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
	// Parallel loading of cases from both projects
	type result struct {
		cases data.GetCasesResponse
		err   error
		pid   int64
	}

	resultChan := make(chan result, 2)

	go func() {
		cases, err := c.GetCases(ctx, pid1, 0, 0)
		resultChan <- result{cases, err, pid1}
	}()

	go func() {
		cases, err := c.GetCases(ctx, pid2, 0, 0)
		resultChan <- result{cases, err, pid2}
	}()

	var cases1, cases2 data.GetCasesResponse
	for i := 0; i < 2; i++ {
		res := <-resultChan
		if res.err != nil {
			return nil, fmt.Errorf("failed to get cases from project %d: %w", res.pid, res.err)
		}
		if res.pid == pid1 {
			cases1 = res.cases
		} else {
			cases2 = res.cases
		}
	}

	firstCases := make(map[int64]data.Case)
	for _, c := range cases1 {
		firstCases[c.ID] = c
	}

	secondCases := make(map[int64]data.Case)
	for _, c := range cases2 {
		secondCases[c.ID] = c
	}

	diffResult := &data.DiffCasesResponse{}

	// Only in first project
	for id, c := range firstCases {
		if _, ok := secondCases[id]; !ok {
			diffResult.OnlyInFirst = append(diffResult.OnlyInFirst, c)
		}
	}

	// Only in second project
	for id, c := range secondCases {
		if _, ok := firstCases[id]; !ok {
			diffResult.OnlyInSecond = append(diffResult.OnlyInSecond, c)
		}
	}

	// Differ by the specified field
	for id, c1 := range firstCases {
		if c2, ok := secondCases[id]; ok {
			if !casesEqualByField(c1, c2, field) {
				diffResult.DiffByField = append(diffResult.DiffByField, struct {
					CaseID int64     `json:"case_id"`
					First  data.Case `json:"first"`
					Second data.Case `json:"second"`
				}{id, c1, c2})
			}
		}
	}

	return diffResult, nil
}

// casesEqualByField compares two cases by the specified field.
func casesEqualByField(c1, c2 data.Case, field string) bool {
	switch field {
	case "title":
		return c1.Title == c2.Title
	case "priority_id":
		return c1.PriorityID == c2.PriorityID
	case "custom_preconds":
		return c1.CustomPreconds == c2.CustomPreconds
	case "id":
		return c1.ID == c2.ID
	case "suite_id":
		return c1.SuiteID == c2.SuiteID
	case "created_by":
		return c1.CreatedBy == c2.CreatedBy
	case "section_id":
		return c1.SectionID == c2.SectionID
	default:
		return false
	}
}

// CopyCasesToSection copies cases to the specified section.
// POST index.php?/api/v2/copy_cases_to_section/:section_id
func (c *HTTPClient) CopyCasesToSection(ctx context.Context, sectionID int64, req *data.CopyCasesRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("copy_cases_to_section/%d", sectionID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return fmt.Errorf("request error CopyCasesToSection for section %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	return nil
}

// MoveCasesToSection moves cases to the specified section.
// POST index.php?/api/v2/move_cases_to_section/:section_id
func (c *HTTPClient) MoveCasesToSection(ctx context.Context, sectionID int64, req *data.MoveCasesRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("move_cases_to_section/%d", sectionID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return fmt.Errorf("request error MoveCasesToSection for section %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	return nil
}

// GetCasesParallelCtx fetches cases from multiple suites in parallel (Stage 6.7).
// Uses streaming parallelization without a pre-count step.
// For progress display, set config.Reporter (implements concurrency.PaginatedProgressReporter).
func (c *HTTPClient) GetCasesParallelCtx(
	ctx context.Context,
	projectID int64,
	suiteIDs []int64,
	config *concurrency.ControllerConfig,
) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
	if len(suiteIDs) == 0 {
		return data.GetCasesResponse{}, &concurrency.ExecutionResult{Cases: []data.Case{}}, nil
	}

	if config == nil {
		config = concurrency.DefaultControllerConfig()
	}

	// Create tasks from suiteIDs
	tasks := make([]concurrency.SuiteTask, len(suiteIDs))
	for i, sid := range suiteIDs {
		tasks[i] = concurrency.SuiteTask{
			SuiteID:   sid,
			ProjectID: projectID,
		}
	}

	// Create fetcher implementation
	fetcher := &casesFetcher{client: c}

	// Execute parallel fetching (Reporter is in config)
	controller := concurrency.NewController(config)
	result, err := controller.Execute(ctx, tasks, fetcher, nil)

	if err != nil && len(result.Cases) == 0 {
		return nil, result, err
	}

	return data.GetCasesResponse(result.Cases), result, nil
}

// casesFetcher implements concurrency.SuiteFetcher for cases
type casesFetcher struct {
	client *HTTPClient
}

// FetchPageCtx fetches a single page of cases.
// client.Get() already checks StatusCode != 200 and returns a formatted error,
// so no duplicate status check is needed here.
// Returns (cases, totalSize, error). totalSize comes from API "size" field (-1 if unavailable).
func (f *casesFetcher) FetchPageCtx(ctx context.Context, req concurrency.PageRequest) ([]data.Case, int64, error) {
	endpoint := fmt.Sprintf("get_cases/%d", req.ProjectID)
	query := map[string]string{
		"suite_id": fmt.Sprintf("%d", req.SuiteID),
		"offset":   fmt.Sprintf("%d", req.Offset),
		"limit":    fmt.Sprintf("%d", req.Limit),
	}

	resp, err := f.client.Get(ctx, endpoint, query)
	if err != nil {
		return nil, -1, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("read body error: %w", err)
	}

	return decodeCasesResponseWithSize(body)
}
