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
