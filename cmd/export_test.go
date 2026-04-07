package cmd

import (
	"testing"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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
