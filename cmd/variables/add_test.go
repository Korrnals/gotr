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

func TestAddCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddVariableFunc: func(ctx context.Context, datasetID int64, name string) (*data.Variable, error) {
			assert.Equal(t, int64(123), datasetID)
			assert.Equal(t, "username", name)
			return &data.Variable{ID: 1, Name: name, DatasetID: datasetID}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "username"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		AddVariableFunc: func(ctx context.Context, datasetID int64, name string) (*data.Variable, error) {
			return &data.Variable{ID: 5, Name: name}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "email", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_InvalidDatasetID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name is required")
}

func TestAddCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		AddVariableFunc: func(ctx context.Context, datasetID int64, name string) (*data.Variable, error) {
			return nil, fmt.Errorf("dataset not found")
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}
	cmd := newAddCmd(getClientForTests)
	ctx := interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter())
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--name", "username"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestAddCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetDatasetsResponse{{ID: 123, Name: "Dataset 123", ProjectID: 1}}, nil
		},
		AddVariableFunc: func(ctx context.Context, datasetID int64, name string) (*data.Variable, error) {
			assert.Equal(t, int64(123), datasetID)
			assert.Equal(t, "username", name)
			return &data.Variable{ID: 1, Name: name, DatasetID: datasetID}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--name", "username"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
