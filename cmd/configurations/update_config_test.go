package configurations

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestUpdateConfigCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateConfigFunc: func(configID int64, req *data.UpdateConfigRequest) (*data.Config, error) {
			assert.Equal(t, int64(10), configID)
			assert.Equal(t, "Chrome 120", req.Name)
			return &data.Config{ID: 10, Name: req.Name}, nil
		},
	}

	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10", "--name", "Chrome 120"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateConfigCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateConfigCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateConfigCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateConfigCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateConfigFunc: func(configID int64, req *data.UpdateConfigRequest) (*data.Config, error) {
			return nil, fmt.Errorf("конфигурация не найдена")
		},
	}

	cmd := newUpdateConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}
