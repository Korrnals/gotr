package configurations

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestDeleteGroupCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		DeleteConfigGroupFunc: func(groupID int64) error {
			assert.Equal(t, int64(5), groupID)
			return nil
		},
	}

	cmd := newDeleteGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteGroupCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteGroupCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteGroupCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		DeleteConfigGroupFunc: func(groupID int64) error {
			return fmt.Errorf("группа не найдена")
		},
	}

	cmd := newDeleteGroupCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}
