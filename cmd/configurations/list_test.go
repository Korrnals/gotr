package configurations

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
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

func TestListCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return data.GetConfigsResponse{
				{ID: 1, Name: "Browsers"},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
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
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

// ==================== Тесты валидации ====================

func TestListCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestListCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestListCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetConfigsResponse{{ID: 1, Name: "Browsers"}}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
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
	cmd.Flags().String("save", "", "")

	err := output.OutputResult(cmd, badData, "configurations")
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
