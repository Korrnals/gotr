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
