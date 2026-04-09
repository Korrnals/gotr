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

func TestUpdateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateVariableFunc: func(ctx context.Context, variableID int64, name string) (*data.Variable, error) {
			assert.Equal(t, int64(789), variableID)
			assert.Equal(t, "new_name", name)
			return &data.Variable{ID: 789, Name: name}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789", "--name", "new_name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		UpdateVariableFunc: func(ctx context.Context, variableID int64, name string) (*data.Variable, error) {
			return &data.Variable{ID: 789, Name: name}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789", "--name", "updated", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name is required")
}

func TestUpdateCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateVariableFunc: func(ctx context.Context, variableID int64, name string) (*data.Variable, error) {
			return nil, fmt.Errorf("variable not found")
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}
	cmd := newUpdateCmd(getClientForTests)
	ctx := interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter())
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--name", "new_name"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestUpdateCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--name", "new_name"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable_id is required in non-interactive mode")
}

func TestUpdateCmd_ResolveInteractiveError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, fmt.Errorf("projects boom")
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewMockPrompter()))
	cmd.SetArgs([]string{"--name", "new_name"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_NoArgs_Interactive(t *testing.T) {
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
			return data.GetVariablesResponse{{ID: 789, Name: "old_name", DatasetID: 123}}, nil
		},
		UpdateVariableFunc: func(ctx context.Context, variableID int64, name string) (*data.Variable, error) {
			assert.Equal(t, int64(789), variableID)
			assert.Equal(t, "new_name", name)
			return &data.Variable{ID: 789, Name: name}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--name", "new_name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
