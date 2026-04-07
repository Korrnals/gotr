package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupExportInputTest() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("project-id", "p", "", "ID проекта (для эндпоинтов с {project_id})")
	return cmd
}

func TestResolveExportInputs_WithArgs(t *testing.T) {
	cmd := setupExportInputTest()

	resource, endpoint, mainID, err := resolveExportInputs(cmd, []string{"cases", "get_cases", "30"})
	assert.NoError(t, err)
	assert.Equal(t, "cases", resource)
	assert.Equal(t, "get_cases", endpoint)
	assert.Equal(t, "30", mainID)
}

func TestResolveExportInputs_WithProjectIDFlagPriority(t *testing.T) {
	cmd := setupExportInputTest()
	_ = cmd.Flags().Set("project-id", "77")

	resource, endpoint, mainID, err := resolveExportInputs(cmd, []string{"cases", "get_cases", "30"})
	assert.NoError(t, err)
	assert.Equal(t, "cases", resource)
	assert.Equal(t, "get_cases", endpoint)
	assert.Equal(t, "77", mainID)
}

func TestResolveExportInputs_Interactive_NoArgs(t *testing.T) {
	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := setupExportInputTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	resource, endpoint, mainID, err := resolveExportInputs(cmd, []string{})
	assert.NoError(t, err)
	assert.NotEmpty(t, resource)
	assert.NotEmpty(t, endpoint)
	assert.Empty(t, mainID)
}

func TestResolveExportInputs_NonInteractive_NoArgs_Error(t *testing.T) {
	cmd := setupExportInputTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))

	_, _, _, err := resolveExportInputs(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestResolveExportInputs_ResourceRequired_NoPrompter(t *testing.T) {
	cmd := setupExportInputTest()

	_, _, _, err := resolveExportInputs(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource required")
}

func TestResolveExportInputs_EndpointRequired_NoPrompter(t *testing.T) {
	cmd := setupExportInputTest()

	_, _, _, err := resolveExportInputs(cmd, []string{"cases"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint required")
}

func TestResolveExportInputs_IDRequired_NonInteractive(t *testing.T) {
	cmd := setupExportInputTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))

	_, _, _, err := resolveExportInputs(cmd, []string{"cases", "get_case/{case_id}"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read id")
}

func TestResolveExportInputs_IDRequired_NoPrompter(t *testing.T) {
	cmd := setupExportInputTest()

	_, _, _, err := resolveExportInputs(cmd, []string{"cases", "get_case/{case_id}"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id required for endpoint")
}

func TestResolveExportInputs_IDFromInteractivePrompt(t *testing.T) {
	p := interactive.NewMockPrompter().WithInputResponses("42")
	cmd := setupExportInputTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	resource, endpoint, mainID, err := resolveExportInputs(cmd, []string{"cases", "get_case/{case_id}"})
	assert.NoError(t, err)
	assert.Equal(t, "cases", resource)
	assert.Equal(t, "get_case/{case_id}", endpoint)
	assert.Equal(t, "42", mainID)
}

func TestResolveExportInputs_IDFromInteractivePrompt_Trimmed(t *testing.T) {
	p := interactive.NewMockPrompter().WithInputResponses("  42  ")
	cmd := setupExportInputTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	resource, endpoint, mainID, err := resolveExportInputs(cmd, []string{"cases", "get_case/{case_id}"})
	assert.NoError(t, err)
	assert.Equal(t, "cases", resource)
	assert.Equal(t, "get_case/{case_id}", endpoint)
	assert.Equal(t, "42", mainID)
}

func TestResolveExportInputs_ResourceLowercasedFromArgs(t *testing.T) {
	cmd := setupExportInputTest()

	resource, endpoint, mainID, err := resolveExportInputs(cmd, []string{"CASES", "get_cases", "30"})
	assert.NoError(t, err)
	assert.Equal(t, "cases", resource)
	assert.Equal(t, "get_cases", endpoint)
	assert.Equal(t, "30", mainID)
}

func TestResolveExportInputs_Interactive_NoGetEndpoints(t *testing.T) {
	// "nonexistent_resource" is not mapped in getResourceEndpoints, returns nil list.
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := setupExportInputTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	_, _, _, err := resolveExportInputs(cmd, []string{"nonexistent_resource"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no export endpoints found")
}

func TestResolveExportInputs_Interactive_ResourceSelectError(t *testing.T) {
	p := interactive.NewMockPrompter() // select queue empty -> error
	cmd := setupExportInputTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	_, _, _, err := resolveExportInputs(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select resource")
}

func TestResolveExportInputs_Interactive_EndpointSelectError(t *testing.T) {
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0}) // только выбор ресурса
	cmd := setupExportInputTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	_, _, _, err := resolveExportInputs(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select endpoint")
}

// ============= LAYER 1 AGGRESSIVE EXPANSION - EXPORT INPUT VALIDATION =============

func TestResolveExportInputs_AllResourceTypes(t *testing.T) {
	resources := []string{"projects", "suites", "cases", "runs", "results", "plans"}
	cmd := setupExportInputTest()

	for _, resource := range resources {
		t.Run(resource, func(t *testing.T) {
			p := interactive.NewMockPrompter().
				WithSelectResponses(interactive.SelectResponse{Index: 0}).
				WithSelectResponses(interactive.SelectResponse{Index: 0})

			cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
		_, endpoint, _, err := resolveExportInputs(cmd, []string{resource})
			assert.NoError(t, err)
			assert.NotEmpty(t, endpoint)
		})
	}
}

func TestResolveExportInputs_IDRequiredEndpoints(t *testing.T) {
	endpoints := []string{
		"get_case/{case_id}",
		"get_run/{run_id}",
		"get_plan/{plan_id}",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			p := interactive.NewMockPrompter().WithInputResponses("123")
			cmd := setupExportInputTest()
			cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

			_, ep, mainID, err := resolveExportInputs(cmd, []string{"cases", endpoint})
			assert.NoError(t, err)
			assert.Equal(t, ep, endpoint)
			assert.Equal(t, "123", mainID)
		})
	}
}

func TestResolveExportInputs_UppercaseNormalization(t *testing.T) {
	cmd := setupExportInputTest()

	testCases := []struct {
		input    string
		expected string
	}{
		{"PROJECTS", "projects"},
		{"Projects", "projects"},
		{"CASES", "cases"},
		{"Cases", "cases"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			resource, _, _, err := resolveExportInputs(cmd, []string{tc.input, "get_projects"})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, resource)
		})
	}
}

func TestResolveExportInputs_ProjectIDFlag_SkipsPromptForIDEndpointInNonInteractive(t *testing.T) {
	cmd := setupExportInputTest()
	_ = cmd.Flags().Set("project-id", "91")
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))

	resource, endpoint, mainID, err := resolveExportInputs(cmd, []string{"cases", "get_case/{case_id}"})
	assert.NoError(t, err)
	assert.Equal(t, "cases", resource)
	assert.Equal(t, "get_case/{case_id}", endpoint)
	assert.Equal(t, "91", mainID)
}

func TestFilterEmpty_RemovesWhitespaceOnlyAndPreservesOrder(t *testing.T) {
	input := []string{" get_cases ", "", "   ", "get_case/{case_id}", "\t", "get_cases/1"}

	got := filterEmpty(input)
	assert.Equal(t, []string{" get_cases ", "get_case/{case_id}", "get_cases/1"}, got)
}

// ============= EXPORT RunE INTEGRATION TESTS =============

// newExportTestServer creates an httptest server that responds with JSON to any request.
func newExportTestServer(t *testing.T, statusCode int, body interface{}) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// setupExportRunECmd creates a cobra.Command with all flags exportCmd.RunE needs,
// and injects the given HTTPClient into its context.
func setupExportRunECmd(t *testing.T, httpClient *client.HTTPClient) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "export"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().StringP("project-id", "p", "", "")
	cmd.Flags().StringP("suite-id", "s", "", "")
	cmd.Flags().String("section-id", "", "")
	cmd.Flags().String("milestone-id", "", "")
	cmd.Flags().String("assignedto-id", "", "")
	cmd.Flags().String("status-id", "", "")
	cmd.Flags().String("priority-id", "", "")
	cmd.Flags().String("type-id", "", "")
	cmd.Flags().String("created-by", "", "")
	cmd.Flags().String("updated-by", "", "")

	ctx := context.WithValue(context.Background(), httpClientKey, httpClient)
	ctx = interactive.WithPrompter(ctx, interactive.NewNonInteractivePrompter())
	cmd.SetContext(ctx)
	return cmd
}

func TestExportRunE_LegacyPath_NoID(t *testing.T) {
	srv := newExportTestServer(t, http.StatusOK, map[string]string{"status": "ok"})
	httpClient, err := client.NewClient(srv.URL, "user", "key", false)
	require.NoError(t, err)

	cmd := setupExportRunECmd(t, httpClient)

	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	err = exportCmd.RunE(cmd, []string{"projects", "get_projects"})
	require.NoError(t, err)

	entries, err := os.ReadDir(filepath.Join(tmpDir, ".testrail"))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].Name(), "projects_")
}

func TestExportRunE_LegacyPath_WithID(t *testing.T) {
	srv := newExportTestServer(t, http.StatusOK, map[string]string{"id": "123"})
	httpClient, err := client.NewClient(srv.URL, "user", "key", false)
	require.NoError(t, err)

	cmd := setupExportRunECmd(t, httpClient)

	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	err = exportCmd.RunE(cmd, []string{"cases", "get_case/{case_id}", "123"})
	require.NoError(t, err)

	entries, err := os.ReadDir(filepath.Join(tmpDir, ".testrail"))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].Name(), "cases_123_")
}

func TestExportRunE_LegacyPath_Quiet(t *testing.T) {
	srv := newExportTestServer(t, http.StatusOK, map[string]string{"ok": "1"})
	httpClient, err := client.NewClient(srv.URL, "user", "key", false)
	require.NoError(t, err)

	cmd := setupExportRunECmd(t, httpClient)
	require.NoError(t, cmd.Flags().Set("quiet", "true"))

	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	err = exportCmd.RunE(cmd, []string{"projects", "get_projects"})
	require.NoError(t, err)

	entries, err := os.ReadDir(filepath.Join(tmpDir, ".testrail"))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestExportRunE_SaveFlag(t *testing.T) {
	srv := newExportTestServer(t, http.StatusOK, map[string]string{"ok": "1"})
	httpClient, err := client.NewClient(srv.URL, "user", "key", false)
	require.NoError(t, err)

	cmd := setupExportRunECmd(t, httpClient)
	require.NoError(t, cmd.Flags().Set("save", "true"))

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	err = exportCmd.RunE(cmd, []string{"projects", "get_projects"})
	require.NoError(t, err)

	// Verify file was saved under ~/.gotr/exports/export/
	exportDir := filepath.Join(tmpDir, ".gotr", "exports", "export")
	entries, err := os.ReadDir(exportDir)
	require.NoError(t, err)
	assert.NotEmpty(t, entries)
}

func TestExportRunE_SaveFlag_Quiet(t *testing.T) {
	srv := newExportTestServer(t, http.StatusOK, map[string]string{"ok": "1"})
	httpClient, err := client.NewClient(srv.URL, "user", "key", false)
	require.NoError(t, err)

	cmd := setupExportRunECmd(t, httpClient)
	require.NoError(t, cmd.Flags().Set("save", "true"))
	require.NoError(t, cmd.Flags().Set("quiet", "true"))

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	err = exportCmd.RunE(cmd, []string{"projects", "get_projects"})
	require.NoError(t, err)
}

func TestExportRunE_ResolveInputsError(t *testing.T) {
	srv := newExportTestServer(t, http.StatusOK, nil)
	httpClient, err := client.NewClient(srv.URL, "user", "key", false)
	require.NoError(t, err)

	cmd := setupExportRunECmd(t, httpClient)

	// No args + non-interactive → resolveExportInputs returns error
	err = exportCmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

func TestExportRunE_GetError_Non200(t *testing.T) {
	srv := newExportTestServer(t, http.StatusNotFound, map[string]string{"error": "not found"})
	httpClient, err := client.NewClient(srv.URL, "user", "key", false)
	require.NoError(t, err)

	cmd := setupExportRunECmd(t, httpClient)

	err = exportCmd.RunE(cmd, []string{"projects", "get_projects"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestExportRunE_MkdirAllError(t *testing.T) {
	srv := newExportTestServer(t, http.StatusOK, map[string]string{"ok": "1"})
	httpClient, err := client.NewClient(srv.URL, "user", "key", false)
	require.NoError(t, err)

	cmd := setupExportRunECmd(t, httpClient)

	// Create a file named ".testrail" so MkdirAll fails
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".testrail"), []byte("block"), 0o644))

	err = exportCmd.RunE(cmd, []string{"projects", "get_projects"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create directory")
}

func TestExportRunE_SaveResponseToFileError(t *testing.T) {
	srv := newExportTestServer(t, http.StatusOK, map[string]string{"ok": "1"})
	httpClient, err := client.NewClient(srv.URL, "user", "key", false)
	require.NoError(t, err)

	cmd := setupExportRunECmd(t, httpClient)

	// Create .testrail as read-only directory
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	testrailDir := filepath.Join(tmpDir, ".testrail")
	require.NoError(t, os.MkdirAll(testrailDir, 0o555))
	t.Cleanup(func() { os.Chmod(testrailDir, 0o755) })

	err = exportCmd.RunE(cmd, []string{"projects", "get_projects"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file export error")
}

func TestExportRunE_SaveFlag_OutputError(t *testing.T) {
	srv := newExportTestServer(t, http.StatusOK, map[string]string{"ok": "1"})
	httpClient, err := client.NewClient(srv.URL, "user", "key", false)
	require.NoError(t, err)

	cmd := setupExportRunECmd(t, httpClient)
	require.NoError(t, cmd.Flags().Set("save", "true"))

	// Set HOME to read-only dir so SaveToFile cannot create .gotr/exports
	tmpDir := t.TempDir()
	require.NoError(t, os.Chmod(tmpDir, 0o555))
	t.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Chmod(tmpDir, 0o755) })

	err = exportCmd.RunE(cmd, []string{"projects", "get_projects"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save error")
}
