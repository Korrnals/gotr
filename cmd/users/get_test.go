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

// ==================== Функциональные тесты с моком ====================

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetUserFunc: func(ctx context.Context, userID int64) (*data.User, error) {
			assert.Equal(t, int64(12345), userID)
			return &data.User{
				ID:       12345,
				Name:     "John Doe",
				Email:    "john@example.com",
				IsActive: true,
				RoleID:   1,
				Role:     "Lead",
				IsAdmin:  false,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NotFound(t *testing.T) {
	mock := &client.MockClient{
		GetUserFunc: func(ctx context.Context, userID int64) (*data.User, error) {
			return nil, fmt.Errorf("user not found")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// ==================== Тесты валидации ====================

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) {
			return data.GetUsersResponse{{ID: 42, Name: "John Doe", Email: "john@example.com"}}, nil
		},
		GetUserFunc: func(ctx context.Context, userID int64) (*data.User, error) {
			assert.Equal(t, int64(42), userID)
			return &data.User{ID: 42, Name: "John Doe", Email: "john@example.com"}, nil
		},
	}
	cmd := newGetCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NoArgs_NonInteractive(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetCmd_NoArgs_NoPrompter(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetCmd_NoArgs_InteractiveResolveError(t *testing.T) {
	mock := &client.MockClient{
		GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) {
			return nil, fmt.Errorf("users boom")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewMockPrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "users boom")
}
