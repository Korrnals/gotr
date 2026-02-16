package roles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetRolesFunc: func() (data.GetRolesResponse, error) {
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
		GetRolesFunc: func() (data.GetRolesResponse, error) {
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
		GetRolesFunc: func() (data.GetRolesResponse, error) {
			return nil, fmt.Errorf("ошибка подключения к API")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка подключения")
}

func TestListCmd_WithOutputFile(t *testing.T) {
	mock := &client.MockClient{
		GetRolesFunc: func() (data.GetRolesResponse, error) {
			return []data.Role{
				{ID: 1, Name: "Administrator"},
			}, nil
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "roles.json")

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"-o", outputFile})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Administrator")
}

// ==================== Тесты валидации ====================

func TestListCmd_ExtraArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"extra"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты вспомогательных функций ====================

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
	ctx := context.WithValue(context.Background(), "other_key", "value")
	cmd.SetContext(ctx)

	result := testhelper.GetClientForTests(cmd)
	assert.Nil(t, result)
}

// ==================== Тесты outputResult ====================

func TestOutputResult_JSONError(t *testing.T) {
	badData := make(chan int)

	cmd := &cobra.Command{}
	cmd.Flags().String("save", "", "")

	err := outputResult(cmd, badData)
	assert.Error(t, err)
}

// ==================== Тесты регистрации ====================

func TestRegister(t *testing.T) {
	root := &cobra.Command{}
	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	})

	// Проверяем что команда добавлена
	rolesCmd, _, err := root.Find([]string{"roles"})
	assert.NoError(t, err)
	assert.NotNil(t, rolesCmd)
	assert.Equal(t, "roles", rolesCmd.Name())

	// Проверяем что подкоманда list существует
	listCmd, _, err := root.Find([]string{"roles", "list"})
	assert.NoError(t, err)
	assert.NotNil(t, listCmd)

	// Проверяем что подкоманда get существует
	getCmd, _, err := root.Find([]string{"roles", "get"})
	assert.NoError(t, err)
	assert.NotNil(t, getCmd)
}
