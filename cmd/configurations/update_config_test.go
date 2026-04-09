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

func TestUpdateConfigCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateConfigFunc: func(ctx context.Context, configID int64, req *data.UpdateConfigRequest) (*data.Config, error) {
			assert.Equal(t, int64(10), configID)
			assert.Equal(t, "Chrome 120", req.Name)
			return &data.Config{ID: 10, Name: req.Name}, nil
		},
	}

	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10", "--name", "Chrome 120"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateConfigCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateConfigCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateConfigCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateConfigCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateConfigFunc: func(ctx context.Context, configID int64, req *data.UpdateConfigRequest) (*data.Config, error) {
			return nil, fmt.Errorf("configuration not found")
		},
	}

	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateConfigCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetConfigsResponse{{
				ID:      5,
				Name:    "Browsers",
				Configs: []data.Config{{ID: 10, Name: "Chrome"}},
			}}, nil
		},
		UpdateConfigFunc: func(ctx context.Context, configID int64, req *data.UpdateConfigRequest) (*data.Config, error) {
			assert.Equal(t, int64(10), configID)
			assert.Equal(t, "Chrome 120", req.Name)
			return &data.Config{ID: configID, Name: req.Name}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--name", "Chrome 120"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateConfigCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateConfigCmd(getClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{"--name", "Chrome 120"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
