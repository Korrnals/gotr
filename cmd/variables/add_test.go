package variables

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddVariableFunc: func(datasetID int64, name string) (*data.Variable, error) {
			assert.Equal(t, int64(123), datasetID)
			assert.Equal(t, "username", name)
			return &data.Variable{ID: 1, Name: name, DatasetID: datasetID}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "username"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_WithOutput(t *testing.T) {
	mock := &client.MockClient{
		AddVariableFunc: func(datasetID int64, name string) (*data.Variable, error) {
			return &data.Variable{ID: 5, Name: name}, nil
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "var.json")

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "email", "-o", outputFile})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "email")
}

func TestAddCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_InvalidDatasetID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name обязателен")
}

func TestAddCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		AddVariableFunc: func(datasetID int64, name string) (*data.Variable, error) {
			return nil, fmt.Errorf("датасет не найден")
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}
