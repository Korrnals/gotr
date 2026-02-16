package users

import (
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Tests for Register ====================

func TestRegister(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}

	// Mock function for getting client
	mockFn := func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	}

	Register(rootCmd, mockFn)

	// Check that users command was added
	usersCmd, _, err := rootCmd.Find([]string{"users"})
	assert.NoError(t, err)
	assert.NotNil(t, usersCmd)
	assert.Equal(t, "users", usersCmd.Name())

	// Check that all subcommands exist
	subcommands := []string{"list", "get", "get-by-email", "add", "update"}
	for _, sub := range subcommands {
		subCmd, _, err := rootCmd.Find([]string{"users", sub})
		assert.NoError(t, err, "subcommand %s should exist", sub)
		assert.NotNil(t, subCmd, "subcommand %s should not be nil", sub)
	}
}

// ==================== Additional Tests for Update Command ====================

func TestUpdateCmd_EmailFlag(t *testing.T) {
	mock := &client.MockClient{
		UpdateUserFunc: func(userID int64, req data.UpdateUserRequest) (*data.User, error) {
			assert.Equal(t, int64(123), userID)
			assert.Equal(t, "newemail@test.com", req.Email)
			return &data.User{
				ID:    123,
				Name:  "Test User",
				Email: req.Email,
			}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--email", "newemail@test.com"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_RoleFlag(t *testing.T) {
	mock := &client.MockClient{
		UpdateUserFunc: func(userID int64, req data.UpdateUserRequest) (*data.User, error) {
			assert.Equal(t, int64(123), userID)
			assert.Equal(t, int64(5), req.RoleID)
			return &data.User{
				ID:     123,
				Name:   "Test User",
				RoleID: 5,
			}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--role", "5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_MultipleFlags(t *testing.T) {
	mock := &client.MockClient{
		UpdateUserFunc: func(userID int64, req data.UpdateUserRequest) (*data.User, error) {
			assert.Equal(t, int64(123), userID)
			assert.Equal(t, "New Name", req.Name)
			assert.Equal(t, "new@test.com", req.Email)
			assert.Equal(t, int64(3), req.RoleID)
			return &data.User{
				ID:     123,
				Name:   req.Name,
				Email:  req.Email,
				RoleID: req.RoleID,
			}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "New Name", "--email", "new@test.com", "--role", "3"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
