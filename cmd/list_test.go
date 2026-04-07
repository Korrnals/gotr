package cmd

import (
	"testing"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func setupListTest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   listCmd.Use,
		Short: listCmd.Short,
		Long:  listCmd.Long,
		Args:  listCmd.Args,
		RunE:  listCmd.RunE,
	}
	cmd.Flags().Bool("json", false, "Вывести в формате JSON")
	cmd.Flags().Bool("short", false, "Краткий вывод (только URI)")
	return cmd
}

func TestList_WithArg_Success(t *testing.T) {
	cmd := setupListTest()
	cmd.SetArgs([]string{"projects"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestList_AutoSelectResource_Success(t *testing.T) {
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := setupListTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestList_NonInteractive_NoArgs_Error(t *testing.T) {
	cmd := setupListTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestList_WithArg_JSONOutput_Success(t *testing.T) {
	cmd := setupListTest()
	cmd.SetArgs([]string{"projects", "--json"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestList_WithArg_ShortOutput_Success(t *testing.T) {
	cmd := setupListTest()
	cmd.SetArgs([]string{"projects", "--short"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestList_TooManyArgs_Error(t *testing.T) {
	cmd := setupListTest()
	cmd.SetArgs([]string{"projects", "extra"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestList_AutoSelectResource_SelectError(t *testing.T) {
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: -1})
	cmd := setupListTest()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select resource")
}

// TestList_NoArgs_NoPrompterInContext_Error covers the branch where
// no args are provided and no prompter is in the context at all.
func TestList_NoArgs_NoPrompterInContext_Error(t *testing.T) {
	cmd := setupListTest()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource required")
}
