package plans

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Dry Run Tests ====================

func TestAddCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name=Test Plan", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_DryRun_WithDescription(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name=Test Plan", "--description=Test Desc", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Functional Tests with Mock ====================

func TestAddCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddPlanFunc: func(projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Sprint 1 Plan", req.Name)
			assert.Equal(t, "Full regression", req.Description)
			return &data.Plan{ID: 100, Name: req.Name}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name=Sprint 1 Plan", "--description=Full regression"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_Success_Minimal(t *testing.T) {
	mock := &client.MockClient{
		AddPlanFunc: func(projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
			assert.Equal(t, int64(5), projectID)
			assert.Equal(t, "Minimal Plan", req.Name)
			return &data.Plan{ID: 200, Name: req.Name}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name=Minimal Plan"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_WithMilestone(t *testing.T) {
	mock := &client.MockClient{
		AddPlanFunc: func(projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Plan with Milestone", req.Name)
			assert.Equal(t, int64(10), req.MilestoneID)
			return &data.Plan{ID: 300, Name: req.Name, MilestoneID: 10}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name=Plan with Milestone", "--milestone-id=10"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		AddPlanFunc: func(projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name=Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

// ==================== Validation Tests ====================

func TestAddCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name=Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestAddCmd_ZeroProjectID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "--name=Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestAddCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
