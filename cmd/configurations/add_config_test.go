package configurations

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddConfigCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddConfigFunc: func(groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
			assert.Equal(t, int64(5), groupID)
			assert.Equal(t, "Chrome", req.Name)
			return &data.Config{ID: 100, Name: req.Name}, nil
		},
	}

	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name", "Chrome"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddConfigCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name", "Firefox", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddConfigCmd_InvalidGroupID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddConfigCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddConfigCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		AddConfigFunc: func(groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
			return nil, fmt.Errorf("группа не найдена")
		},
	}

	cmd := newAddConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}
