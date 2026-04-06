package tests

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

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(100), runID)
			assert.Empty(t, filters)
			return []data.Test{
				{ID: 1, CaseID: 101, RunID: runID, Title: "Test 1", StatusID: 1},
				{ID: 2, CaseID: 102, RunID: runID, Title: "Test 2", StatusID: 5},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_WithStatusFilter(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(100), runID)
			assert.Equal(t, "5", filters["status_id"])
			return []data.Test{
				{ID: 2, CaseID: 102, RunID: runID, Title: "Failed Test", StatusID: 5},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100", "--status-id", "5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_DryRun_WithStatusFilter(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100", "--status-id", "5", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestListCmd_ZeroRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{{ID: 100, Name: "Run 1", ProjectID: projectID}}, nil
		},
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(100), runID)
			return []data.Test{{ID: 1, CaseID: 101, RunID: runID, Title: "Test 1", StatusID: 1}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(testCmd.Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(testCmd.Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestListCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run_id is required in non-interactive mode")
}

func TestListCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run not found")
}

func TestListCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{
				{ID: 1, CaseID: 101, RunID: runID, Title: "Test 1", StatusID: 1},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_EmptyList(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_ResolveInteractiveError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, fmt.Errorf("projects boom")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(testCmd.Context(), interactive.NewMockPrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
