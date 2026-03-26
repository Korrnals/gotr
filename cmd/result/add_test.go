package result

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
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
