package attachments

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseID_Valid tests valid ID parsing
func TestParseID_Valid(t *testing.T) {
	id, err := parseID("12345", "test_id")
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), id)
}

// TestParseID_Invalid tests invalid ID parsing
func TestParseID_Invalid(t *testing.T) {
	_, err := parseID("invalid", "test_id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid test_id")
}

// TestParseID_Zero tests zero ID
func TestParseID_Zero(t *testing.T) {
	_, err := parseID("0", "test_id")
	assert.Error(t, err)
}

// TestParseID_Negative tests negative ID
func TestParseID_Negative(t *testing.T) {
	_, err := parseID("-1", "test_id")
	assert.Error(t, err)
}

// TestValidateFileExists_Exists tests file exists check
func TestValidateFileExists_Exists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = validateFileExists(tmpFile.Name())
	assert.NoError(t, err)
}

// TestValidateFileExists_NotExists tests file not exists check
func TestValidateFileExists_NotExists(t *testing.T) {
	err := validateFileExists("/nonexistent/path/to/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

// TestNewAddCaseCmd_Creation tests command creation
func TestNewAddCaseCmd_Creation(t *testing.T) {
	// Just verify command is created correctly
	cmd := newAddCaseCmd(nil)
	assert.Equal(t, "case <case_id> <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

// TestNewAddPlanCmd_Creation tests command creation
func TestNewAddPlanCmd_Creation(t *testing.T) {
	cmd := newAddPlanCmd(nil)
	assert.Equal(t, "plan <plan_id> <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

// TestNewAddPlanEntryCmd_Creation tests command creation
func TestNewAddPlanEntryCmd_Creation(t *testing.T) {
	cmd := newAddPlanEntryCmd(nil)
	assert.Equal(t, "plan-entry <plan_id> <entry_id> <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

// TestNewAddResultCmd_Creation tests command creation
func TestNewAddResultCmd_Creation(t *testing.T) {
	cmd := newAddResultCmd(nil)
	assert.Equal(t, "result <result_id> <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

// TestNewAddRunCmd_Creation tests command creation
func TestNewAddRunCmd_Creation(t *testing.T) {
	cmd := newAddRunCmd(nil)
	assert.Equal(t, "run <run_id> <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}
