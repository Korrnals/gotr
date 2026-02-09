package cmd

import (
	"context"
	"os"
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
	cmd.Flags().String("description", "", "Описание/announcement")
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


// ==================== Attachment Tests ====================

// TestAdd_AttachmentCase_Success проверяет добавление вложения к кейсу
func TestAdd_AttachmentCase_Success(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test-attachment-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("test content")
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToCaseFunc: func(caseID int64, filePath string) (*data.AttachmentResponse, error) {
			assert.Equal(t, int64(12345), caseID)
			assert.Equal(t, tmpFile.Name(), filePath)
			return &data.AttachmentResponse{AttachmentID: 999, URL: "https://example.com/attachment/999"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "case", "12345", tmpFile.Name()})

	err = cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_AttachmentPlan_Success проверяет добавление вложения к плану
func TestAdd_AttachmentPlan_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-plan-*.pdf")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("plan content")
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToPlanFunc: func(planID int64, filePath string) (*data.AttachmentResponse, error) {
			assert.Equal(t, int64(100), planID)
			assert.Equal(t, tmpFile.Name(), filePath)
			return &data.AttachmentResponse{AttachmentID: 888, URL: "https://example.com/attachment/888"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "plan", "100", tmpFile.Name()})

	err = cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_AttachmentPlanEntry_Success проверяет добавление вложения к plan entry
func TestAdd_AttachmentPlanEntry_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-entry-*.doc")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("entry content")
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToPlanEntryFunc: func(planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
			assert.Equal(t, int64(200), planID)
			assert.Equal(t, "entry-abc123", entryID)
			assert.Equal(t, tmpFile.Name(), filePath)
			return &data.AttachmentResponse{AttachmentID: 777, URL: "https://example.com/attachment/777"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "plan-entry", "200", "entry-abc123", tmpFile.Name()})

	err = cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_AttachmentResult_Success проверяет добавление вложения к результату
func TestAdd_AttachmentResult_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-result-*.log")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("result log content")
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToResultFunc: func(resultID int64, filePath string) (*data.AttachmentResponse, error) {
			assert.Equal(t, int64(98765), resultID)
			assert.Equal(t, tmpFile.Name(), filePath)
			return &data.AttachmentResponse{AttachmentID: 666, URL: "https://example.com/attachment/666"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "result", "98765", tmpFile.Name()})

	err = cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_AttachmentRun_Success проверяет добавление вложения к рану
func TestAdd_AttachmentRun_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-run-*.png")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("png binary content")
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToRunFunc: func(runID int64, filePath string) (*data.AttachmentResponse, error) {
			assert.Equal(t, int64(555), runID)
			assert.Equal(t, tmpFile.Name(), filePath)
			return &data.AttachmentResponse{AttachmentID: 555, URL: "https://example.com/attachment/555"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "run", "555", tmpFile.Name()})

	err = cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_Attachment_MissingArgs проверяет ошибку при недостаточных аргументах
func TestAdd_Attachment_MissingArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "тип вложения")
}

// TestAdd_Attachment_InvalidCaseID проверяет ошибку при неверном case_id
func TestAdd_Attachment_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "case", "invalid", "/tmp/test.txt"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case_id")
}

// TestAdd_Attachment_UnsupportedType проверяет ошибку при неподдерживаемом типе
func TestAdd_Attachment_UnsupportedType(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "invalid-type", "123", "/tmp/test.txt"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "неподдерживаемый тип")
}

// TestAdd_Attachment_MissingFilePath проверяет ошибку при отсутствии пути к файлу
func TestAdd_Attachment_MissingFilePath(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "case", "12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "использование")
}

// TestAdd_Attachment_MissingPlanEntryArgs проверяет ошибку при недостаточных аргументах для plan-entry
func TestAdd_Attachment_MissingPlanEntryArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "plan-entry", "100", "entry-id"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "использование")
}
