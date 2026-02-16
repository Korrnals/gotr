package cases

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	// Create a root command
	rootCmd := &cobra.Command{Use: "gotr"}

	// Mock getClient function
	mockGetClient := func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	}

	// Register cases commands
	Register(rootCmd, mockGetClient)

	// Find cases command
	var casesCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "cases" {
			casesCmd = cmd
			break
		}
	}

	// Verify cases command exists
	assert.NotNil(t, casesCmd, "cases command should be registered")
	assert.Equal(t, "cases", casesCmd.Use)
	assert.NotEmpty(t, casesCmd.Short)
	assert.NotEmpty(t, casesCmd.Long)

	// Verify all subcommands are registered
	subcommands := casesCmd.Commands()
	assert.Len(t, subcommands, 6, "should have 6 subcommands")

	// Check subcommand names
	subNames := make(map[string]bool)
	for _, sub := range subcommands {
		subNames[sub.Name()] = true
	}

	expectedSubcommands := []string{"add", "get", "list", "update", "delete", "bulk"}
	for _, expected := range expectedSubcommands {
		assert.True(t, subNames[expected], "subcommand %s should be registered", expected)
	}
}

func TestRegister_BulkSubcommands(t *testing.T) {
	// Create a root command
	rootCmd := &cobra.Command{Use: "gotr"}

	// Mock getClient function
	mockGetClient := func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	}

	// Register cases commands
	Register(rootCmd, mockGetClient)

	// Find cases command
	var casesCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "cases" {
			casesCmd = cmd
			break
		}
	}
	assert.NotNil(t, casesCmd)

	// Find bulk command
	var bulkCmd *cobra.Command
	for _, cmd := range casesCmd.Commands() {
		if cmd.Name() == "bulk" {
			bulkCmd = cmd
			break
		}
	}
	assert.NotNil(t, bulkCmd, "bulk command should be registered")

	// Verify bulk subcommands
	bulkSubcommands := bulkCmd.Commands()
	assert.Len(t, bulkSubcommands, 4, "bulk should have 4 subcommands")

	bulkSubNames := make(map[string]bool)
	for _, sub := range bulkSubcommands {
		bulkSubNames[sub.Name()] = true
	}

	expectedBulkSubcommands := []string{"update", "delete", "copy", "move"}
	for _, expected := range expectedBulkSubcommands {
		assert.True(t, bulkSubNames[expected], "bulk subcommand %s should be registered", expected)
	}
}
