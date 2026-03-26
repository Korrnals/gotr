package configurations

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddConfigCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddConfigFunc: func(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
			assert.Equal(t, int64(5), groupID)
			assert.Equal(t, "Chrome", req.Name)
			return &data.Config{ID: 100, Name: req.Name}, nil
		},
	}

	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name", "Chrome"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddConfigCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name", "Firefox", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddConfigCmd_InvalidGroupID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddConfigCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddConfigCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		AddConfigFunc: func(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
			return nil, fmt.Errorf("group not found")
		},
	}

	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddConfigCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetConfigsResponse{{ID: 5, Name: "Browsers"}}, nil
		},
		AddConfigFunc: func(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
			assert.Equal(t, int64(5), groupID)
			assert.Equal(t, "Chrome", req.Name)
			return &data.Config{ID: 100, Name: req.Name}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--name", "Chrome"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddConfigCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddConfigCmd(getClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{"--name", "Chrome"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
