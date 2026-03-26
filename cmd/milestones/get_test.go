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

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetMilestoneFunc: func(ctx context.Context, milestoneID int64) (*data.Milestone, error) {
			assert.Equal(t, int64(12345), milestoneID)
			return &data.Milestone{ID: 12345, Name: "Release 1.0", ProjectID: 1}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NotFound(t *testing.T) {
	mock := &client.MockClient{
		GetMilestoneFunc: func(ctx context.Context, milestoneID int64) (*data.Milestone, error) {
			return nil, fmt.Errorf("milestone not found")
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "milestone not found")
}

// ==================== Тесты валидации ====================

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Milestone{{ID: 100, Name: "Release 1.0"}}, nil
		},
		GetMilestoneFunc: func(ctx context.Context, milestoneID int64) (*data.Milestone, error) {
			assert.Equal(t, int64(100), milestoneID)
			return &data.Milestone{ID: 100, Name: "Release 1.0", ProjectID: 1}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
		)

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
