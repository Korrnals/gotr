package run

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

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetRunFunc: func(ctx context.Context, runID int64) (*data.Run, error) {
			assert.Equal(t, int64(12345), runID)
			return &data.Run{
				ID:          runID,
				Name:        "Test Run",
				Description: "Test description",
				ProjectID:   30,
				SuiteID:     100,
				IsCompleted: false,
				PassedCount: 10,
				FailedCount: 2,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_ZeroRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NegativeRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"-1"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetRunFunc: func(ctx context.Context, runID int64) (*data.Run, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetRunsResponse{{ID: 12345, Name: "Run 12345"}}, nil
		},
		GetRunFunc: func(ctx context.Context, runID int64) (*data.Run, error) {
			assert.Equal(t, int64(12345), runID)
			return &data.Run{ID: runID, Name: "Run 12345"}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetCmd_CompletedRun(t *testing.T) {
	mock := &client.MockClient{
		GetRunFunc: func(ctx context.Context, runID int64) (*data.Run, error) {
			return &data.Run{
				ID:           runID,
				Name:         "Completed Run",
				IsCompleted:  true,
				PassedCount:  50,
				FailedCount:  0,
				BlockedCount: 0,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"54321"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NilClient(t *testing.T) {
	// Test when getClient returns nil
	cmd := newGetCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}
