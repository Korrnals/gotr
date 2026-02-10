package variables

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		DeleteVariableFunc: func(variableID int64) error {
			assert.Equal(t, int64(789), variableID)
			return nil
		},
	}

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"789", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		DeleteVariableFunc: func(variableID int64) error {
			return fmt.Errorf("переменная не найдена")
		},
	}

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}
