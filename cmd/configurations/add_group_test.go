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

func TestAddGroupCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddConfigGroupFunc: func(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Browsers", req.Name)
			return &data.ConfigGroup{ID: 10, Name: req.Name}, nil
		},
	}

	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "Browsers"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddGroupCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		AddConfigGroupFunc: func(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
			return &data.ConfigGroup{ID: 5, Name: req.Name}, nil
		},
	}

	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "OS", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddGroupCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddGroupCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddGroupCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name is required")
}

func TestAddGroupCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		AddConfigGroupFunc: func(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddGroupCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		AddConfigGroupFunc: func(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Browsers", req.Name)
			return &data.ConfigGroup{ID: 10, Name: req.Name}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--name", "Browsers"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddGroupCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddGroupCmd(getClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{"--name", "Browsers"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
