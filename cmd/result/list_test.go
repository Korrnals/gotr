package result

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Tests for result list (direct mode) ====================

func TestListCmd_Direct_Success(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForRunFunc: func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), runID)
			return data.GetResultsResponse{
				{ID: 1, TestID: 100, StatusID: 1, Comment: "Passed"},
				{ID: 2, TestID: 101, StatusID: 5, Comment: "Failed"},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_Direct_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run")
}

func TestListCmd_Direct_ZeroRunID(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForRunFunc: func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("invalid run_id")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_Direct_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForRunFunc: func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListCmd_Direct_EmptyResults(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForRunFunc: func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for result list (interactive mode) ====================

func TestListCmd_Interactive_NoProjects(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter()

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	// No arguments passed - should enable interactive mode
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no projects")
}

// ==================== Tests for edge cases ====================

func TestListCmd_NilClient(t *testing.T) {
	cmd := newListCmd(func(cmd *cobra.Command) client.ClientInterface { return nil })
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client")
}

func TestListCmd_NegativeRunID(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForRunFunc: func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("invalid run_id")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"-1"})

	err := cmd.Execute()
	assert.Error(t, err)
}
