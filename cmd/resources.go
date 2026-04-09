package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Korrnals/gotr/internal/debug"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/Korrnals/gotr/pkg/testrailapi"

	"github.com/spf13/cobra"
)

// Global initialization of TestRailAPI structures (initialized once).
var api = testrailapi.New()

// resourcePathsRegistry maps resource names to their API path providers.
// Single source of truth for all resource dispatch in resources.go.
var resourcePathsRegistry = map[string]func() []testrailapi.APIPath{
	"all":            api.Paths,
	"cases":          api.Cases.Paths,
	"casefields":     api.CaseFields.Paths,
	"casetypes":      api.CaseTypes.Paths,
	"configurations": api.Configurations.Paths,
	"projects":       api.Projects.Paths,
	"priorities":     api.Priorities.Paths,
	"runs":           api.Runs.Paths,
	"tests":          api.Tests.Paths,
	"suites":         api.Suites.Paths,
	"sections":       api.Sections.Paths,
	"statuses":       api.Statuses.Paths,
	"milestones":     api.Milestones.Paths,
	"plans":          api.Plans.Paths,
	"results":        api.Results.Paths,
	"resultfields":   api.ResultFields.Paths,
	"reports":        api.Reports.Paths,
	"attachments":    api.Attachments.Paths,
	"users":          api.Users.Paths,
	"roles":          api.Roles.Paths,
	"templates":      api.Templates.Paths,
	"groups":         api.Groups.Paths,
	"sharedsteps":    api.SharedSteps.Paths,
	"variables":      api.Variables.Paths,
	"labels":         api.Labels.Paths,
	"datasets":       api.Datasets.Paths,
	"bdds":           api.BDDs.Paths,
}

// contextKey is an unexported key type, scoped to the cmd package.
type contextKey string

// httpClientKey is used to store/retrieve the HTTP client from context.
const httpClientKey contextKey = "httpClient"

// ValidResources is a dynamically generated list of all resource names.
var ValidResources []string

func init() {
	resources := make([]string, 0, len(resourcePathsRegistry))
	for name, pathsFn := range resourcePathsRegistry {
		if len(pathsFn()) > 0 {
			resources = append(resources, name)
		}
	}
	sort.Strings(resources)
	ValidResources = resources
}

// extractGetEndpointName reliably extracts the name after "/get_".
func extractGetEndpointName(uri string) string {
	// Find the position of "/get_"
	idx := strings.LastIndex(uri, "/get_")
	if idx == -1 {
		return "" // not a standard TestRail GET endpoint
	}

	name := uri[idx+1:] // everything after "/get_"

	// Trim query parameters starting with "&"
	if qIdx := strings.Index(name, "&"); qIdx != -1 {
		name = name[:qIdx]
	}

	// Trim placeholders starting with "{"
	if phIdx := strings.Index(name, "{"); phIdx != -1 {
		name = name[:phIdx]
	}

	// Clean trailing slashes and spaces
	name = strings.Trim(name, "/ ")

	if name == "" || name == "get_" {
		return ""
	}

	return name
}

// getValidGetEndpoints returns all clean GET endpoint names for shell completion.
func getValidGetEndpoints() []string {
	var names []string
	seen := make(map[string]bool)

	for _, p := range api.Paths() {
		if p.Method != "GET" {
			continue
		}
		name := extractGetEndpointName(p.URI)
		if name != "" && !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}

	// Sort for consistent output
	sort.Strings(names)
	return names
}

// extractAllEndpointName extracts a clean endpoint with placeholders (no query or trailing slashes).
func extractAllEndpointName(uri string) string {
	// Strip "index.php?/api/v2/" prefix
	uri = strings.TrimPrefix(uri, "index.php?/api/v2/")

	// Remove query parameters starting with "&"
	if qIdx := strings.Index(uri, "&"); qIdx != -1 {
		uri = uri[:qIdx]
	}

	// Clean trailing slashes and spaces
	uri = strings.Trim(uri, "/ ")

	if uri == "" {
		return ""
	}

	return uri
}

// getResourceEndpoints returns a list of endpoints for the given resource.
func getResourceEndpoints(resource, outputType string) ([]string, error) {
	paths := getResourcePaths(resource)
	if paths == nil {
		ui.Warningf(os.Stdout, "Unknown resource: %s", resource)
		fmt.Println("\nAvailable resources:")
		fmt.Println("  " + strings.Join(ValidResources, ", "))
		return nil, nil
	}

	sort.Slice(paths, func(i, j int) bool {
		return paths[i].URI < paths[j].URI
	})

	return formatEndpoints(paths, resource, outputType)
}

// formatEndpoints renders API paths according to the requested output format.
func formatEndpoints(paths []testrailapi.APIPath, resource, outputType string) ([]string, error) {
	switch outputType {
	case "json":
		data, err := json.MarshalIndent(paths, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal endpoints: %w", err)
		}
		fmt.Println(string(data))
		return nil, nil
	case "short":
		for _, p := range paths {
			fmt.Printf("%s %s\n", p.Method, p.URI)
		}
		return nil, nil
	case "list":
		var endpoints []string
		for _, p := range paths {
			name := extractGetEndpointName(p.URI)
			endpoints = append(endpoints, name)
		}
		return endpoints, nil
	default:
		fmt.Printf("Endpoints for %s (%d):\n\n", resource, len(paths))
		for _, p := range paths {
			fmt.Printf("  %s %s\n      %s\n", p.Method, p.URI, p.Description)
			if len(p.Params) > 0 {
				fmt.Print("      Parameters:\n")
				for name, desc := range p.Params {
					fmt.Printf("        - %s: %s\n", name, desc)
				}
			}
			fmt.Println()
		}
		return nil, nil
	}
}

// getAllShortEndpoints returns all short endpoint names for a resource (GET, POST, DELETE).
func getAllShortEndpoints(resource string) []string {
	paths := getResourcePaths(resource)
	if paths == nil {
		return nil
	}

	var endpoints []string
	seen := make(map[string]bool)
	for _, p := range paths {
		name := extractAllEndpointName(p.URI)
		if name != "" && !seen[name] {
			seen[name] = true
			endpoints = append(endpoints, name)
		}
	}

	sort.Strings(endpoints)
	return endpoints
}

// Wrapper function that returns ALL endpoints for a specific resource (with caching).
var endpointsCache = make(map[string][]string)

// GetEndpoints returns all short endpoint names for a resource (with caching).
func GetEndpoints(resource string) []string {
	if cached, ok := endpointsCache[resource]; ok {
		return cached
	}

	var endpoints []string
	if resource == "all" {
		seen := make(map[string]bool)
		for _, r := range ValidResources {
			if r == "all" {
				continue
			}
			resEndpoints := GetEndpoints(r) // recursive, but cache prevents re-computation
			for _, e := range resEndpoints {
				if !seen[e] {
					seen[e] = true
					endpoints = append(endpoints, e)
				}
			}
		}
		sort.Strings(endpoints)
	} else {
		endpoints = getAllShortEndpoints(resource)
		if endpoints == nil {
			return nil
		}
	}

	endpointsCache[resource] = endpoints
	return endpoints
}

// replaceAllPlaceholders substitutes all known TestRail path placeholders with the given id.
func replaceAllPlaceholders(uri, id string) string {
	placeholders := []string{
		"{project_id}", "{case_id}", "{run_id}", "{test_id}", "{section_id}",
		"{suite_id}", "{milestone_id}", "{plan_id}", "{user_id}", "{role_id}",
		"{group_id}", "{dataset_id}", "{shared_step_id}", "{report_template_id}",
		"{email}",
	}
	for _, ph := range placeholders {
		uri = strings.ReplaceAll(uri, ph, id)
	}
	return uri
}

// buildRequestParams assembles the full endpoint path and query parameters from flags and positional ID.
func buildRequestParams(endpoint, mainID string, cmd *cobra.Command) (fullEndpoint string, queryParams map[string]string, err error) {
	fullEndpoint = endpoint
	queryParams = make(map[string]string)

	// Substitute the main ID (project_id, run_id, etc.)
	if mainID != "" {
		fullEndpoint = replaceAllPlaceholders(fullEndpoint, mainID)
		if !strings.Contains(fullEndpoint, mainID) {
			fullEndpoint += "/" + mainID
		}
		debug.DebugPrint("{resources} - fullEndpoint after ID: %s", fullEndpoint)
	}

	// Query params — only if value is non-empty
	flags := []struct {
		flagName string // Cobra flag name
		queryKey string // TestRail API parameter name
	}{
		{"suite-id", "suite_id"},
		{"section-id", "section_id"},
		{"milestone-id", "milestone_id"},
		{"assignedto-id", "assignedto_id"},
		{"status-id", "status_id"},
		{"priority-id", "priority_id"},
		{"type-id", "type_id"},
		{"created-by", "created_by"},
		{"updated-by", "updated_by"},
		// Add more as needed
	}

	for _, f := range flags {
		if val, err := cmd.Flags().GetString(f.flagName); err == nil && val != "" {
			queryParams[f.queryKey] = val
					debug.DebugPrint("{resources} - Added parameter: %s = %s", f.queryKey, val)
		}
	}

	return fullEndpoint, queryParams, nil
}

// getResourcePaths returns API paths for the given resource name.
func getResourcePaths(resource string) []testrailapi.APIPath {
	if fn, ok := resourcePathsRegistry[resource]; ok {
		return fn()
	}
	return nil
}
