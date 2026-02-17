package labels

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Tests for parseLabels ====================

// TestParseLabels_Single tests parsing single label
func TestParseLabels_Single(t *testing.T) {
	labels := parseLabels("smoke")
	assert.Equal(t, []string{"smoke"}, labels)
}

// TestParseLabels_Multiple tests parsing multiple labels
func TestParseLabels_Multiple(t *testing.T) {
	labels := parseLabels("smoke,critical,regression")
	assert.Equal(t, []string{"smoke", "critical", "regression"}, labels)
}

// TestParseLabels_WithSpaces tests parsing labels with spaces
func TestParseLabels_WithSpaces(t *testing.T) {
	labels := parseLabels("smoke, critical, regression")
	assert.Equal(t, []string{"smoke", "critical", "regression"}, labels)
}

// TestParseLabels_Empty tests parsing empty string
func TestParseLabels_Empty(t *testing.T) {
	labels := parseLabels("")
	assert.Empty(t, labels)
}

// TestParseLabels_WithEmptyParts tests parsing labels with empty parts
func TestParseLabels_WithEmptyParts(t *testing.T) {
	labels := parseLabels("smoke,,critical,")
	assert.Equal(t, []string{"smoke", "critical"}, labels)
}

// ==================== Tests for parseIntList ====================

// TestParseIntList_Single tests parsing single ID
func TestParseIntList_Single(t *testing.T) {
	ids := parseIntList("123")
	assert.Equal(t, []int64{123}, ids)
}

// TestParseIntList_Multiple tests parsing multiple IDs
func TestParseIntList_Multiple(t *testing.T) {
	ids := parseIntList("1,2,3,4,5")
	assert.Equal(t, []int64{1, 2, 3, 4, 5}, ids)
}

// TestParseIntList_WithSpaces tests parsing IDs with spaces
func TestParseIntList_WithSpaces(t *testing.T) {
	ids := parseIntList("1, 2, 3")
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

// TestParseIntList_Invalid tests parsing with invalid values
func TestParseIntList_Invalid(t *testing.T) {
	ids := parseIntList("1,invalid,3")
	assert.Equal(t, []int64{1, 3}, ids)
}

// TestParseIntList_ZeroAndNegative tests filtering zero and negative
func TestParseIntList_ZeroAndNegative(t *testing.T) {
	ids := parseIntList("1,0,-1,2")
	assert.Equal(t, []int64{1, 2}, ids)
}

// TestParseIntList_Empty tests parsing empty string
func TestParseIntList_Empty(t *testing.T) {
	ids := parseIntList("")
	assert.Empty(t, ids)
}

// TestParseIntList_WithEmptyParts tests parsing IDs with empty parts
func TestParseIntList_WithEmptyParts(t *testing.T) {
	ids := parseIntList("1,,2,")
	assert.Equal(t, []int64{1, 2}, ids)
}

// ==================== Tests for newUpdateTestCmd ====================

// TestNewUpdateTestCmd_Creation tests command creation
func TestNewUpdateTestCmd_Creation(t *testing.T) {
	cmd := newUpdateTestCmd(nil)
	assert.Equal(t, "test <test_id>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	// Check required flag
	labelsFlag := cmd.Flags().Lookup("labels")
	assert.NotNil(t, labelsFlag)
}

// TestNewUpdateTestCmd_Success tests successful label update
func TestNewUpdateTestCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateTestLabelsFunc: func(testID int64, labels []string) error {
			assert.Equal(t, int64(123), testID)
			assert.Equal(t, []string{"smoke", "critical"}, labels)
			return nil
		},
	}

	cmd := newUpdateTestCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--labels", "smoke,critical"})

	// Set output to capture the success message
	var stdoutBuf bytes.Buffer
	cmd.SetOut(&stdoutBuf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdoutBuf.String(), "Labels updated")
}

// TestNewUpdateTestCmd_DryRun tests dry-run mode
func TestNewUpdateTestCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	// Add dry-run flag locally since it's a persistent flag on parent command
	cmd.Flags().Bool("dry-run", false, "")
	cmd.SetArgs([]string{"123", "--labels", "smoke", "--dry-run"})

	// Dry-run output goes to os.Stderr directly, so we need to capture it
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := cmd.Execute()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "DRY RUN")
}

// TestNewUpdateTestCmd_InvalidTestID tests invalid test_id
func TestNewUpdateTestCmd_InvalidTestID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid test_id")
}

// TestNewUpdateTestCmd_ZeroTestID tests zero test_id
func TestNewUpdateTestCmd_ZeroTestID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid test_id")
}

// TestNewUpdateTestCmd_NegativeTestID tests negative test_id
func TestNewUpdateTestCmd_NegativeTestID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	// Set flags first, then positional args with -- separator
	cmd.SetArgs([]string{"--labels", "smoke", "--", "-1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid test_id")
}

// TestNewUpdateTestCmd_MissingLabelsFlag tests missing labels flag (handled by cobra)
func TestNewUpdateTestCmd_MissingLabelsFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	assert.Error(t, err)
	// Cobra returns "required flag(s)" error
	assert.Contains(t, err.Error(), "required flag")
}

// TestNewUpdateTestCmd_EmptyLabelsFlag tests empty labels flag
func TestNewUpdateTestCmd_EmptyLabelsFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--labels", ""})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--labels flag is required")
}

// TestNewUpdateTestCmd_APIError tests API error handling
func TestNewUpdateTestCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		UpdateTestLabelsFunc: func(testID int64, labels []string) error {
			return fmt.Errorf("connection refused")
		},
	}

	cmd := newUpdateTestCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update labels")
	assert.Contains(t, err.Error(), "connection refused")
}

// TestNewUpdateTestCmd_NoArgs tests no arguments provided
func TestNewUpdateTestCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

// TestNewUpdateTestCmd_TooManyArgs tests too many arguments
func TestNewUpdateTestCmd_TooManyArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "456", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Tests for newUpdateTestsCmd ====================

// TestNewUpdateTestsCmd_Creation tests command creation
func TestNewUpdateTestsCmd_Creation(t *testing.T) {
	cmd := newUpdateTestsCmd(nil)
	assert.Equal(t, "tests", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	// Check required flags
	assert.NotNil(t, cmd.Flags().Lookup("run-id"))
	assert.NotNil(t, cmd.Flags().Lookup("test-ids"))
	assert.NotNil(t, cmd.Flags().Lookup("labels"))
}

// TestNewUpdateTestsCmd_Success tests successful labels update
func TestNewUpdateTestsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateTestsLabelsFunc: func(runID int64, testIDs []int64, labels []string) error {
			assert.Equal(t, int64(100), runID)
			assert.Equal(t, []int64{1, 2, 3}, testIDs)
			assert.Equal(t, []string{"smoke", "critical"}, labels)
			return nil
		},
	}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "100", "--test-ids", "1,2,3", "--labels", "smoke,critical"})

	// Set output to capture the success message
	var stdoutBuf bytes.Buffer
	cmd.SetOut(&stdoutBuf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdoutBuf.String(), "Labels updated")
}

// TestNewUpdateTestsCmd_DryRun tests dry-run mode
func TestNewUpdateTestsCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	// Add dry-run flag locally since it's a persistent flag on parent command
	cmd.Flags().Bool("dry-run", false, "")
	cmd.SetArgs([]string{"--run-id", "100", "--test-ids", "1,2,3", "--labels", "smoke", "--dry-run"})

	// Dry-run output goes to os.Stderr directly, so we need to capture it
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := cmd.Execute()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "DRY RUN")
}

// TestNewUpdateTestsCmd_MissingRunID tests missing run-id (handled by cobra)
func TestNewUpdateTestsCmd_MissingRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--test-ids", "1,2,3", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

// TestNewUpdateTestsCmd_ZeroRunID tests zero run-id
func TestNewUpdateTestsCmd_ZeroRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "0", "--test-ids", "1,2,3", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--run-id is required and must be positive")
}

// TestNewUpdateTestsCmd_NegativeRunID tests negative run-id
func TestNewUpdateTestsCmd_NegativeRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "-1", "--test-ids", "1,2,3", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--run-id is required and must be positive")
}

// TestNewUpdateTestsCmd_MissingTestIDs tests missing test-ids (handled by cobra)
func TestNewUpdateTestsCmd_MissingTestIDs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "100", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

// TestNewUpdateTestsCmd_EmptyTestIDs tests empty test-ids
func TestNewUpdateTestsCmd_EmptyTestIDs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "100", "--test-ids", "", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--test-ids is required")
}

// TestNewUpdateTestsCmd_InvalidTestIDs tests invalid test-ids format
func TestNewUpdateTestsCmd_InvalidTestIDs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "100", "--test-ids", "invalid,also-invalid", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid test-ids")
}

// TestNewUpdateTestsCmd_ZeroAndNegativeTestIDs tests filtering zero and negative test-ids
func TestNewUpdateTestsCmd_ZeroAndNegativeTestIDs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "100", "--test-ids", "0,-1,-5", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid test-ids")
}

// TestNewUpdateTestsCmd_MissingLabelsFlag tests missing labels flag (handled by cobra)
func TestNewUpdateTestsCmd_MissingLabelsFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "100", "--test-ids", "1,2,3"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

// TestNewUpdateTestsCmd_EmptyLabelsFlag tests empty labels flag
func TestNewUpdateTestsCmd_EmptyLabelsFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "100", "--test-ids", "1,2,3", "--labels", ""})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--labels flag is required")
}

// TestNewUpdateTestsCmd_APIError tests API error handling
func TestNewUpdateTestsCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		UpdateTestsLabelsFunc: func(runID int64, testIDs []int64, labels []string) error {
			return fmt.Errorf("permission denied")
		},
	}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "100", "--test-ids", "1,2,3", "--labels", "smoke"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update labels")
	assert.Contains(t, err.Error(), "permission denied")
}

// TestNewUpdateTestsCmd_SingleTestID tests with single test ID
func TestNewUpdateTestsCmd_SingleTestID(t *testing.T) {
	mock := &client.MockClient{
		UpdateTestsLabelsFunc: func(runID int64, testIDs []int64, labels []string) error {
			assert.Equal(t, []int64{42}, testIDs)
			return nil
		},
	}

	cmd := newUpdateTestsCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--run-id", "100", "--test-ids", "42", "--labels", "smoke"})

	var stdoutBuf bytes.Buffer
	cmd.SetOut(&stdoutBuf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for Register ====================

// TestRegister tests that Register function doesn't panic and registers all commands
func TestRegister(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	getClient := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	// Should not panic
	assert.NotPanics(t, func() {
		Register(root, getClient)
	})

	// Verify labels command was added
	labelsCmd, _, err := root.Find([]string{"labels"})
	assert.NoError(t, err)
	assert.NotNil(t, labelsCmd)
	assert.Equal(t, "labels", labelsCmd.Use)

	// Verify get subcommand exists
	getCmd, _, err := root.Find([]string{"labels", "get"})
	assert.NoError(t, err)
	assert.NotNil(t, getCmd)

	// Verify list subcommand exists
	listCmd, _, err := root.Find([]string{"labels", "list"})
	assert.NoError(t, err)
	assert.NotNil(t, listCmd)

	// Verify update-label subcommand exists
	updateLabelCmd, _, err := root.Find([]string{"labels", "update-label"})
	assert.NoError(t, err)
	assert.NotNil(t, updateLabelCmd)

	// Verify update subcommand exists
	updateCmd, _, err := root.Find([]string{"labels", "update"})
	assert.NoError(t, err)
	assert.NotNil(t, updateCmd)

	// Verify update test subcommand exists
	updateTestCmd, _, err := root.Find([]string{"labels", "update", "test"})
	assert.NoError(t, err)
	assert.NotNil(t, updateTestCmd)

	// Verify update tests subcommand exists
	updateTestsCmd, _, err := root.Find([]string{"labels", "update", "tests"})
	assert.NoError(t, err)
	assert.NotNil(t, updateTestsCmd)

	// Verify dry-run persistent flag exists on update command
	dryRunFlag := updateCmd.PersistentFlags().Lookup("dry-run")
	assert.NotNil(t, dryRunFlag)
}
