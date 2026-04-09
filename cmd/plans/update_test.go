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

func TestUpdateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--name=New Name", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_DryRun_NoFlags(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Functional Tests with Mock ====================

func TestUpdateCmd_Success_Name(t *testing.T) {
	mock := &client.MockClient{
		UpdatePlanFunc: func(ctx context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
			assert.Equal(t, int64(12345), planID)
			assert.Equal(t, "Updated Plan Name", req.Name)
			return &data.Plan{ID: 12345, Name: req.Name}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--name=Updated Plan Name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_Success_Description(t *testing.T) {
	mock := &client.MockClient{
		UpdatePlanFunc: func(ctx context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
			assert.Equal(t, int64(12345), planID)
			assert.Equal(t, "New Description", req.Description)
			return &data.Plan{ID: 12345, Description: req.Description}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--description=New Description"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_Success_Milestone(t *testing.T) {
	mock := &client.MockClient{
		UpdatePlanFunc: func(ctx context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
			assert.Equal(t, int64(12345), planID)
			assert.Equal(t, int64(20), req.MilestoneID)
			return &data.Plan{ID: 12345, MilestoneID: 20}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--milestone-id=20"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdatePlanFunc: func(ctx context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
			return nil, fmt.Errorf("plan not found")
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999", "--name=New"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Validation Tests ====================

func TestUpdateCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name=New"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return data.GetPlansResponse{{ID: 100, Name: "Plan 1"}}, nil
		},
	}
	cmd := newUpdateCmd(getClientForTests)
	ctx := interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter())
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestUpdateCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return data.GetPlansResponse{{ID: 100, Name: "Plan 1"}}, nil
		},
		UpdatePlanFunc: func(ctx context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
			assert.Equal(t, int64(100), planID)
			return &data.Plan{ID: 100, Name: "Updated Plan"}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--name", "Updated Plan"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
