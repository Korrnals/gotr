package users

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateUserFunc: func(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
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
		UpdateUserFunc: func(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
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
		UpdateUserFunc: func(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
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

func TestUpdateCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) {
			return data.GetUsersResponse{{ID: 123, Name: "Test User", Email: "user@test.com"}}, nil
		},
		UpdateUserFunc: func(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
			assert.Equal(t, int64(123), userID)
			assert.Equal(t, "Updated Name", req.Name)
			return &data.User{ID: 123, Name: req.Name, Email: "user@test.com"}, nil
		},
	}
	cmd := newUpdateCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})))
	cmd.SetArgs([]string{"--name", "Updated Name"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_NoArgs_NonInteractive(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestUpdateCmd_Error(t *testing.T) {
	mock := &client.MockClient{
		UpdateUserFunc: func(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
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

func TestUpdateCmd_DryRun_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		UpdateUserFunc: func(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
			called = true
			return &data.User{ID: userID, Name: req.Name}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--name", "Updated Name", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.False(t, called)
}

func TestUpdateCmd_NoArgs_NoPrompter(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestUpdateCmd_ResetAdminAndInactiveFalse(t *testing.T) {
	mock := &client.MockClient{
		UpdateUserFunc: func(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
			assert.Equal(t, int64(123), userID)
			assert.Equal(t, 0, req.IsAdmin)
			assert.Equal(t, 1, req.IsActive)
			return &data.User{ID: 123, Name: "Test User", IsAdmin: false, IsActive: true}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "--admin=false", "--inactive=false"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_NoArgs_InteractiveResolveError(t *testing.T) {
	mock := &client.MockClient{
		GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) {
			return nil, fmt.Errorf("users boom")
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewMockPrompter()))
	cmd.SetArgs([]string{"--name", "Updated Name"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "users boom")
}
