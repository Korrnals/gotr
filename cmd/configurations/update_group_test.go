package configurations

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestUpdateGroupCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateConfigGroupFunc: func(groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
			assert.Equal(t, int64(5), groupID)
			assert.Equal(t, "New Name", req.Name)
			return &data.ConfigGroup{ID: 5, Name: req.Name}, nil
		},
	}

	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name", "New Name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateGroupCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateGroupCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateGroupCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateGroupCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateConfigGroupFunc: func(groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
			return nil, fmt.Errorf("группа не найдена")
		},
	}

	cmd := newUpdateGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}
