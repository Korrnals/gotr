package milestones

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Tests for Register ====================

func TestRegister(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}

	// Mock function for getting client
	mockFn := func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	}

	Register(rootCmd, mockFn)

	// Check that milestones command was added
	milestonesCmd, _, err := rootCmd.Find([]string{"milestones"})
	assert.NoError(t, err)
	assert.NotNil(t, milestonesCmd)
	assert.Equal(t, "milestones", milestonesCmd.Name())

	// Check that all subcommands exist
	subcommands := []string{"add", "get", "list", "update", "delete"}
	for _, sub := range subcommands {
		subCmd, _, err := rootCmd.Find([]string{"milestones", sub})
		assert.NoError(t, err, "subcommand %s should exist", sub)
		assert.NotNil(t, subCmd, "subcommand %s should not be nil", sub)
	}
}

// ==================== Tests for outputResult ====================

func TestOutputResult_Default(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")

	data := &data.Milestone{ID: 1, Name: "Test Milestone"}
	err := outputResult(cmd, data)
	assert.NoError(t, err)
}

func TestOutputResult_JSON(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "json", "")

	data := &data.Milestone{ID: 1, Name: "Test Milestone"}
	err := outputResult(cmd, data)
	assert.NoError(t, err)
}

// ==================== Tests for outputList ====================

func TestOutputList(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")

	data := []data.Milestone{
		{ID: 1, Name: "Milestone 1"},
		{ID: 2, Name: "Milestone 2"},
	}
	err := outputList(cmd, data)
	assert.NoError(t, err)
}
