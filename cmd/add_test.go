package cmd

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// setupAddTest настраивает тестовое окружение для add команды
func setupAddTest(t *testing.T, mock *client.MockClient) *cobra.Command {
	// Создаем новую команду для теста
	cmd := &cobra.Command{
		Use:   addCmd.Use,
		Short: addCmd.Short,
		Long:  addCmd.Long,
		RunE:  runAdd,
	}
	
	// Добавляем флаги
	cmd.Flags().StringP("name", "n", "", "Название ресурса")
	cmd.Flags().StringP("description", "d", "", "Описание/announcement")
	cmd.Flags().String("announcement", "", "Announcement (для проекта)")
	cmd.Flags().Bool("show-announcement", false, "Показывать announcement")
	cmd.Flags().Int64("suite-id", 0, "ID сьюта")
	cmd.Flags().Int64("section-id", 0, "ID секции")
	cmd.Flags().Int64("milestone-id", 0, "ID milestone")
	cmd.Flags().Int64("template-id", 0, "ID шаблона (для case)")
	cmd.Flags().Int64("type-id", 0, "ID типа (для case)")
	cmd.Flags().Int64("priority-id", 0, "ID приоритета (для case)")
	cmd.Flags().String("title", "", "Заголовок (для case)")
	cmd.Flags().String("refs", "", "Ссылки (references)")
	cmd.Flags().String("comment", "", "Комментарий (для result)")
	cmd.Flags().Int64("status-id", 0, "ID статуса (для result)")
	cmd.Flags().String("elapsed", "", "Время выполнения (для result)")
	cmd.Flags().String("defects", "", "Дефекты (для result)")
	cmd.Flags().Int64("assignedto-id", 0, "ID назначенного пользователя")
	cmd.Flags().String("case-ids", "", "ID кейсов через запятую (для run)")
	cmd.Flags().Bool("include-all", true, "Включить все кейсы (для run)")
	cmd.Flags().String("json-file", "", "Путь к JSON-файлу с данными")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл")
	
	// Создаем контекст с mock клиентом
	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	cmd.SetContext(ctx)
	
	return cmd
}

// TestAdd_Project_Success проверяет создание проекта
func TestAdd_Project_Success(t *testing.T) {
	mock := &client.MockClient{
		AddProjectFunc: func(req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: 999, Name: req.Name}, nil
		},
	}
	
	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"project", "--name", "Test Project", "--announcement", "Test Announcement"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_Suite_Success проверяет создание сьюта
func TestAdd_Suite_Success(t *testing.T) {
	mock := &client.MockClient{
		AddSuiteFunc: func(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
			assert.Equal(t, int64(1), projectID)
			return &data.Suite{ID: 100, Name: req.Name}, nil
		},
	}
	
	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"suite", "1", "--name", "Test Suite", "--description", "Suite desc"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_Case_Success проверяет создание кейса
func TestAdd_Case_Success(t *testing.T) {
	mock := &client.MockClient{
		AddCaseFunc: func(sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(100), sectionID)
			assert.Equal(t, "Test Case", req.Title)
			return &data.Case{ID: 999, Title: req.Title}, nil
		},
	}
	
	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"case", "100", "--title", "Test Case", "--template-id", "1", "--priority-id", "2"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_Run_Success проверяет создание рана
func TestAdd_Run_Success(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Test Run", req.Name)
			return &data.Run{ID: 999, Name: req.Name}, nil
		},
	}
	
	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"run", "1", "--name", "Test Run", "--suite-id", "100"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_Result_Success проверяет добавление результата
func TestAdd_Result_Success(t *testing.T) {
	mock := &client.MockClient{
		AddResultFunc: func(testID int64, req *data.AddResultRequest) (*data.Result, error) {
			assert.Equal(t, int64(12345), testID)
			assert.Equal(t, int64(1), req.StatusID)
			return &data.Result{ID: 999, StatusID: req.StatusID}, nil
		},
	}
	
	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"result", "12345", "--status-id", "1", "--comment", "Test passed"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_SharedStep_Success проверяет создание shared step
func TestAdd_SharedStep_Success(t *testing.T) {
	mock := &client.MockClient{
		AddSharedStepFunc: func(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
			assert.Equal(t, int64(1), projectID)
			return &data.SharedStep{ID: 999, Title: req.Title}, nil
		},
	}
	
	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"shared-step", "1", "--title", "Test Shared Step"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_NoEndpoint проверяет ошибку при отсутствии endpoint
func TestAdd_NoEndpoint(t *testing.T) {
	mock := &client.MockClient{}
	
	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{})
	
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint")
}

// TestAdd_UnsupportedEndpoint проверяет ошибку при неподдерживаемом endpoint
func TestAdd_UnsupportedEndpoint(t *testing.T) {
	mock := &client.MockClient{}
	
	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"unsupported", "1"})
	
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "неподдерживаемый")
}
