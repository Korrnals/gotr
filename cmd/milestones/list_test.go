package milestones

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Functional tests with mock ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Milestone{
				{ID: 100, Name: "Release 1.0"},
				{ID: 200, Name: "Release 2.0"},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Validation tests ====================

func TestListCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Milestone{{ID: 100, Name: "Release 1.0"}}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
