package roles

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Functional tests with mock ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetRolesFunc: func(ctx context.Context) (data.GetRolesResponse, error) {
			return []data.Role{
				{ID: 1, Name: "Administrator"},
				{ID: 2, Name: "Tester"},
				{ID: 3, Name: "Guest"},
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
		GetRolesFunc: func(ctx context.Context) (data.GetRolesResponse, error) {
			return []data.Role{}, nil
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
		GetRolesFunc: func(ctx context.Context) (data.GetRolesResponse, error) {
			return nil, fmt.Errorf("API connection error")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection error")
}

func TestListCmd_WithSaveFlag(t *testing.T) {
	mock := &client.MockClient{
		GetRolesFunc: func(ctx context.Context) (data.GetRolesResponse, error) {
			return []data.Role{
				{ID: 1, Name: "Administrator"},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Validation tests ====================

func TestListCmd_ExtraArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"extra"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Helper function tests ====================

func TestGetClientForTests_NilCmd(t *testing.T) {
	result := testhelper.GetClientForTests(nil)
	assert.Nil(t, result)
}

func TestGetClientForTests_NilContext(t *testing.T) {
	cmd := &cobra.Command{}
	result := testhelper.GetClientForTests(cmd)
	assert.Nil(t, result)
}

func TestGetClientForTests_NoMockInContext(t *testing.T) {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testhelper.ContextKey("other_key"), "value")
	cmd.SetContext(ctx)

	result := testhelper.GetClientForTests(cmd)
	assert.Nil(t, result)
}

// ==================== outputResult tests ====================

func TestOutputResult_JSONError(t *testing.T) {
	badData := make(chan int)

	cmd := &cobra.Command{}
	cmd.Flags().String("save", "", "")

	err := output.OutputResult(cmd, badData, "roles")
	assert.Error(t, err)
}

// ==================== Registration tests ====================

func TestRegister(t *testing.T) {
	root := &cobra.Command{}
	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	})

	// Verify the command is added
	rolesCmd, _, err := root.Find([]string{"roles"})
	assert.NoError(t, err)
	assert.NotNil(t, rolesCmd)
	assert.Equal(t, "roles", rolesCmd.Name())

	// Verify list subcommand exists
	listCmd, _, err := root.Find([]string{"roles", "list"})
	assert.NoError(t, err)
	assert.NotNil(t, listCmd)

	// Verify get subcommand exists
	getCmd, _, err := root.Find([]string{"roles", "get"})
	assert.NoError(t, err)
	assert.NotNil(t, getCmd)
}
