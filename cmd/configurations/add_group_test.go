package configurations

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddGroupCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddConfigGroupFunc: func(projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Browsers", req.Name)
			return &data.ConfigGroup{ID: 10, Name: req.Name}, nil
		},
	}

	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "Browsers"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddGroupCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		AddConfigGroupFunc: func(projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
			return &data.ConfigGroup{ID: 5, Name: req.Name}, nil
		},
	}

	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "OS", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddGroupCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddGroupCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddGroupCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name обязателен")
}

func TestAddGroupCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		AddConfigGroupFunc: func(projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
			return nil, fmt.Errorf("проект не найден")
		},
	}

	cmd := newAddGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}
