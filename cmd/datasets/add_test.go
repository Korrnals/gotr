package datasets

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
		AddDatasetFunc: func(projectID int64, name string) (*data.Dataset, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "New Dataset", name)
			return &data.Dataset{ID: 100, Name: name, ProjectID: projectID}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "New Dataset"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_WithSaveFlag(t *testing.T) {
	mock := &client.MockClient{
		AddDatasetFunc: func(projectID int64, name string) (*data.Dataset, error) {
			return &data.Dataset{ID: 200, Name: name}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "My Data", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		AddDatasetFunc: func(projectID int64, name string) (*data.Dataset, error) {
			return nil, fmt.Errorf("проект не найден")
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "проект не найден")
}

// ==================== Dry-run тесты ====================

func TestAddCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты валидации ====================

func TestAddCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный project_id")
}

func TestAddCmd_ZeroProjectID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный project_id")
}

func TestAddCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

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
	assert.Contains(t, err.Error(), "--name обязателен")
}
