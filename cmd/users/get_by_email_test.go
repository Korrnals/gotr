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

func TestGetByEmailCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetUserByEmailFunc: func(email string) (*data.User, error) {
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
		GetUserByEmailFunc: func(email string) (*data.User, error) {
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
		GetUserByEmailFunc: func(email string) (*data.User, error) {
			return nil, fmt.Errorf("invalid email format")
		},
	}

	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid-email"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты валидации ====================

func TestGetByEmailCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetByEmailCmd_TooManyArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetByEmailCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"one", "two"})

	err := cmd.Execute()
	assert.Error(t, err)
}
