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

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetRunsResponse{
				{ID: 1, Name: "Run 1", ProjectID: 30, PassedCount: 10, FailedCount: 2},
				{ID: 2, Name: "Run 2", ProjectID: 30, PassedCount: 5, FailedCount: 0},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_ZeroProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_NegativeProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"-5"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListCmd_EmptyList(t *testing.T) {
	mock := &client.MockClient{
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_InteractiveMode(t *testing.T) {
	// Test interactive mode - it will fail since we don't have real projects
	// but we verify the code path is taken
	mock := &client.MockClient{
		GetProjectsFunc: func() (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 30, Name: "Test Project"},
			}, nil
		},
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{
				{ID: 1, Name: "Run 1", ProjectID: 30},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	// No args triggers interactive mode
	cmd.SetArgs([]string{})

	// This will fail in test because interactive mode requires stdin
	// but we verify the path is executed
	err := cmd.Execute()
	// Expect error because interactive mode needs real stdin
	assert.Error(t, err)
}

func TestListCmd_WithLargeProjectID(t *testing.T) {
	mock := &client.MockClient{
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(999999999), projectID)
			return data.GetRunsResponse{
				{ID: 1, Name: "Large Project Run", ProjectID: projectID},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999999999"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_NilClient(t *testing.T) {
	// Test when getClient returns nil
	cmd := newListCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestListCmd_MockClientTypeAssertion(t *testing.T) {
	// Test interactive mode with mock client (non-HTTPClient type)
	mock := &client.MockClient{
		GetProjectsFunc: func() (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 30, Name: "Test Project"},
			}, nil
		},
	}

	cmd := newListCmd(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	// No args triggers interactive mode
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	// Should fail with specific error about interactive mode
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "интерактивный режим недоступен в тестовом режиме")
}

func TestListCmd_InvalidProjectIDFormat(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"abc"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID проекта")
}
