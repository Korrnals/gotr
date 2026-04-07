package attachments

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Тесты для flags.ValidateRequiredID ====================

func TestValidateRequiredID_Valid(t *testing.T) {
	id, err := flags.ValidateRequiredID([]string{"12345"}, 0, "test_id")
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), id)
}

func TestValidateRequiredID_Invalid(t *testing.T) {
	_, err := flags.ValidateRequiredID([]string{"invalid"}, 0, "test_id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test_id")
}

func TestValidateRequiredID_Zero(t *testing.T) {
	_, err := flags.ValidateRequiredID([]string{"0"}, 0, "test_id")
	assert.Error(t, err)
}

func TestValidateRequiredID_Negative(t *testing.T) {
	_, err := flags.ValidateRequiredID([]string{"-1"}, 0, "test_id")
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
	assert.Equal(t, "case [case_id] <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewAddCaseCmd_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mockCalled := false
	mock := &client.MockClient{
		AddAttachmentToCaseFunc: func(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
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
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "invalid", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case_id")
}

func TestNewAddCaseCmd_ZeroCaseID(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "0", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case_id")
}

func TestNewAddCaseCmd_FileNotFound(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "12345", "/some/file.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddCaseCmd_NoID_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetSuitesResponse{{ID: 10, Name: "Suite 10", ProjectID: 1}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, int64(10), suiteID)
			return data.GetCasesResponse{{ID: 123, Title: "Case 123"}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	parentCmd.SetArgs([]string{"case", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddCaseCmd_NoID_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	parentCmd.SetArgs([]string{"case", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestNewAddCaseCmd_APIError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToCaseFunc: func(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
			return nil, fmt.Errorf("API error: case not found")
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "99999", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

func TestNewAddCaseCmd_WithSaveFlag(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToCaseFunc: func(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
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
	output.AddFlag(parentCmd)

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "12345", tmpFile.Name(), "--save"})

	err = parentCmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для newAddPlanCmd ====================

func TestNewAddPlanCmd_Creation(t *testing.T) {
	cmd := newAddPlanCmd(nil)
	assert.Equal(t, "plan [plan_id] <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewAddPlanCmd_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mockCalled := false
	mock := &client.MockClient{
		AddAttachmentToPlanFunc: func(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error) {
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
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

	cmd := newAddPlanCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan", "invalid", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_id")
}

func TestNewAddPlanCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddPlanCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan", "100", "/some/file.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddPlanCmd_NoID_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetPlansResponse{{ID: 100, Name: "Plan 100"}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddPlanCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	parentCmd.SetArgs([]string{"plan", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddPlanCmd_NoID_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddPlanCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	parentCmd.SetArgs([]string{"plan", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestNewAddPlanCmd_APIError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToPlanFunc: func(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error) {
			return nil, fmt.Errorf("plan not found")
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

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
	assert.Equal(t, "plan-entry [plan_id] [entry_id] <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewAddPlanEntryCmd_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mockCalled := false
	mock := &client.MockClient{
		AddAttachmentToPlanEntryFunc: func(ctx context.Context, planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
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
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan-entry", "invalid", "entry-abc123", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_id")
}

func TestNewAddPlanEntryCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

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
		AddAttachmentToPlanEntryFunc: func(ctx context.Context, planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
			return nil, fmt.Errorf("entry not found")
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"plan-entry"})

	err := parentCmd.Execute()
	assert.Error(t, err)
}

func TestNewAddPlanEntryCmd_NoIDs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetPlansResponse{{ID: 200, Name: "Plan 200"}}, nil
		},
		GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
			assert.Equal(t, int64(200), planID)
			return &data.Plan{ID: 200, Name: "Plan 200", Entries: []data.PlanEntry{{ID: "entry-xyz", Name: "Entry XYZ"}}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	parentCmd.SetArgs([]string{"plan-entry", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddPlanEntryCmd_OnlyPlanID_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
			assert.Equal(t, int64(200), planID)
			return &data.Plan{ID: 200, Name: "Plan 200", Entries: []data.PlanEntry{{ID: "entry-abc123", Name: "Entry ABC"}}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	parentCmd.SetArgs([]string{"plan-entry", "200", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddPlanEntryCmd_NoIDs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	parentCmd.SetArgs([]string{"plan-entry", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

// ==================== Тесты для newAddResultCmd ====================

func TestNewAddResultCmd_Creation(t *testing.T) {
	cmd := newAddResultCmd(nil)
	assert.Equal(t, "result [result_id] <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewAddResultCmd_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mockCalled := false
	mock := &client.MockClient{
		AddAttachmentToResultFunc: func(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error) {
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
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

	cmd := newAddResultCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"result", "invalid", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "result_id")
}

func TestNewAddResultCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

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
		AddAttachmentToResultFunc: func(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error) {
			return nil, fmt.Errorf("result not found")
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

	cmd := newAddResultCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"result", "98765", "/nonexistent/file.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

func TestNewAddResultCmd_NoID_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{{ID: 555, Name: "Run 555"}}, nil
		},
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(555), runID)
			return []data.Test{{ID: 444, CaseID: 12, Title: "Test 444"}}, nil
		},
		GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(444), testID)
			return data.GetResultsResponse{{ID: 98765, TestID: 444, StatusID: 1}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddResultCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	parentCmd.SetArgs([]string{"result", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddResultCmd_NoID_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddResultCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	parentCmd.SetArgs([]string{"result", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

// ==================== Тесты для newAddRunCmd ====================

func TestNewAddRunCmd_Creation(t *testing.T) {
	cmd := newAddRunCmd(nil)
	assert.Equal(t, "run [run_id] <file_path>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewAddRunCmd_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mockCalled := false
	mock := &client.MockClient{
		AddAttachmentToRunFunc: func(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error) {
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
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

	cmd := newAddRunCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"run", "invalid", "/tmp/test.txt"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run_id")
}

func TestNewAddRunCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddRunCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"run", "555", "/some/file.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddRunCmd_NoID_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{{ID: 555, Name: "Run 555"}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddRunCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	parentCmd.SetArgs([]string{"run", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddRunCmd_NoID_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

	cmd := newAddRunCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	parentCmd.SetArgs([]string{"run", "./dummy.txt", "--dry-run"})

	err := parentCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestNewAddRunCmd_APIError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToRunFunc: func(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	parentCmd := &cobra.Command{Use: "add"}
	parentCmd.PersistentFlags().Bool("dry-run", false, "")
	output.AddFlag(parentCmd)

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
	output.AddFlag(parentCmd)

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
		AddAttachmentToCaseFunc: func(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
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
	output.AddFlag(parentCmd)

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	parentCmd.AddCommand(cmd)

	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	parentCmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	parentCmd.SetArgs([]string{"case", "12345", tmpFile.Name()})

	err = parentCmd.Execute()
	assert.NoError(t, err)
}

// TestOutputResult_StdoutOnly tests outputResult directly when save flag is not set
func TestOutputResult_StdoutOnly(t *testing.T) {
	// Create a command with save flag not set
	cmd := &cobra.Command{Use: "test"}
	output.AddFlag(cmd)

	data := &data.AttachmentResponse{
		AttachmentID: 123,
		URL:          "https://example.com/test",
		Name:         "test.txt",
		Size:         1024,
	}

	err := output.OutputResult(cmd, data, "attachments")
	assert.NoError(t, err)
}

// ==================== Тесты для newListCmd ======================================

func TestNewListCmd_Creation(t *testing.T) {
	cmd := newListCmd(nil)
	assert.Equal(t, "list", cmd.Use)
	// newListCmd has no RunE, it's a parent command for subcommands
	// Verify all subcommands are added
	assert.Equal(t, 6, len(cmd.Commands()))
}

// TestOutputResult_MarshalError tests outputResult when JSON marshaling fails
func TestOutputResult_MarshalError(t *testing.T) {
	// Create temp home directory for exports
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	cmd := &cobra.Command{Use: "test"}
	output.AddFlag(cmd)
	cmd.SetArgs([]string{"--save"})
	require.NoError(t, cmd.Execute())

	// Channel cannot be marshaled to JSON
	invalidData := make(chan int)

	err := output.OutputResult(cmd, invalidData, "attachments")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "marshaling")
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

	// Verify save flag is registered on subcommand (bool type)
	caseCmd, _, err := rootCmd.Find([]string{"attachments", "add", "case"})
	assert.NoError(t, err)
	saveFlag, err := caseCmd.Flags().GetBool("save")
	assert.NoError(t, err)
	assert.False(t, saveFlag)
}

// ============= LAYER 2: !HasPrompterInContext branches =============

func TestNewAddCaseCmd_NoID_NoPrompter_Error(t *testing.T) {
parentCmd := &cobra.Command{Use: "add"}
parentCmd.PersistentFlags().Bool("dry-run", false, "")
output.AddFlag(parentCmd)

cmd := newAddCaseCmd(testhelper.GetClientForTests)
parentCmd.AddCommand(cmd)

parentCmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
parentCmd.SetArgs([]string{"case", "./dummy.txt"})

err := parentCmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "case_id required")
}

func TestNewAddPlanCmd_NoID_NoPrompter_Error(t *testing.T) {
parentCmd := &cobra.Command{Use: "add"}
parentCmd.PersistentFlags().Bool("dry-run", false, "")
output.AddFlag(parentCmd)

cmd := newAddPlanCmd(testhelper.GetClientForTests)
parentCmd.AddCommand(cmd)

parentCmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
parentCmd.SetArgs([]string{"plan", "./dummy.txt"})

err := parentCmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "plan_id required")
}

func TestNewAddPlanEntryCmd_OnlyPlanID_NoPrompter_Error(t *testing.T) {
parentCmd := &cobra.Command{Use: "add"}
parentCmd.PersistentFlags().Bool("dry-run", false, "")
output.AddFlag(parentCmd)

cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
parentCmd.AddCommand(cmd)

parentCmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
parentCmd.SetArgs([]string{"plan-entry", "100", "./dummy.txt"})

err := parentCmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "entry_id required")
}

func TestNewAddResultCmd_NoID_NoPrompter_Error(t *testing.T) {
parentCmd := &cobra.Command{Use: "add"}
parentCmd.PersistentFlags().Bool("dry-run", false, "")
output.AddFlag(parentCmd)

cmd := newAddResultCmd(testhelper.GetClientForTests)
parentCmd.AddCommand(cmd)

parentCmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
parentCmd.SetArgs([]string{"result", "./dummy.txt"})

err := parentCmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "result_id required")
}

func TestNewAddRunCmd_NoID_NoPrompter_Error(t *testing.T) {
parentCmd := &cobra.Command{Use: "add"}
parentCmd.PersistentFlags().Bool("dry-run", false, "")
output.AddFlag(parentCmd)

cmd := newAddRunCmd(testhelper.GetClientForTests)
parentCmd.AddCommand(cmd)

parentCmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
parentCmd.SetArgs([]string{"run", "./dummy.txt"})

err := parentCmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "run_id required")
}

// ============= LAYER 2 EXTENSION: newAddPlanEntryCmd remaining branches =============

func TestNewAddPlanEntryCmd_OnlyFilePath_NoPrompter_Error(t *testing.T) {
// default case (1 arg) with no prompter → plan_id required
parentCmd := &cobra.Command{Use: "add"}
parentCmd.PersistentFlags().Bool("dry-run", false, "")
output.AddFlag(parentCmd)

cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
parentCmd.AddCommand(cmd)

parentCmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
parentCmd.SetArgs([]string{"plan-entry", "./dummy.txt"})

err := parentCmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "plan_id required")
}

func TestNewAddPlanEntryCmd_InvalidPlanIDTwoArgs(t *testing.T) {
// case 2 with invalid planID → ValidateRequiredID error
parentCmd := &cobra.Command{Use: "add"}
parentCmd.PersistentFlags().Bool("dry-run", false, "")
output.AddFlag(parentCmd)

cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
parentCmd.AddCommand(cmd)

parentCmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
parentCmd.SetArgs([]string{"plan-entry", "invalid", "./dummy.txt"})

err := parentCmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "plan_id")
}

func TestNewAddPlanEntryCmd_ResolvePlanEntryError(t *testing.T) {
// case 2 with valid planID, interactive, but resolvePlanEntryIDInteractive fails
mock := &client.MockClient{
GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
return nil, assert.AnError
},
}

p := interactive.NewMockPrompter()

parentCmd := &cobra.Command{Use: "add"}
parentCmd.PersistentFlags().Bool("dry-run", false, "")
output.AddFlag(parentCmd)

cmd := newAddPlanEntryCmd(testhelper.GetClientForTests)
parentCmd.AddCommand(cmd)

parentCmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
parentCmd.SetArgs([]string{"plan-entry", "200", "./dummy.txt"})

err := parentCmd.Execute()
assert.Error(t, err)
}
