package roles

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

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetRoleFunc: func(ctx context.Context, roleID int64) (*data.Role, error) {
			assert.Equal(t, int64(1), roleID)
			return &data.Role{
				ID:   1,
				Name: "Administrator",
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_WithSaveFlag(t *testing.T) {
	mock := &client.MockClient{
		GetRoleFunc: func(ctx context.Context, roleID int64) (*data.Role, error) {
			return &data.Role{ID: 2, Name: "Tester"}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"2", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NotFound(t *testing.T) {
	mock := &client.MockClient{
		GetRoleFunc: func(ctx context.Context, roleID int64) (*data.Role, error) {
			return nil, fmt.Errorf("role not found")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
}

// ==================== Validation tests ====================

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role_id")
}

func TestGetCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role_id")
}

func TestGetCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetRolesFunc: func(ctx context.Context) (data.GetRolesResponse, error) {
			return data.GetRolesResponse{{ID: 2, Name: "Tester"}}, nil
		},
		GetRoleFunc: func(ctx context.Context, roleID int64) (*data.Role, error) {
			assert.Equal(t, int64(2), roleID)
			return &data.Role{ID: 2, Name: "Tester"}, nil
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

func TestGetCmd_ResolveInteractiveError(t *testing.T) {
	mock := &client.MockClient{
		GetRolesFunc: func(ctx context.Context) (data.GetRolesResponse, error) {
			return nil, fmt.Errorf("roles boom")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	base := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(interactive.WithPrompter(base.Context(), interactive.NewMockPrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
