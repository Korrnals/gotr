package datasets

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
		GetDatasetsFunc: func(projectID int64) (data.GetDatasetsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Dataset{
				{ID: 101, Name: "Login Data", ProjectID: 1},
				{ID: 102, Name: "Search Data", ProjectID: 1},
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
		GetDatasetsFunc: func(projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
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
		GetDatasetsFunc: func(projectID int64) (data.GetDatasetsResponse, error) {
			return nil, fmt.Errorf("проект не найден")
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "проект не найден")
}

func TestListCmd_WithOutputFile(t *testing.T) {
	mock := &client.MockClient{
		GetDatasetsFunc: func(projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{{ID: 1, Name: "Test Data"}}, nil
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "datasets.json")

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "-o", outputFile})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test Data")
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

	datasetsCmd, _, err := root.Find([]string{"datasets"})
	assert.NoError(t, err)
	assert.NotNil(t, datasetsCmd)
	assert.Equal(t, "datasets", datasetsCmd.Name())

	// Проверяем все подкоманды
	listCmd, _, _ := root.Find([]string{"datasets", "list"})
	assert.NotNil(t, listCmd)

	getCmd, _, _ := root.Find([]string{"datasets", "get"})
	assert.NotNil(t, getCmd)

	addCmd, _, _ := root.Find([]string{"datasets", "add"})
	assert.NotNil(t, addCmd)

	updateCmd, _, _ := root.Find([]string{"datasets", "update"})
	assert.NotNil(t, updateCmd)

	deleteCmd, _, _ := root.Find([]string{"datasets", "delete"})
	assert.NotNil(t, deleteCmd)
}
