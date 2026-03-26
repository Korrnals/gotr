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

// ==================== Функциональные тесты с моком ====================

func TestUpdateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateMilestoneFunc: func(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
			assert.Equal(t, int64(12345), milestoneID)
			assert.Equal(t, "Updated Name", req.Name)
			return &data.Milestone{ID: 12345, Name: "Updated Name"}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--name", "Updated Name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_AllFlags(t *testing.T) {
	mock := &client.MockClient{
		UpdateMilestoneFunc: func(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
			assert.Equal(t, int64(12345), milestoneID)
			assert.Equal(t, "Updated", req.Name)
			assert.Equal(t, "Description", req.Description)
			assert.Equal(t, "2025-12-31", req.DueOn)
			assert.True(t, req.IsCompleted)
			return &data.Milestone{ID: 12345, Name: "Updated"}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{
		"12345",
		"--name", "Updated",
		"--description", "Description",
		"--due-on", "2025-12-31",
		"--is-completed=true",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateMilestoneFunc: func(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
			return nil, fmt.Errorf("milestone not found")
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты сухого запуска ====================

func TestUpdateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты валидации ====================

func TestUpdateCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Milestone{{ID: 100, Name: "Release 1.0"}}, nil
		},
		UpdateMilestoneFunc: func(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
			assert.Equal(t, int64(100), milestoneID)
			assert.Equal(t, "New Name", req.Name)
			return &data.Milestone{ID: 100, Name: "New Name"}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
		)

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--name", "New Name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"--name", "New Name"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
