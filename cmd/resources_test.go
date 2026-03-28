package cmd

import (
	"bytes"
	"io"
	"os"
	"sort"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetValidGetEndpoints(t *testing.T) {
	endpoints := getValidGetEndpoints()
	require.NotEmpty(t, endpoints)
	assert.True(t, sort.StringsAreSorted(endpoints))
}

func TestExtractAllEndpointName(t *testing.T) {
	assert.Equal(t, "get_projects", extractAllEndpointName("index.php?/api/v2/get_projects"))
	assert.Equal(t, "get_cases/{project_id}", extractAllEndpointName("index.php?/api/v2/get_cases/{project_id}&suite_id=1"))
	assert.Equal(t, "", extractAllEndpointName("index.php?/api/v2/"))
}

func TestGetAllShortEndpoints(t *testing.T) {
	endpoints := getAllShortEndpoints("projects")
	require.NotEmpty(t, endpoints)
	assert.True(t, sort.StringsAreSorted(endpoints))
}

func TestGetEndpoints(t *testing.T) {
	endpointsCache = make(map[string][]string)

	projects := GetEndpoints("projects")
	require.NotEmpty(t, projects)

	endpointsCache["projects"] = []string{"cached_endpoint"}
	cached := GetEndpoints("projects")
	assert.Equal(t, []string{"cached_endpoint"}, cached)

	all := GetEndpoints("all")
	require.NotEmpty(t, all)
}

func TestReplaceAllPlaceholders(t *testing.T) {
	uri := "get_case/{case_id}/project/{project_id}/user/{user_id}"
	assert.Equal(t, "get_case/10/project/10/user/10", replaceAllPlaceholders(uri, "10"))
}

func TestBuildRequestParams(t *testing.T) {
	cmd := &cobra.Command{Use: "resources"}
	cmd.Flags().String("suite-id", "", "")
	cmd.Flags().String("section-id", "", "")
	cmd.Flags().String("milestone-id", "", "")

	require.NoError(t, cmd.Flags().Set("suite-id", "11"))
	require.NoError(t, cmd.Flags().Set("milestone-id", "22"))

	fullEndpoint, queryParams, err := buildRequestParams("get_cases/{project_id}", "42", cmd)
	require.NoError(t, err)
	assert.Equal(t, "get_cases/42", fullEndpoint)
	assert.Equal(t, "11", queryParams["suite_id"])
	assert.Equal(t, "22", queryParams["milestone_id"])
	assert.Empty(t, queryParams["section_id"])
}

func TestGetResourcePaths(t *testing.T) {
	assert.NotNil(t, getResourcePaths("projects"))
	assert.Nil(t, getResourcePaths("unknown-resource"))
}

func TestGetResourceEndpoints_AllKnownResourcesListMode(t *testing.T) {
	resources := []string{
		"all", "cases", "casefields", "casetypes", "configurations", "projects", "priorities",
		"runs", "tests", "suites", "sections", "statuses", "milestones", "plans", "results",
		"resultfields", "reports", "attachments", "users", "roles", "templates", "groups",
		"sharedsteps", "variables", "labels", "datasets", "bdds",
	}

	for _, resource := range resources {
		t.Run(resource, func(t *testing.T) {
			endpoints, err := getResourceEndpoints(resource, "list")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to format resource list")
			assert.NotNil(t, endpoints)
		})
	}
}

func TestGetResourceEndpoints_UnknownResource(t *testing.T) {
	endpoints, err := getResourceEndpoints("unknown-resource", "list")
	assert.NoError(t, err)
	assert.Nil(t, endpoints)
}

func TestGetResourceEndpoints_JSONAndShortModes(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		_ = r.Close()
	}()

	jsonEndpoints, jsonErr := getResourceEndpoints("projects", "json")
	assert.NoError(t, jsonErr)
	assert.Nil(t, jsonEndpoints)

	shortEndpoints, shortErr := getResourceEndpoints("projects", "short")
	assert.Error(t, shortErr)
	assert.Contains(t, shortErr.Error(), "failed to format short resource list")
	assert.Nil(t, shortEndpoints)

	require.NoError(t, w.Close())

	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, r)
	require.NoError(t, copyErr)
	assert.NotEmpty(t, buf.String())
}
