package groups

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestUpdateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateGroupFunc: func(groupID int64, name string, userIDs []int64) (*data.Group, error) {
			assert.Equal(t, int64(123), groupID)
			assert.Equal(t, "Updated Group Name", name)
			return &data.Group{
				ID:      groupID,
				Name:    name,
				UserIDs: []int64{101, 102},
			}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "Updated Group Name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_DifferentGroup(t *testing.T) {
	mock := &client.MockClient{
		UpdateGroupFunc: func(groupID int64, name string, userIDs []int64) (*data.Group, error) {
			assert.Equal(t, int64(456), groupID)
			assert.Equal(t, "New Name", name)
			return &data.Group{
				ID:      groupID,
				Name:    name,
				UserIDs: []int64{201, 202},
			}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"456", "--name", "New Name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "Test Name", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		UpdateGroupFunc: func(groupID int64, name string, userIDs []int64) (*data.Group, error) {
			return nil, fmt.Errorf("group not found")
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test Name"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Тесты валидации ====================

func TestUpdateCmd_InvalidGroupID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test Name"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "group_id должен быть положительным числом")
}

func TestUpdateCmd_ZeroGroupID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "--name", "Test Name"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "group_id должен быть положительным числом")
}

func TestUpdateCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name обязателен")
}

func TestUpdateCmd_EmptyName(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", ""})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name обязателен")
}

func TestUpdateCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
