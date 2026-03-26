package plans

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Dry Run Tests ====================

func TestCloseCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newCloseCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Functional Tests with Mock ====================

func TestCloseCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		ClosePlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
			assert.Equal(t, int64(12345), planID)
			return &data.Plan{ID: 12345, Name: "Test Plan", IsCompleted: true}, nil
		},
	}

	cmd := newCloseCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCloseCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		ClosePlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
			return nil, fmt.Errorf("plan already closed")
		},
	}

	cmd := newCloseCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already closed")
}

// ==================== Validation Tests ====================

func TestCloseCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newCloseCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCloseCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newCloseCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCloseCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return data.GetPlansResponse{{ID: 100, Name: "Plan 1"}}, nil
		},
	}
	cmd := newCloseCmd(getClientForTests)
	ctx := interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter())
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestCloseCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return data.GetPlansResponse{{ID: 100, Name: "Plan 1"}}, nil
		},
		ClosePlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
			assert.Equal(t, int64(100), planID)
			return &data.Plan{ID: 100, Name: "Plan 1"}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newCloseCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}
