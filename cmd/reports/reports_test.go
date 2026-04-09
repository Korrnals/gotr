package reports

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Tests for Register ====================

func TestRegister(t *testing.T) {
	root := &cobra.Command{}
	mock := &client.MockClient{}

	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	// Verify reports command is added
	reportsCmd, _, err := root.Find([]string{"reports"})
	assert.NoError(t, err)
	assert.NotNil(t, reportsCmd)

	// Verify subcommands are present
	listCmd, _, _ := root.Find([]string{"reports", "list"})
	assert.NotNil(t, listCmd)

	listCrossCmd, _, _ := root.Find([]string{"reports", "list-cross-project"})
	assert.NotNil(t, listCrossCmd)

	runCmd, _, _ := root.Find([]string{"reports", "run"})
	assert.NotNil(t, runCmd)

	runCrossCmd, _, _ := root.Find([]string{"reports", "run-cross-project"})
	assert.NotNil(t, runCrossCmd)
}

func TestRegister_Help(t *testing.T) {
	root := &cobra.Command{}
	mock := &client.MockClient{}

	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	// Verify calling without arguments shows help
	root.SetArgs([]string{"reports"})
	err := root.Execute()
	assert.NoError(t, err)
}
