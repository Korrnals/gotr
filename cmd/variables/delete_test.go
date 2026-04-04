package variables

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		DeleteVariableFunc: func(ctx context.Context, variableID int64) error {
			assert.Equal(t, int64(789), variableID)
			return nil
		},
	}

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}
	cmd := newDeleteCmd(getClientForTests)
	ctx := interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter())
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestDeleteCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable_id is required in non-interactive mode")
}

func TestDeleteCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetDatasetsResponse{{ID: 123, Name: "Dataset 123", ProjectID: 1}}, nil
		},
		GetVariablesFunc: func(ctx context.Context, datasetID int64) (data.GetVariablesResponse, error) {
			assert.Equal(t, int64(123), datasetID)
			return data.GetVariablesResponse{{ID: 789, Name: "username", DatasetID: 123}}, nil
		},
		DeleteVariableFunc: func(ctx context.Context, variableID int64) error {
			assert.Equal(t, int64(789), variableID)
			return nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		DeleteVariableFunc: func(ctx context.Context, variableID int64) error {
			return fmt.Errorf("variable not found")
		},
	}

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}
