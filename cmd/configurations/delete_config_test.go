package configurations

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestDeleteConfigCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		DeleteConfigFunc: func(configID int64) error {
			assert.Equal(t, int64(10), configID)
			return nil
		},
	}

	cmd := newDeleteConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteConfigCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteConfigCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteConfigCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		DeleteConfigFunc: func(configID int64) error {
			return fmt.Errorf("конфигурация не найдена")
		},
	}

	cmd := newDeleteConfigCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}
