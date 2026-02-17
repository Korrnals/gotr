package users

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddUserFunc: func(req data.AddUserRequest) (*data.User, error) {
			assert.Equal(t, "John Doe", req.Name)
			assert.Equal(t, "john@example.com", req.Email)
			return &data.User{
				ID:       123,
				Name:     req.Name,
				Email:    req.Email,
				IsActive: true,
			}, nil
		},
	}

	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--name", "John Doe", "--email", "john@example.com"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_WithAdmin(t *testing.T) {
	mock := &client.MockClient{
		AddUserFunc: func(req data.AddUserRequest) (*data.User, error) {
			assert.Equal(t, "Admin User", req.Name)
			assert.Equal(t, 1, req.IsAdmin)
			return &data.User{
				ID:       124,
				Name:     req.Name,
				Email:    req.Email,
				IsAdmin:  true,
				IsActive: true,
			}, nil
		},
	}

	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--name", "Admin User", "--email", "admin@test.com", "--admin"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_MissingName(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--email", "test@test.com"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddCmd_MissingEmail(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--name", "Test User"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddCmd_Error(t *testing.T) {
	mock := &client.MockClient{
		AddUserFunc: func(req data.AddUserRequest) (*data.User, error) {
			return nil, fmt.Errorf("email already exists")
		},
	}

	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--name", "Test", "--email", "existing@test.com"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already exists")
}
