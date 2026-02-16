package attachments

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для parseID ====================

func TestParseID_Valid(t *testing.T) {
	id, err := parseID("12345", "test_id")
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), id)
}

func TestParseID_Invalid(t *testing.T) {
	_, err := parseID("invalid", "test_id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid test_id")
}

func TestParseID_Zero(t *testing.T) {
	_, err := parseID("0", "test_id")
	assert.Error(t, err)
}

func TestParseID_Negative(t *testing.T) {
	_, err := parseID("-1", "test_id")
	assert.Error(t, err)
}

// ==================== Тесты для validateFileExists ====================

func TestValidateFileExists_Exists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = validateFileExists(tmpFile.Name())
	assert.NoError(t, err)
}

func TestValidateFileExists_NotExists(t *testing.T) {
	err := validateFileExists("/nonexistent/path/to/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

// ==================== Тесты для newAddCaseCmd ====================

func TestNewAddCaseCmd_Creation(t *testing.T) {
	cmd := newAddCaseCmd(nil)
	assert.Equal(t, "case <case_id> <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewAddCaseCmd_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mockCalled := false
	mock := &client.MockClient{
		AddAttachmentToCaseFunc: func(caseID int64, filePath string) (*data.AttachmentResponse, error) {
			mockCalled = true
			assert.Equal(t, int64(12345), caseID)
			return &data.AttachmentResponse{
				AttachmentID: 999,
				URL:          "https://example.com/attachment/999",
				Name:         "test.txt",
				Size:         1024,
			}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "12345", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.NoError(t, err)
	assert.True(t, mockCalled, "Mock should have been called")
}

func TestNewAddCaseCmd_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "invalid", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid case_id")
}

func TestNewAddCaseCmd_ZeroCaseID(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "0", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid case_id")
}

func TestNewAddCaseCmd_FileNotFound(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "12345", "/nonexistent/file.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

func TestNewAddCaseCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "12345", "/some/file.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddCaseCmd_APIError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToCaseFunc: func(caseID int64, filePath string) (*data.AttachmentResponse, error) {
			return nil, fmt.Errorf("API error: case not found")
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "99999", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

func TestNewAddCaseCmd_WithJSONOutput(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.json")

	mock := &client.MockClient{
		AddAttachmentToCaseFunc: func(caseID int64, filePath string) (*data.AttachmentResponse, error) {
			return &data.AttachmentResponse{
				AttachmentID: 888,
				URL:          "https://example.com/attachment/888",
				Name:         "test.txt",
				Size:         2048,
			}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "12345", tmpFile.Name(), "-o", outputFile})

	err = parentCmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "888")
	assert.Contains(t, string(content), "https://example.com/attachment/888")
}

// ==================== Тесты для newAddPlanCmd ====================

func TestNewAddPlanCmd_Creation(t *testing.T) {
	cmd := newAddPlanCmd(nil)
	assert.Equal(t, "plan <plan_id> <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewAddPlanCmd_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mockCalled := false
	mock := &client.MockClient{
		AddAttachmentToPlanFunc: func(planID int64, filePath string) (*data.AttachmentResponse, error) {
			mockCalled = true
			assert.Equal(t, int64(100), planID)
			return &data.AttachmentResponse{
				AttachmentID: 200,
				URL:          "https://example.com/attachment/200",
				Name:         "plan-file.txt",
				Size:         512,
			}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan", "100", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.NoError(t, err)
	assert.True(t, mockCalled, "Mock should have been called")
}

func TestNewAddPlanCmd_InvalidPlanID(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan", "invalid", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid plan_id")
}

func TestNewAddPlanCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan", "100", "/some/file.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddPlanCmd_APIError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToPlanFunc: func(planID int64, filePath string) (*data.AttachmentResponse, error) {
			return nil, fmt.Errorf("plan not found")
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan", "999", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}

func TestNewAddPlanCmd_FileNotFound(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan", "100", "/nonexistent/file.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

// ==================== Тесты для newAddPlanEntryCmd ====================

func TestNewAddPlanEntryCmd_Creation(t *testing.T) {
	cmd := newAddPlanEntryCmd(nil)
	assert.Equal(t, "plan-entry <plan_id> <entry_id> <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewAddPlanEntryCmd_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mockCalled := false
	mock := &client.MockClient{
		AddAttachmentToPlanEntryFunc: func(planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
			mockCalled = true
			assert.Equal(t, int64(200), planID)
			assert.Equal(t, "entry-abc123", entryID)
			return &data.AttachmentResponse{
				AttachmentID: 300,
				URL:          "https://example.com/attachment/300",
				Name:         "entry-file.txt",
				Size:         256,
			}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan-entry", "200", "entry-abc123", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.NoError(t, err)
	assert.True(t, mockCalled, "Mock should have been called")
}

func TestNewAddPlanEntryCmd_InvalidPlanID(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan-entry", "invalid", "entry-abc123", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid plan_id")
}

func TestNewAddPlanEntryCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan-entry", "200", "entry-xyz", "/some/file.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddPlanEntryCmd_APIError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToPlanEntryFunc: func(planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
			return nil, fmt.Errorf("entry not found")
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan-entry", "200", "entry-xyz", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entry not found")
}

func TestNewAddPlanEntryCmd_FileNotFound(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan-entry", "200", "entry-abc123", "/nonexistent/file.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

func TestNewAddPlanEntryCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan-entry"})

	err := parentCmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты для newAddResultCmd ====================

func TestNewAddResultCmd_Creation(t *testing.T) {
	cmd := newAddResultCmd(nil)
	assert.Equal(t, "result <result_id> <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewAddResultCmd_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mockCalled := false
	mock := &client.MockClient{
		AddAttachmentToResultFunc: func(resultID int64, filePath string) (*data.AttachmentResponse, error) {
			mockCalled = true
			assert.Equal(t, int64(98765), resultID)
			return &data.AttachmentResponse{
				AttachmentID: 500,
				URL:          "https://example.com/attachment/500",
				Name:         "result-file.txt",
				Size:         128,
			}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddResultCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"result", "98765", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.NoError(t, err)
	assert.True(t, mockCalled, "Mock should have been called")
}

func TestNewAddResultCmd_InvalidResultID(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddResultCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"result", "invalid", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid result_id")
}

func TestNewAddResultCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddResultCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"result", "98765", "/some/file.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddResultCmd_APIError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToResultFunc: func(resultID int64, filePath string) (*data.AttachmentResponse, error) {
			return nil, fmt.Errorf("result not found")
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddResultCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"result", "98765", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "result not found")
}

func TestNewAddResultCmd_FileNotFound(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddResultCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"result", "98765", "/nonexistent/file.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

// ==================== Тесты для newAddRunCmd ====================

func TestNewAddRunCmd_Creation(t *testing.T) {
	cmd := newAddRunCmd(nil)
	assert.Equal(t, "run <run_id> <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewAddRunCmd_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mockCalled := false
	mock := &client.MockClient{
		AddAttachmentToRunFunc: func(runID int64, filePath string) (*data.AttachmentResponse, error) {
			mockCalled = true
			assert.Equal(t, int64(555), runID)
			return &data.AttachmentResponse{
				AttachmentID: 600,
				URL:          "https://example.com/attachment/600",
				Name:         "run-file.txt",
				Size:         1024,
			}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddRunCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"run", "555", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.NoError(t, err)
	assert.True(t, mockCalled, "Mock should have been called")
}

func TestNewAddRunCmd_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddRunCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"run", "invalid", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid run_id")
}

func TestNewAddRunCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddRunCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"run", "555", "/some/file.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddRunCmd_APIError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToRunFunc: func(runID int64, filePath string) (*data.AttachmentResponse, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddRunCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"run", "555", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run not found")
}

func TestNewAddRunCmd_FileNotFound(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddRunCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"run", "555", "/nonexistent/file.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

// ==================== Тесты для outputResult ====================

func TestOutputResult_Stdout(t *testing.T) {
	mock := &client.MockClient{
		AddAttachmentToCaseFunc: func(caseID int64, filePath string) (*data.AttachmentResponse, error) {
			return &data.AttachmentResponse{
				AttachmentID: 100,
				URL:          "https://example.com/attachment/100",
				Name:         "test.txt",
				Size:         512,
			}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	parentCmd.PersistentFlags().StringP("output", "o", "", "")

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "12345", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.NoError(t, err)
}

// TestOutputResult_StdoutOnly tests outputResult directly when output flag is not set
func TestOutputResult_StdoutOnly(t *testing.T) {
	// Create a command with output flag set to empty
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "", "")

	data := &data.AttachmentResponse{
		AttachmentID: 123,
		URL:          "https://example.com/test",
		Name:         "test.txt",
		Size:         1024,
	}

	err := outputResult(cmd, data)
	assert.NoError(t, err)
}

// ==================== Тесты для newListCmd ====================

func TestNewListCmd_Creation(t *testing.T) {
	cmd := newListCmd(nil)
	assert.Equal(t, "list", cmd.Use)
	// newListCmd has no RunE, it's a parent command for subcommands
	// Verify all subcommands are added
	assert.Equal(t, 5, len(cmd.Commands()))
}

// TestOutputResult_MarshalError tests outputResult when JSON marshaling fails
func TestOutputResult_MarshalError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "", "")

	// Channel cannot be marshaled to JSON
	invalidData := make(chan int)

	err := outputResult(cmd, invalidData)
	assert.Error(t, err)
}

// ==================== Тесты для Register ====================

func TestRegister(t *testing.T) {
	rootCmd := &cobra.Command{Use: "gotr"}

	mock := &client.MockClient{}
	getClient := func(cmd *cobra.Command) client.ClientInterface {
		return mock
	}

	Register(rootCmd, getClient)

	// Verify attachments command is registered
	attachmentsCmd, _, err := rootCmd.Find([]string{"attachments"})
	assert.NoError(t, err)
	assert.NotNil(t, attachmentsCmd)
	assert.Equal(t, "attachments", attachmentsCmd.Name())

	// Verify add command is registered
	addCmd, _, err := rootCmd.Find([]string{"attachments", "add"})
	assert.NoError(t, err)
	assert.NotNil(t, addCmd)
	assert.Equal(t, "add", addCmd.Name())

	// Verify all subcommands are registered
	commands := []string{"case", "plan", "plan-entry", "result", "run"}
	for _, cmdName := range commands {
		cmd, _, err := rootCmd.Find([]string{"attachments", "add", cmdName})
		assert.NoError(t, err, "Command %s should be registered", cmdName)
		assert.NotNil(t, cmd, "Command %s should not be nil", cmdName)
	}

	// Verify persistent flags are set
	dryRunFlag, err := addCmd.PersistentFlags().GetBool("dry-run")
	assert.NoError(t, err)
	assert.False(t, dryRunFlag)

	outputFlag, err := addCmd.PersistentFlags().GetString("output")
	assert.NoError(t, err)
	assert.Equal(t, "", outputFlag)
}
