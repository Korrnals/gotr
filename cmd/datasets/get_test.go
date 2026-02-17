package datasets

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetDatasetFunc: func(datasetID int64) (*data.Dataset, error) {
			assert.Equal(t, int64(123), datasetID)
			return &data.Dataset{ID: 123, Name: "Test Data", ProjectID: 1}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_WithSaveFlag(t *testing.T) {
	mock := &client.MockClient{
		GetDatasetFunc: func(datasetID int64) (*data.Dataset, error) {
			return &data.Dataset{ID: 456, Name: "My Dataset"}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"456", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NotFound(t *testing.T) {
	mock := &client.MockClient{
		GetDatasetFunc: func(datasetID int64) (*data.Dataset, error) {
			return nil, fmt.Errorf("датасет не найден")
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "датасет не найден")
}

// ==================== Тесты валидации ====================

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный dataset_id")
}

func TestGetCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный dataset_id")
}

func TestGetCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
