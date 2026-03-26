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

func TestUpdateGroupCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateConfigGroupFunc: func(ctx context.Context, groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
			assert.Equal(t, int64(5), groupID)
			assert.Equal(t, "New Name", req.Name)
			return &data.ConfigGroup{ID: 5, Name: req.Name}, nil
		},
	}

	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name", "New Name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateGroupCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateGroupCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateGroupCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateGroupCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateConfigGroupFunc: func(ctx context.Context, groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
			return nil, fmt.Errorf("group not found")
		},
	}

	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateGroupCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetConfigsResponse{{ID: 5, Name: "Browsers"}}, nil
		},
		UpdateConfigGroupFunc: func(ctx context.Context, groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
			assert.Equal(t, int64(5), groupID)
			assert.Equal(t, "New Name", req.Name)
			return &data.ConfigGroup{ID: groupID, Name: req.Name}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--name", "New Name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateGroupCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateGroupCmd(getClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{"--name", "New Name"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
