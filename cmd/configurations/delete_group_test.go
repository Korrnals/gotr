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

func TestDeleteGroupCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		DeleteConfigGroupFunc: func(ctx context.Context, groupID int64) error {
			assert.Equal(t, int64(5), groupID)
			return nil
		},
	}

	cmd := newDeleteGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteGroupCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteGroupCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteGroupCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		DeleteConfigGroupFunc: func(ctx context.Context, groupID int64) error {
			return fmt.Errorf("group not found")
		},
	}

	cmd := newDeleteGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteGroupCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetConfigsResponse{{ID: 5, Name: "Browsers"}}, nil
		},
		DeleteConfigGroupFunc: func(ctx context.Context, groupID int64) error {
			assert.Equal(t, int64(5), groupID)
			return nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newDeleteGroupCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteGroupCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteGroupCmd(getClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
