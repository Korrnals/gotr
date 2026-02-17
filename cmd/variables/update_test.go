package variables

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateVariableFunc: func(variableID int64, name string) (*data.Variable, error) {
			assert.Equal(t, int64(789), variableID)
			assert.Equal(t, "new_name", name)
			return &data.Variable{ID: 789, Name: name}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789", "--name", "new_name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		UpdateVariableFunc: func(variableID int64, name string) (*data.Variable, error) {
			return &data.Variable{ID: 789, Name: name}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789", "--name", "updated", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name обязателен")
}

func TestUpdateCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateVariableFunc: func(variableID int64, name string) (*data.Variable, error) {
			return nil, fmt.Errorf("переменная не найдена")
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}
