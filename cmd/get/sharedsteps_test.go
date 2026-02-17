package get

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для get sharedsteps ====================

func TestSharedStepsCmd_WithProjectID(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(projectID int64) (data.GetSharedStepsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSharedStepsResponse{
				{ID: 1, Title: "Shared Step 1", ProjectID: 30},
				{ID: 2, Title: "Shared Step 2", ProjectID: 30},
			}, nil
		},
	}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSharedStepsCmd_WithProjectIDFlag(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(projectID int64) (data.GetSharedStepsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSharedStepsResponse{
				{ID: 1, Title: "Shared Step 1", ProjectID: 30},
			}, nil
		},
	}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--project-id", "30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSharedStepsCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID проекта")
}

func TestSharedStepsCmd_InvalidProjectIDFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--project-id", "invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSharedStepsCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(projectID int64) (data.GetSharedStepsResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Тесты для get sharedstep ====================

func TestSharedStepCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepFunc: func(stepID int64) (*data.SharedStep, error) {
			assert.Equal(t, int64(12345), stepID)
			return &data.SharedStep{
				ID:        12345,
				Title:     "Test Shared Step",
				ProjectID: 30,
			}, nil
		},
	}

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSharedStepCmd_InvalidStepID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID шага")
}

func TestSharedStepCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSharedStepCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepFunc: func(stepID int64) (*data.SharedStep, error) {
			return nil, fmt.Errorf("shared step not found")
		},
	}

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSharedStepsCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newSharedStepsCmd(nilClientFunc)
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestSharedStepCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newSharedStepCmd(nilClientFunc)
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}
