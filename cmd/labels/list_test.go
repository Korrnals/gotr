// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Success scenario tests ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Label{
				{ID: 1, Name: "Bug"},
				{ID: 2, Name: "Feature"},
				{ID: 3, Name: "Critical"},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Bug")
	assert.Contains(t, output, "Feature")
	assert.Contains(t, output, "Critical")
}

func TestListCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No labels found")
}

func TestListCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{
				{ID: 1, Name: "Bug"},
				{ID: 2, Name: "Feature"},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_WithSaveFlag(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{
				{ID: 1, Name: "Bug"},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== API error tests ====================

func TestListCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return nil, fmt.Errorf("API connection error")
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API connection error")
}

func TestListCmd_ProjectNotFound(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

// ==================== Validation tests ====================

func TestListCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestListCmd_ZeroProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestListCmd_NegativeProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	// Use -- to stop flag parsing so -1 is treated as an argument
	cmd.SetArgs([]string{"--", "-1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestListCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Label{{ID: 10, Name: "Bug"}}, nil
		},
	}
	p := interactive.NewMockPrompter().WithSelectResponses(
		interactive.SelectResponse{Index: 0}, // project
	)
	cmd := newListCmd(getClientForTests)
	var buf bytes.Buffer
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Bug")
}

func TestListCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestListCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project_id is required in non-interactive mode")
}

func TestListCmd_TooManyArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "2"})

	err := cmd.Execute()
	assert.Error(t, err)
}
