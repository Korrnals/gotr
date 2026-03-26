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

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			assert.Equal(t, int64(12345), testID)
			return &data.Test{
				ID:       testID,
				CaseID:   100,
				RunID:    200,
				Title:    "Test Case Title",
				StatusID: 1,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestGetCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NegativeID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"-1"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NoArgs_Interactive(t *testing.T) {
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
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			assert.Equal(t, int64(1), testID)
			return &data.Test{ID: 1, CaseID: 101, RunID: 100, Title: "Test 1", StatusID: 1}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(testCmd.Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(testCmd.Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			return nil, fmt.Errorf("test not found")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test not found")
}

func TestGetCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			return &data.Test{
				ID:       testID,
				CaseID:   100,
				RunID:    200,
				Title:    "Test Case Title",
				StatusID: 1,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
