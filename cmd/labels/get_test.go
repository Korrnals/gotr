// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты успешных сценариев ====================

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetLabelFunc: func(labelID int64) (*data.Label, error) {
			assert.Equal(t, int64(1), labelID)
			return &data.Label{
				ID:   1,
				Name: "Bug",
			}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetLabelFunc: func(labelID int64) (*data.Label, error) {
			return &data.Label{ID: 5, Name: "Critical"}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_WithSaveFlag(t *testing.T) {
	mock := &client.MockClient{
		GetLabelFunc: func(labelID int64) (*data.Label, error) {
			return &data.Label{ID: 10, Name: "Regression"}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты ошибок API ====================

func TestGetCmd_NotFound(t *testing.T) {
	mock := &client.MockClient{
		GetLabelFunc: func(labelID int64) (*data.Label, error) {
			return nil, fmt.Errorf("label not found")
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetLabelFunc: func(labelID int64) (*data.Label, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// ==================== Тесты валидации ====================

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid label_id")
}

func TestGetCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid label_id")
}

func TestGetCmd_NegativeID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	// Use -- to stop flag parsing so -1 is treated as an argument
	cmd.SetArgs([]string{"--", "-1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid label_id")
}

func TestGetCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_TooManyArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "2"})

	err := cmd.Execute()
	assert.Error(t, err)
}
