package groups

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestAddCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddGroupFunc: func(projectID int64, name string, userIDs []int64) (*data.Group, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "QA Team", name)
			return &data.Group{
				ID:        1,
				Name:      name,
				ProjectID: projectID,
				UserIDs:   []int64{},
			}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "QA Team"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_WithDifferentName(t *testing.T) {
	mock := &client.MockClient{
		AddGroupFunc: func(projectID int64, name string, userIDs []int64) (*data.Group, error) {
			assert.Equal(t, int64(5), projectID)
			assert.Equal(t, "Developers", name)
			return &data.Group{
				ID:        2,
				Name:      name,
				ProjectID: projectID,
				UserIDs:   []int64{},
			}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name", "Developers"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "Test Group", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		AddGroupFunc: func(projectID int64, name string, userIDs []int64) (*data.Group, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test Group"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Тесты валидации ====================

func TestAddCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test Group"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project_id должен быть положительным числом")
}

func TestAddCmd_ZeroProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "--name", "Test Group"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project_id должен быть положительным числом")
}

func TestAddCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name обязателен")
}

func TestAddCmd_EmptyName(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", ""})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name обязателен")
}

func TestAddCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
