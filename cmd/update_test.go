package cmd

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// setupUpdateTest настраивает тестовое окружение для update команды
func setupUpdateTest(t *testing.T, mock *client.MockClient) *cobra.Command {
	// Создаем новую команду для теста
	cmd := &cobra.Command{
		Use:   updateCmd.Use,
		Short: updateCmd.Short,
		Long:  updateCmd.Long,
		RunE:  runUpdate,
	}
	
	// Добавляем флаги
	cmd.Flags().StringP("name", "n", "", "Название ресурса")
	cmd.Flags().String("description", "", "Описание")
	cmd.Flags().String("announcement", "", "Announcement (для проекта)")
	cmd.Flags().Bool("show-announcement", false, "Показывать announcement")
	cmd.Flags().Bool("is-completed", false, "Отметить как завершённый")
	cmd.Flags().String("title", "", "Заголовок (для case)")
	cmd.Flags().Int64("type-id", 0, "ID типа (для case)")
	cmd.Flags().Int64("priority-id", 0, "ID приоритета (для case)")
	cmd.Flags().String("refs", "", "Ссылки (references)")
	cmd.Flags().Int64("suite-id", 0, "ID сьюта")
	cmd.Flags().Int64("milestone-id", 0, "ID milestone")
	cmd.Flags().Int64("assignedto-id", 0, "ID назначенного пользователя")
	cmd.Flags().String("case-ids", "", "ID кейсов через запятую (для run)")
	cmd.Flags().Bool("include-all", false, "Включить все кейсы (для run)")
	cmd.Flags().String("json-file", "", "Путь к JSON-файлу с данными")
	save.AddFlag(cmd)
	
	// Создаем контекст с mock клиентом
	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	cmd.SetContext(ctx)
	
	return cmd
}

// TestUpdate_Project_Success проверяет обновление проекта
func TestUpdate_Project_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateProjectFunc: func(projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Updated Project", req.Name)
			return &data.GetProjectResponse{ID: projectID, Name: req.Name}, nil
		},
	}
	
	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"project", "1", "--name", "Updated Project"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_Suite_Success проверяет обновление сьюта
func TestUpdate_Suite_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateSuiteFunc: func(suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
			assert.Equal(t, int64(100), suiteID)
			assert.Equal(t, "Updated Suite", req.Name)
			return &data.Suite{ID: suiteID, Name: req.Name}, nil
		},
	}
	
	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"suite", "100", "--name", "Updated Suite"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_Section_Success проверяет обновление секции
func TestUpdate_Section_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateSectionFunc: func(sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
			assert.Equal(t, int64(200), sectionID)
			assert.Equal(t, "Updated Section", req.Name)
			return &data.Section{ID: sectionID, Name: req.Name}, nil
		},
	}
	
	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"section", "200", "--name", "Updated Section"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_Case_Success проверяет обновление кейса
func TestUpdate_Case_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateCaseFunc: func(caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(12345), caseID)
			// Title передается как указатель
			assert.NotNil(t, req.Title)
			assert.Equal(t, "Updated Case", *req.Title)
			return &data.Case{ID: caseID, Title: *req.Title}, nil
		},
	}
	
	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"case", "12345", "--title", "Updated Case", "--priority-id", "1"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_Run_Success проверяет обновление рана
func TestUpdate_Run_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(1000), runID)
			// Name передается как указатель
			assert.NotNil(t, req.Name)
			assert.Equal(t, "Updated Run", *req.Name)
			return &data.Run{ID: runID, Name: *req.Name}, nil
		},
	}
	
	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"run", "1000", "--name", "Updated Run"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_SharedStep_Success проверяет обновление shared step
func TestUpdate_SharedStep_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateSharedStepFunc: func(stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
			assert.Equal(t, int64(50), stepID)
			assert.Equal(t, "Updated Step", req.Title)
			return &data.SharedStep{ID: stepID, Title: req.Title}, nil
		},
	}
	
	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"shared-step", "50", "--title", "Updated Step"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_NoArgs проверяет ошибку при отсутствии аргументов
func TestUpdate_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	
	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{})
	
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint и id")
}

// TestUpdate_InvalidID проверяет ошибку при неверном ID
func TestUpdate_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	
	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"project", "invalid"})
	
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID")
}

// TestUpdate_UnsupportedEndpoint проверяет ошибку при неподдерживаемом endpoint
func TestUpdate_UnsupportedEndpoint(t *testing.T) {
	mock := &client.MockClient{}
	
	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"unsupported", "1"})
	
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "неподдерживаемый")
}
