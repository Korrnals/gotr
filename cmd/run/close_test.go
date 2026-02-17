package run

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCloseCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		CloseRunFunc: func(runID int64) (*data.Run, error) {
			assert.Equal(t, int64(12345), runID)
			return &data.Run{ID: runID, IsCompleted: true}, nil
		},
	}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCloseCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCloseCmd_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCloseCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		CloseRunFunc: func(runID int64) (*data.Run, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCloseCmd_NilClient(t *testing.T) {
	// Test when getClient returns nil
	cmd := newCloseCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestCloseCmd_ZeroRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCloseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}
