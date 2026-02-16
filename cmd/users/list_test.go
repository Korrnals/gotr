package users

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetUsersFunc: func() (data.GetUsersResponse, error) {
			return []data.User{
				{ID: 1, Name: "John Doe", Email: "john@example.com", IsActive: true, Role: "Lead"},
				{ID: 2, Name: "Jane Smith", Email: "jane@example.com", IsActive: true, Role: "Tester"},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetUsersFunc: func() (data.GetUsersResponse, error) {
			return []data.User{}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		GetUsersFunc: func() (data.GetUsersResponse, error) {
			return nil, fmt.Errorf("API error")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_WithProjectID(t *testing.T) {
	mock := &client.MockClient{
		GetUsersByProjectFunc: func(projectID int64) (data.GetUsersResponse, error) {
			assert.Equal(t, int64(123), projectID)
			return []data.User{
				{ID: 1, Name: "Project User", Email: "user@example.com", IsActive: true, Role: "Tester"},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_WithProjectID_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetUsersByProjectFunc: func(projectID int64) (data.GetUsersResponse, error) {
			return []data.User{}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_WithProjectID_Error(t *testing.T) {
	mock := &client.MockClient{
		GetUsersByProjectFunc: func(projectID int64) (data.GetUsersResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты валидации ====================

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


