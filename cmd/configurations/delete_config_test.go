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

func TestDeleteConfigCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		DeleteConfigFunc: func(ctx context.Context, configID int64) error {
			assert.Equal(t, int64(10), configID)
			return nil
		},
	}

	cmd := newDeleteConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteConfigCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteConfigCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteConfigCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		DeleteConfigFunc: func(ctx context.Context, configID int64) error {
			return fmt.Errorf("configuration not found")
		},
	}

	cmd := newDeleteConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteConfigCmd_NoArgs_Interactive(t *testing.T) {
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
		DeleteConfigFunc: func(ctx context.Context, configID int64) error {
			assert.Equal(t, int64(10), configID)
			return nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newDeleteConfigCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteConfigCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteConfigCmd(getClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
