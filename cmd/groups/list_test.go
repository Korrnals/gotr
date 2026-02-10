package groups

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetGroupsFunc: func(projectID int64) (data.GetGroupsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Group{
				{ID: 1, Name: "QA Team"},
				{ID: 2, Name: "Developers"},
				{ID: 3, Name: "Managers"},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetGroupsFunc: func(projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		GetGroupsFunc: func(projectID int64) (data.GetGroupsResponse, error) {
			return nil, fmt.Errorf("ошибка подключения к API")
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка подключения")
}

func TestListCmd_WithOutputFile(t *testing.T) {
	mock := &client.MockClient{
		GetGroupsFunc: func(projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{
				{ID: 1, Name: "QA Team"},
			}, nil
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "groups.json")

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "-o", outputFile})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "QA Team")
}

// ==================== Тесты валидации ====================

func TestListCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный project_id")
}

func TestListCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный project_id")
}

func TestListCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты вспомогательных функций ====================

func TestGetClientForTests_NilCmd(t *testing.T) {
	result := getClientForTests(nil)
	assert.Nil(t, result)
}

func TestGetClientForTests_NilContext(t *testing.T) {
	cmd := &cobra.Command{}
	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

func TestGetClientForTests_NoMockInContext(t *testing.T) {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), "other_key", "value")
	cmd.SetContext(ctx)

	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

// ==================== Тесты outputResult ====================

func TestOutputResult_JSONError(t *testing.T) {
	badData := make(chan int)

	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")

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
	groupsCmd, _, err := root.Find([]string{"groups"})
	assert.NoError(t, err)
	assert.NotNil(t, groupsCmd)
	assert.Equal(t, "groups", groupsCmd.Name())

	// Проверяем что подкоманда list существует
	listCmd, _, err := root.Find([]string{"groups", "list"})
	assert.NoError(t, err)
	assert.NotNil(t, listCmd)

	// Проверяем что подкоманда get существует
	getCmd, _, err := root.Find([]string{"groups", "get"})
	assert.NoError(t, err)
	assert.NotNil(t, getCmd)
}
