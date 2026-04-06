// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Тесты для attachments list case ====================

func TestListCaseCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(123), caseID)
			return data.GetAttachmentsResponse{
				{ID: 1, Name: "screenshot.png", Size: 1024, CreatedOn: 1704067200},
				{ID: 2, Name: "log.txt", Size: 512, CreatedOn: 1704153600},
			}, nil
		},
	}

	cmd := newListCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "screenshot.png")
	assert.Contains(t, output, "log.txt")
	assert.Contains(t, output, "1024")
	assert.Contains(t, output, "512")
}

func TestListCaseCmd_EmptyList(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{}, nil
		},
	}

	cmd := newListCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No attachments found")
}

func TestListCaseCmd_WithSaveFlag(t *testing.T) {
	// Create temp home directory for exports
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	mock := &client.MockClient{
		GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{
				{ID: 1, Name: "test.pdf", Size: 2048, CreatedOn: 1704067200},
			}, nil
		},
	}

	cmd := newListCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--save"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify file was created in exports directory
	exportsDir := filepath.Join(tempHome, ".gotr", "exports", "attachments")
	files, err := os.ReadDir(exportsDir)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Verify content
	content, err := os.ReadFile(filepath.Join(exportsDir, files[0].Name()))
	require.NoError(t, err)
	assert.Contains(t, string(content), "test.pdf")
	assert.Contains(t, string(content), "2048")
}

func TestListCaseCmd_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case_id")
}

func TestListCaseCmd_ZeroCaseID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case_id")
}

func TestListCaseCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
			return nil, fmt.Errorf("case not found")
		},
	}

	cmd := newListCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListCaseCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCaseCmd_NoArgs_Interactive(t *testing.T) {
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
		GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(123), caseID)
			return data.GetAttachmentsResponse{{ID: 1, Name: "attachment.txt", Size: 100, CreatedOn: 1704067200}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newListCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "attachment.txt")
}

func TestListCaseCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newListCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

// ==================== Тесты для attachments list plan ====================

func TestListPlanCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForPlanFunc: func(ctx context.Context, planID int64) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(456), planID)
			return data.GetAttachmentsResponse{
				{ID: 3, Name: "plan-doc.pdf", Size: 4096, CreatedOn: 1704067200},
			}, nil
		},
	}

	cmd := newListPlanCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"456"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "plan-doc.pdf")
}

func TestListPlanCmd_InvalidPlanID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListPlanCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_id")
}

func TestListPlanCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForPlanFunc: func(ctx context.Context, planID int64) (data.GetAttachmentsResponse, error) {
			return nil, fmt.Errorf("plan not found")
		},
	}

	cmd := newListPlanCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListPlanCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetPlansResponse{{ID: 456, Name: "Plan 456"}}, nil
		},
		GetAttachmentsForPlanFunc: func(ctx context.Context, planID int64) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(456), planID)
			return data.GetAttachmentsResponse{{ID: 10, Name: "plan-attachment.txt", Size: 10, CreatedOn: 1704067200}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newListPlanCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "plan-attachment.txt")
}

func TestListPlanCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newListPlanCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

// ==================== Тесты для attachments list plan-entry ====================

func TestListPlanEntryCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForPlanEntryFunc: func(ctx context.Context, planID int64, entryID string) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(100), planID)
			assert.Equal(t, "entry-abc123", entryID)
			return data.GetAttachmentsResponse{
				{ID: 4, Name: "entry-data.csv", Size: 2048, CreatedOn: 1704067200},
			}, nil
		},
	}

	cmd := newListPlanEntryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "entry-abc123"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "entry-data.csv")
}

func TestListPlanEntryCmd_InvalidPlanID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListPlanEntryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "entry-abc123"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_id")
}

func TestListPlanEntryCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForPlanEntryFunc: func(ctx context.Context, planID int64, entryID string) (data.GetAttachmentsResponse, error) {
			return nil, fmt.Errorf("plan entry not found")
		},
	}

	cmd := newListPlanEntryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999", "entry-xyz"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListPlanEntryCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListPlanEntryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListPlanEntryCmd_OnlyOneArg(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListPlanEntryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListPlanEntryCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetPlansResponse{{ID: 100, Name: "Plan 100"}}, nil
		},
		GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
			assert.Equal(t, int64(100), planID)
			return &data.Plan{ID: 100, Name: "Plan 100", Entries: []data.PlanEntry{{ID: "entry-1", Name: "Entry 1"}}}, nil
		},
		GetAttachmentsForPlanEntryFunc: func(ctx context.Context, planID int64, entryID string) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(100), planID)
			assert.Equal(t, "entry-1", entryID)
			return data.GetAttachmentsResponse{{ID: 11, Name: "entry-attachment.txt", Size: 10, CreatedOn: 1704067200}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newListPlanEntryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "entry-attachment.txt")
}

func TestListPlanEntryCmd_OnlyPlanID_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
			assert.Equal(t, int64(100), planID)
			return &data.Plan{ID: 100, Name: "Plan 100", Entries: []data.PlanEntry{{ID: "entry-2", Name: "Entry 2"}}}, nil
		},
		GetAttachmentsForPlanEntryFunc: func(ctx context.Context, planID int64, entryID string) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(100), planID)
			assert.Equal(t, "entry-2", entryID)
			return data.GetAttachmentsResponse{{ID: 12, Name: "entry-two.txt", Size: 10, CreatedOn: 1704067200}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newListPlanEntryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"100"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "entry-two.txt")
}

func TestListPlanEntryCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newListPlanEntryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

// ==================== Тесты для attachments list run ====================

func TestListRunCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForRunFunc: func(ctx context.Context, runID int64) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(789), runID)
			return data.GetAttachmentsResponse{
				{ID: 5, Name: "run-report.html", Size: 8192, CreatedOn: 1704067200},
			}, nil
		},
	}

	cmd := newListRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "run-report.html")
}

func TestListRunCmd_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run_id")
}

func TestListRunCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForRunFunc: func(ctx context.Context, runID int64) (data.GetAttachmentsResponse, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	cmd := newListRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListRunCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{{ID: 789, Name: "Run 789"}}, nil
		},
		GetAttachmentsForRunFunc: func(ctx context.Context, runID int64) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(789), runID)
			return data.GetAttachmentsResponse{{ID: 13, Name: "run-attachment.txt", Size: 10, CreatedOn: 1704067200}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newListRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "run-attachment.txt")
}

func TestListRunCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newListRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

// ==================== Тесты для attachments list test ====================

func TestListTestCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForTestFunc: func(ctx context.Context, testID int64) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(321), testID)
			return data.GetAttachmentsResponse{
				{ID: 6, Name: "test-screenshot.png", Size: 3072, CreatedOn: 1704067200},
			}, nil
		},
	}

	cmd := newListTestCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"321"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "test-screenshot.png")
}

func TestListTestCmd_InvalidTestID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListTestCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test_id")
}

func TestListTestCmd_ZeroTestID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListTestCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test_id")
}

func TestListTestCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForTestFunc: func(ctx context.Context, testID int64) (data.GetAttachmentsResponse, error) {
			return nil, fmt.Errorf("test not found")
		},
	}

	cmd := newListTestCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListTestCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{{ID: 200, Name: "Run 200"}}, nil
		},
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(200), runID)
			return []data.Test{{ID: 321, CaseID: 55, Title: "Test 321"}}, nil
		},
		GetAttachmentsForTestFunc: func(ctx context.Context, testID int64) (data.GetAttachmentsResponse, error) {
			assert.Equal(t, int64(321), testID)
			return data.GetAttachmentsResponse{{ID: 14, Name: "test-attachment.txt", Size: 10, CreatedOn: 1704067200}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newListTestCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "test-attachment.txt")
}

func TestListTestCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newListTestCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

// ==================== Тесты для outputAttachmentsList ====================

func TestOutputAttachmentsList_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{}, nil
		},
	}

	cmd := newListCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No attachments found")
}

// ============= LAYER 2: !HasPrompterInContext branches + invalid ID in case 1 =============

func TestListPlanCmd_NoArgs_NoPrompter_Error(t *testing.T) {
cmd := newListPlanCmd(testhelper.GetClientForTests)
cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
cmd.SetArgs([]string{})

err := cmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "plan_id required")
}

func TestListRunCmd_NoArgs_NoPrompter_Error(t *testing.T) {
cmd := newListRunCmd(testhelper.GetClientForTests)
cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
cmd.SetArgs([]string{})

err := cmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "run_id required")
}

func TestListTestCmd_NoArgs_NoPrompter_Error(t *testing.T) {
cmd := newListTestCmd(testhelper.GetClientForTests)
cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
cmd.SetArgs([]string{})

err := cmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "test_id required")
}

func TestListPlanEntryCmd_OnlyPlanID_InvalidPlanID(t *testing.T) {
cmd := newListPlanEntryCmd(testhelper.GetClientForTests)
cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
cmd.SetArgs([]string{"invalid"})

err := cmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "plan_id")
}

// ============= LAYER 2 EXTENSION: newListPlanEntryCmd remaining branches =============

func TestListPlanEntryCmd_OnlyPlanID_ResolvePlanEntryError(t *testing.T) {
// case 1 with valid planID, interactive, but resolvePlanEntryIDInteractive fails
mock := &client.MockClient{
GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
return nil, assert.AnError
},
}

p := interactive.NewMockPrompter()
cmd := newListPlanEntryCmd(testhelper.GetClientForTests)
cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
cmd.SetArgs([]string{"100"})

err := cmd.Execute()
assert.Error(t, err)
}
