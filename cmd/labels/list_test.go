// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты успешных сценариев ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Label{
				{ID: 1, Name: "Bug"},
				{ID: 2, Name: "Feature"},
				{ID: 3, Name: "Critical"},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Bug")
	assert.Contains(t, output, "Feature")
	assert.Contains(t, output, "Critical")
}

func TestListCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No labels found")
}

func TestListCmd_WithJSONOutput(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{
				{ID: 1, Name: "Bug"},
				{ID: 2, Name: "Feature"},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "-o", "json"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_WithOutputFile(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{
				{ID: 1, Name: "Bug"},
			}, nil
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "labels.json")

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	// Note: list command supports -o json for JSON output, not file path
	// When a file path is provided, it outputs table format to stdout
	cmd.SetArgs([]string{"1", "-o", outputFile})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	// Output goes to stdout in table format, not to file
	assert.Contains(t, buf.String(), "Bug")
}

// ==================== Тесты ошибок API ====================

func TestListCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			return nil, fmt.Errorf("API connection error")
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API connection error")
}

func TestListCmd_ProjectNotFound(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

// ==================== Тесты валидации ====================

func TestListCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestListCmd_ZeroProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestListCmd_NegativeProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	// Use -- to stop flag parsing so -1 is treated as an argument
	cmd.SetArgs([]string{"--", "-1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestListCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_TooManyArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "2"})

	err := cmd.Execute()
	assert.Error(t, err)
}
