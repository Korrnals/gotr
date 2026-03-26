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

func TestCloseCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		CloseRunFunc: func(ctx context.Context, runID int64) (*data.Run, error) {
			assert.Equal(t, int64(12345), runID)
			return &data.Run{ID: runID, IsCompleted: true}, nil
		},
	}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCloseCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCloseCmd_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCloseCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		CloseRunFunc: func(ctx context.Context, runID int64) (*data.Run, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCloseCmd_NilClient(t *testing.T) {
	// Test when getClient returns nil
	cmd := newCloseCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestCloseCmd_ZeroRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCloseCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetRunsResponse{{ID: 12345, Name: "Run 12345"}}, nil
		},
		CloseRunFunc: func(ctx context.Context, runID int64) (*data.Run, error) {
			assert.Equal(t, int64(12345), runID)
			return &data.Run{ID: runID, IsCompleted: true}, nil
		},
	}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCloseCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
	}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
