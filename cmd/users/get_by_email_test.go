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

// ==================== Functional tests with mock ====================

func TestGetByEmailCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*data.User, error) {
			assert.Equal(t, "john@example.com", email)
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

	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"john@example.com"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetByEmailCmd_NotFound(t *testing.T) {
	mock := &client.MockClient{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*data.User, error) {
			return nil, fmt.Errorf("user not found")
		},
	}

	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"notfound@example.com"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestGetByEmailCmd_InvalidEmail(t *testing.T) {
	mock := &client.MockClient{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*data.User, error) {
			return nil, fmt.Errorf("invalid email format")
		},
	}

	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid-email"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Validation tests ====================

func TestGetByEmailCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) {
			return data.GetUsersResponse{{ID: 7, Name: "John Doe", Email: "john@example.com"}}, nil
		},
		GetUserByEmailFunc: func(ctx context.Context, email string) (*data.User, error) {
			assert.Equal(t, "john@example.com", email)
			return &data.User{ID: 7, Name: "John Doe", Email: email}, nil
		},
	}
	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetByEmailCmd_NoArgs_NonInteractive(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetByEmailCmd_TooManyArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"one", "two"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetByEmailCmd_NoArgs_NoPrompter(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetByEmailCmd_InteractiveEmptyEmail(t *testing.T) {
	mock := &client.MockClient{
		GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) {
			return data.GetUsersResponse{{ID: 7, Name: "John Doe", Email: ""}}, nil
		},
	}

	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be empty")
}

func TestGetByEmailCmd_NoArgs_InteractiveResolveError(t *testing.T) {
	mock := &client.MockClient{
		GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) {
			return nil, fmt.Errorf("users boom")
		},
	}

	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewMockPrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "users boom")
}
