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

// ==================== Тесты для get projects ====================

func TestProjectsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func() (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1", SuiteMode: 1},
				{ID: 2, Name: "Project 2", SuiteMode: 2},
				{ID: 3, Name: "Project 3", SuiteMode: 3},
			}, nil
		},
	}

	cmd := newProjectsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestProjectsCmd_EmptyList(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func() (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{}, nil
		},
	}

	cmd := newProjectsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestProjectsCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func() (data.GetProjectsResponse, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	cmd := newProjectsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// ==================== Тесты для get project ====================

func TestProjectCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return &data.GetProjectResponse{
				ID:        30,
				Name:      "Test Project",
				SuiteMode: 2,
			}, nil
		},
	}

	cmd := newProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestProjectCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID проекта")
}

func TestProjectCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestProjectCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestProjectsCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newProjectsCmd(nilClientFunc)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestProjectCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newProjectCmd(nilClientFunc)
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}
