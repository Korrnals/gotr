package users

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateUserFunc: func(userID int64, req data.UpdateUserRequest) (*data.User, error) {
			assert.Equal(t, int64(123), userID)
			assert.Equal(t, "Updated Name", req.Name)
			return &data.User{
				ID:       123,
				Name:     req.Name,
				Email:    "original@test.com",
				IsActive: true,
			}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "Updated Name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_SetAdmin(t *testing.T) {
	mock := &client.MockClient{
		UpdateUserFunc: func(userID int64, req data.UpdateUserRequest) (*data.User, error) {
			assert.Equal(t, int64(123), userID)
			assert.Equal(t, 1, req.IsAdmin)
			return &data.User{
				ID:      123,
				Name:    "Test User",
				IsAdmin: true,
			}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--admin"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_Deactivate(t *testing.T) {
	mock := &client.MockClient{
		UpdateUserFunc: func(userID int64, req data.UpdateUserRequest) (*data.User, error) {
			assert.Equal(t, int64(123), userID)
			assert.Equal(t, 0, req.IsActive)
			return &data.User{
				ID:       123,
				Name:     "Test User",
				IsActive: false,
			}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--inactive"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_Error(t *testing.T) {
	mock := &client.MockClient{
		UpdateUserFunc: func(userID int64, req data.UpdateUserRequest) (*data.User, error) {
			return nil, fmt.Errorf("user not found")
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}
