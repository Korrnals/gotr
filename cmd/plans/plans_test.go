package plans

import (
	"os"
	"path/filepath"
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

	// Check that plans command was added
	plansCmd, _, err := rootCmd.Find([]string{"plans"})
	assert.NoError(t, err)
	assert.NotNil(t, plansCmd)
	assert.Equal(t, "plans", plansCmd.Name())

	// Check that all subcommands exist
	subcommands := []string{"add", "get", "list", "update", "close", "delete", "entry"}
	for _, sub := range subcommands {
		subCmd, _, err := rootCmd.Find([]string{"plans", sub})
		assert.NoError(t, err, "subcommand %s should exist", sub)
		assert.NotNil(t, subCmd, "subcommand %s should not be nil", sub)
	}
}

// ==================== Tests for outputResult ====================

func TestOutputResult_JSON(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("save", "", "")

	data := &data.Plan{ID: 1, Name: "Test Plan"}
	err := outputResult(cmd, data)
	assert.NoError(t, err)
}

func TestOutputResult_ToFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "plan.json")

	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", outputFile, "")

	plan := &data.Plan{ID: 1, Name: "Test Plan"}
	err := outputResult(cmd, plan)
	assert.NoError(t, err)

	// Check file was created
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test Plan")
}

func TestOutputResult_InvalidData(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("save", "", "")

	// Channel cannot be serialized to JSON
	invalidData := make(chan int)
	err := outputResult(cmd, invalidData)
	assert.Error(t, err)
}
