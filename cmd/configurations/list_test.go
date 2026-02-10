package configurations

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
		GetConfigsFunc: func(projectID int64) (data.GetConfigsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetConfigsResponse{
				{
					ID:   1,
					Name: "Browsers",
					Configs: []data.Config{
						{ID: 10, Name: "Chrome"},
						{ID: 11, Name: "Firefox"},
					},
				},
				{
					ID:   2,
					Name: "OS",
					Configs: []data.Config{
						{ID: 20, Name: "Windows"},
						{ID: 21, Name: "Linux"},
					},
				},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_WithOutputFile(t *testing.T) {
	mock := &client.MockClient{
		GetConfigsFunc: func(projectID int64) (data.GetConfigsResponse, error) {
			return data.GetConfigsResponse{
				{ID: 1, Name: "Browsers"},
			}, nil
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "configs.json")

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "-o", outputFile})

	err := cmd.Execute()
	assert.NoError(t, err)

	// Проверяем что файл создан
	_, err = os.Stat(outputFile)
	assert.NoError(t, err)

	// Проверяем содержимое
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Browsers")
}

func TestListCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetConfigsFunc: func(projectID int64) (data.GetConfigsResponse, error) {
			return data.GetConfigsResponse{}, nil
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
		GetConfigsFunc: func(projectID int64) (data.GetConfigsResponse, error) {
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

// ==================== Тесты регистрации ====================

func TestRegister(t *testing.T) {
	root := &cobra.Command{}
	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	})

	// Проверяем что команда добавлена
	configsCmd, _, err := root.Find([]string{"configurations"})
	assert.NoError(t, err)
	assert.NotNil(t, configsCmd)
	assert.Equal(t, "configurations", configsCmd.Name())

	// Проверяем что подкоманда list существует
	listCmd, _, err := root.Find([]string{"configurations", "list"})
	assert.NoError(t, err)
	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Name())
}

// ==================== Тесты outputResult ====================

func TestOutputResult_JSONError(t *testing.T) {
	// Создаём канал, который нельзя сериализовать в JSON
	badData := make(chan int)

	cmd := &cobra.Command{}
	cmd.SetArgs([]string{})
	cmd.Flags().StringP("output", "o", "", "")

	err := outputResult(cmd, badData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported type")
}

// ==================== Тесты getClientForTests ====================

func TestGetClientForTests_NilCmd(t *testing.T) {
	result := getClientForTests(nil)
	assert.Nil(t, result)
}

func TestGetClientForTests_NilContext(t *testing.T) {
	cmd := &cobra.Command{}
	// Не устанавливаем контекст — он будет nil
	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

func TestGetClientForTests_NoMockInContext(t *testing.T) {
	cmd := &cobra.Command{}
	// Устанавливаем контекст без мока
	ctx := context.WithValue(context.Background(), "other_key", "value")
	cmd.SetContext(ctx)

	result := getClientForTests(cmd)
	assert.Nil(t, result)
}
