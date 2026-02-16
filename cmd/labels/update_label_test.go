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

func TestUpdateLabelCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateLabelFunc: func(labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
			assert.Equal(t, int64(1), labelID)
			assert.Equal(t, int64(10), req.ProjectID)
			assert.Equal(t, "Updated Label", req.Title)
			return &data.Label{
				ID:   1,
				Name: "Updated Label",
			}, nil
		},
	}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--project", "10", "--title", "Updated Label"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateLabelCmd_WithShortFlags(t *testing.T) {
	mock := &client.MockClient{
		UpdateLabelFunc: func(labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
			assert.Equal(t, int64(5), labelID)
			assert.Equal(t, int64(20), req.ProjectID)
			assert.Equal(t, "New Title", req.Title)
			return &data.Label{
				ID:   5,
				Name: "New Title",
			}, nil
		},
	}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "-p", "20", "-t", "New Title"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateLabelCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		UpdateLabelFunc: func(labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
			return &data.Label{ID: 3, Name: "Saved Label"}, nil
		},
	}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"3", "--project", "1", "--title", "Saved Label", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateLabelCmd_WithSaveFlag(t *testing.T) {
	mock := &client.MockClient{
		UpdateLabelFunc: func(labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
			return &data.Label{ID: 7, Name: "JSON Label"}, nil
		},
	}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"7", "--project", "1", "--title", "JSON Label", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты ошибок API ====================

func TestUpdateLabelCmd_NotFound(t *testing.T) {
	mock := &client.MockClient{
		UpdateLabelFunc: func(labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
			return nil, fmt.Errorf("label not found")
		},
	}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--project", "1", "--title", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateLabelCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		UpdateLabelFunc: func(labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
			return nil, fmt.Errorf("permission denied")
		},
	}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--project", "1", "--title", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestUpdateLabelCmd_ProjectNotFound(t *testing.T) {
	mock := &client.MockClient{
		UpdateLabelFunc: func(labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--project", "99999", "--title", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

// ==================== Тесты валидации ====================

func TestUpdateLabelCmd_InvalidLabelID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--project", "1", "--title", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid label_id")
}

func TestUpdateLabelCmd_ZeroLabelID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "--project", "1", "--title", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid label_id")
}

func TestUpdateLabelCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateLabelCmd_TooManyArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "2", "--project", "1", "--title", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateLabelCmd_MissingProjectFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	// Missing --project flag
	cmd.SetArgs([]string{"1", "--title", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateLabelCmd_MissingTitleFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	// Missing --title flag
	cmd.SetArgs([]string{"1", "--project", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateLabelCmd_EmptyTitle(t *testing.T) {
	mock := &client.MockClient{
		UpdateLabelFunc: func(labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
			// API might return an error for empty title
			if req.Title == "" {
				return nil, fmt.Errorf("title cannot be empty")
			}
			return &data.Label{ID: labelID, Name: req.Title}, nil
		},
	}

	cmd := newUpdateLabelCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--project", "1", "--title", ""})

	err := cmd.Execute()
	// Should still call the API, which may return an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title cannot be empty")
}
