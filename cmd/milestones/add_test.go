package milestones

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты сухого запуска ====================

func TestAddCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name=Release 1.0", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_DryRun_WithDueDate(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name=Sprint 5", "--due-on=2026-03-01", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Функциональные тесты с моком ====================

func TestAddCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddMilestoneFunc: func(projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Release 1.0", req.Name)
			assert.Equal(t, "2026-03-01", req.DueOn)
			return &data.Milestone{ID: 100, Name: req.Name}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name=Release 1.0", "--due-on=2026-03-01"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_Success_Minimal(t *testing.T) {
	mock := &client.MockClient{
		AddMilestoneFunc: func(projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
			assert.Equal(t, int64(5), projectID)
			assert.Equal(t, "Sprint 1", req.Name)
			return &data.Milestone{ID: 200, Name: req.Name}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name=Sprint 1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_WithDescription(t *testing.T) {
	mock := &client.MockClient{
		AddMilestoneFunc: func(projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Release 2.0", req.Name)
			assert.Equal(t, "Major release", req.Description)
			return &data.Milestone{ID: 300, Name: req.Name, Description: "Major release"}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name=Release 2.0", "--description=Major release"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		AddMilestoneFunc: func(projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
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

// ==================== Тесты валидации ====================

func TestAddCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name=Test"})

	err := cmd.Execute()
	assert.Error(t, err)
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
