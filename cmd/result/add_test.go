package result

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Тесты для result add ====================

func TestAddCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddResultFunc: func(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
			assert.Equal(t, int64(12345), testID)
			assert.Equal(t, int64(1), req.StatusID)
			return &data.Result{
				ID:       1,
				TestID:   testID,
				StatusID: req.StatusID,
				Comment:  req.Comment,
			}, nil
		},
	}

	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--status-id", "1", "--comment", "Test passed"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_WithDefects(t *testing.T) {
	mock := &client.MockClient{
		AddResultFunc: func(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
			assert.Equal(t, int64(5), req.StatusID)
			assert.Equal(t, "BUG-123", req.Defects)
			return &data.Result{ID: 1}, nil
		},
	}

	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--status-id", "5", "--defects", "BUG-123"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--status-id", "1", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_MissingStatusID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"}) // Без --status-id

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status-id")
}

func TestAddCmd_InvalidTestID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		AddResultFunc: func(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
			return nil, fmt.Errorf("test not found")
		},
	}

	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999", "--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAddCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{{ID: 100, Name: "Run 1"}}, nil
		},
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(100), runID)
			return []data.Test{{ID: 200, CaseID: 300, Title: "Test 1"}}, nil
		},
		AddResultFunc: func(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
			assert.Equal(t, int64(200), testID)
			assert.Equal(t, int64(1), req.StatusID)
			return &data.Result{ID: 1, TestID: testID, StatusID: req.StatusID}, nil
		},
	}
	p := interactive.NewMockPrompter().WithSelectResponses(
		interactive.SelectResponse{Index: 0},
		interactive.SelectResponse{Index: 0},
		interactive.SelectResponse{Index: 0},
	)
	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--status-id", "1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_NoArgs_NonInteractive(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestAddCmd_NoArgs_NoPrompter(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test_id required in non-interactive mode")
}

func TestAddCmd_NilClient(t *testing.T) {
	nilClientFunc := func(*cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newAddCmd(nilClientFunc)
	cmd.SetArgs([]string{"12345", "--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

// ==================== Тесты для result add-case ====================

func TestAddCaseCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddResultForCaseFunc: func(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
			assert.Equal(t, int64(100), runID)
			assert.Equal(t, int64(200), caseID)
			assert.Equal(t, int64(1), req.StatusID)
			return &data.Result{ID: 1}, nil
		},
	}

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--case-id", "200", "--status-id", "1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCaseCmd_MissingCaseID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--status-id", "1"}) // Без --case-id

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestBuildAddResultRequest_AllFields(t *testing.T) {
	cmd := newAddCmd(testhelper.GetClientForTests)
	require.NoError(t, cmd.Flags().Set("status-id", "5"))
	require.NoError(t, cmd.Flags().Set("comment", "failed"))
	require.NoError(t, cmd.Flags().Set("version", "v2.1.0"))
	require.NoError(t, cmd.Flags().Set("elapsed", "3m"))
	require.NoError(t, cmd.Flags().Set("defects", "BUG-1"))
	require.NoError(t, cmd.Flags().Set("assigned-to", "42"))

	req, err := buildAddResultRequest(cmd)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), req.StatusID)
	assert.Equal(t, "failed", req.Comment)
	assert.Equal(t, "v2.1.0", req.Version)
	assert.Equal(t, "3m", req.Elapsed)
	assert.Equal(t, "BUG-1", req.Defects)
	assert.Equal(t, int64(42), req.AssignedTo)
}

func TestBuildAddResultRequest_RequiresStatusID(t *testing.T) {
	cmd := newAddCmd(testhelper.GetClientForTests)

	req, err := buildAddResultRequest(cmd)
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "--status-id is required")
}

func TestAddBulkCmd_FileReadError(t *testing.T) {
	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{"123", "--results-file", filepath.Join(t.TempDir(), "missing.json")})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file read error")
}

func TestAddBulkCmd_DryRun_FromAddTest(t *testing.T) {
	tmpDir := t.TempDir()
	resultsFile := filepath.Join(tmpDir, "results.json")
	require.NoError(t, os.WriteFile(resultsFile, []byte(`[{"test_id":1,"status_id":1}]`), 0o600))

	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{"123", "--results-file", resultsFile, "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddBulkCmd_InvalidRunID_FromAddTest(t *testing.T) {
	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{"invalid", "--results-file", filepath.Join(t.TempDir(), "any.json")})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid run ID")
}
