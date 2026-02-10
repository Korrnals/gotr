package datasets

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestDeleteCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		DeleteDatasetFunc: func(datasetID int64) error {
			assert.Equal(t, int64(123), datasetID)
			return nil
		},
	}

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		DeleteDatasetFunc: func(datasetID int64) error {
			return fmt.Errorf("датасет не найден")
		},
	}

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "датасет не найден")
}

// ==================== Dry-run тесты ====================

func TestDeleteCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты валидации ====================

func TestDeleteCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный dataset_id")
}

func TestDeleteCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный dataset_id")
}

func TestDeleteCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
