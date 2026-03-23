package result

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// containsStr проверяет содержит ли строка substr
func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ==================== Тесты для saveToFile ====================

func TestSaveToFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_output.json")

	data := map[string]interface{}{
		"id":   123,
		"name": "test",
	}

	err := saveToFile(data, filename)
	assert.NoError(t, err)

	// Проверяем что файл создан
	content, err := os.ReadFile(filename)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "123")
	assert.Contains(t, string(content), "test")
}

func TestSaveToFile_InvalidData(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_output.json")

	// Канал нельзя сериализовать в JSON
	invalidData := make(chan int)

	err := saveToFile(invalidData, filename)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serialization")
}

func TestSaveToFile_InvalidPath(t *testing.T) {
	// Путь к несуществующей директории без прав на создание
	invalidPath := "/nonexistent_dir_xyz/test.json"

	data := map[string]string{"key": "value"}

	err := saveToFile(data, invalidPath)
	assert.Error(t, err)
}

// ==================== Тесты для service_wrapper ====================

func TestResultServiceWrapper_AddResults(t *testing.T) {
	mock := &client.MockClient{
		AddResultsFunc: func(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Len(t, req.Results, 2)
			return data.GetResultsResponse{
				{ID: 1, TestID: 101, StatusID: 1},
				{ID: 2, TestID: 102, StatusID: 5},
			}, nil
		},
	}

	wrapper := &resultServiceWrapper{svc: nil}
	// Проверяем что wrapper реализует интерфейс
	var _ ResultServiceInterface = wrapper

	// Создадим сервис через конструктор
	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	results, err := svc.AddResults(ctx, 12345, &data.AddResultsRequest{
		Results: []data.ResultEntry{
			{TestID: 101, StatusID: 1},
			{TestID: 102, StatusID: 5},
		},
	})

	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestResultServiceWrapper_AddResults_Error(t *testing.T) {
	mock := &client.MockClient{
		AddResultsFunc: func(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	results, err := svc.AddResults(ctx, 99999, &data.AddResultsRequest{
		Results: []data.ResultEntry{{TestID: 101, StatusID: 1}},
	})

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestResultServiceWrapper_AddResultsForCases(t *testing.T) {
	mock := &client.MockClient{
		AddResultsForCasesFunc: func(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Len(t, req.Results, 2)
			return data.GetResultsResponse{
				{ID: 1, TestID: 201, StatusID: 1},
				{ID: 2, TestID: 202, StatusID: 1},
			}, nil
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	results, err := svc.AddResultsForCases(ctx, 12345, &data.AddResultsForCasesRequest{
		Results: []data.ResultForCaseEntry{
			{CaseID: 301, StatusID: 1},
			{CaseID: 302, StatusID: 1},
		},
	})

	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestResultServiceWrapper_AddResultsForCases_Error(t *testing.T) {
	mock := &client.MockClient{
		AddResultsForCasesFunc: func(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("invalid case_id")
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	results, err := svc.AddResultsForCases(ctx, 12345, &data.AddResultsForCasesRequest{
		Results: []data.ResultForCaseEntry{{CaseID: 301, StatusID: 1}},
	})

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestResultServiceWrapper_GetRunsForProject(t *testing.T) {
	mock := &client.MockClient{
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{
				{ID: 101, Name: "Run 1", ProjectID: 1},
				{ID: 102, Name: "Run 2", ProjectID: 1},
			}, nil
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	runs, err := svc.GetRunsForProject(ctx, 1)

	assert.NoError(t, err)
	assert.Len(t, runs, 2)
	assert.Equal(t, int64(101), runs[0].ID)
}

func TestResultServiceWrapper_GetRunsForProject_Error(t *testing.T) {
	mock := &client.MockClient{
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	runs, err := svc.GetRunsForProject(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, runs)
}

// ==================== Тесты для newResultServiceFromInterface ====================

func TestNewResultServiceFromInterface_WithHTTPClient(t *testing.T) {
	// Создаем mock HTTPClient
	mock := &client.MockClient{}

	// Передаем как ClientInterface
	svc := newResultServiceFromInterface(mock)
	assert.NotNil(t, svc)
}

// ==================== Тесты для SetGetClientForTests и getClientSafe ====================

func TestSetGetClientForTests(t *testing.T) {
	// Сохраняем текущее состояние
	oldAccessor := clientAccessor
	defer func() {
		clientAccessor = oldAccessor
	}()

	// Сбрасываем accessor
	clientAccessor = nil

	// Устанавливаем тестовую функцию
	mockFn := func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	}

	SetGetClientForTests(mockFn)
	assert.NotNil(t, clientAccessor)

	// Повторный вызов должен обновить функцию
	SetGetClientForTests(mockFn)
	assert.NotNil(t, clientAccessor)
}

func TestGetClientSafe_WithNilAccessor(t *testing.T) {
	// Сохраняем текущее состояние
	oldAccessor := clientAccessor
	defer func() {
		clientAccessor = oldAccessor
	}()

	// Сбрасываем accessor
	clientAccessor = nil

	// Должен вернуть nil когда accessor nil
	cmd := &cobra.Command{}
	cli := getClientSafe(cmd)
	assert.Nil(t, cli)
}

func TestGetClientSafe_WithAccessor(t *testing.T) {
	// Сохраняем текущее состояние
	oldAccessor := clientAccessor
	defer func() {
		clientAccessor = oldAccessor
	}()

	// Создаем accessor с тестовой функцией
	mockFn := func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	}
	clientAccessor = client.NewAccessor(mockFn)

	// Должен вернуть nil (так как mockFn возвращает nil)
	cmd := &cobra.Command{}
	cli := getClientSafe(cmd)
	assert.Nil(t, cli)
}

// ==================== Тесты для Register ====================

func TestRegister(t *testing.T) {
	// Сохраняем текущее состояние
	oldAccessor := clientAccessor
	defer func() {
		clientAccessor = oldAccessor
	}()

	// Сбрасываем accessor
	clientAccessor = nil

	// Создаем корневую команду
	rootCmd := &cobra.Command{Use: "gotr"}

	// Mock функция получения clientа
	mockFn := func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	}

	// Регистрируем result команду
	Register(rootCmd, mockFn)

	// Проверяем что команда добавлена
	assert.NotNil(t, clientAccessor)

	// Проверяем что result команда есть в root
	resultCmd, _, err := rootCmd.Find([]string{"result"})
	assert.NoError(t, err)
	assert.NotNil(t, resultCmd)

	// Проверяем что подкоманды добавлены
	subcommands := []string{"list", "get", "get-case", "add", "add-case", "add-bulk", "fields"}
	for _, sub := range subcommands {
		cmd, _, err := rootCmd.Find([]string{"result", sub})
		assert.NoError(t, err, "subcommand %s should exist", sub)
		assert.NotNil(t, cmd, "subcommand %s should not be nil", sub)

		// Проверяем что флаги save и quiet добавлены
		saveFlag := cmd.Flags().Lookup("save")
		assert.NotNil(t, saveFlag, "save flag should exist on %s", sub)

		quietFlag := cmd.Flags().Lookup("quiet")
		assert.NotNil(t, quietFlag, "quiet flag should exist on %s", sub)
	}
}

// ==================== Тесты для list command (интерактивный режим) ====================

func TestListCmd_Interactive_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{
				{ID: 12345, Name: "Test Run", ProjectID: 1},
			}, nil
		},
		GetResultsForRunFunc: func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), runID)
			return []data.Result{{ID: 1, TestID: 100, StatusID: 1}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}). // выбор проекта
		WithSelectResponses(interactive.SelectResponse{Index: 0})  // выбор run

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	// Без аргументов - должен включиться интерактивный режим
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_Interactive_SelectProjectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
	}

	// Пустой MockPrompter — очередь исчерпана, SelectProject вернёт ошибку
	p := interactive.NewMockPrompter()

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_Interactive_GetRunsError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return nil, fmt.Errorf("failed to get runs")
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}) // выбор проекта

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "runs")
}

func TestListCmd_Interactive_EmptyRuns(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}) // выбор проекта

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no test runs found")
}

func TestListCmd_Interactive_SelectRunError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{
				{ID: 12345, Name: "Test Run", ProjectID: 1},
			}, nil
		},
	}

	// MockPrompter: 1 SelectResponse для проекта, ничего для run — очередь исчерпана
	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты для outputResult (через команду с флагом output) ====================

func TestOutputResult_WithSaveFlag(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForRunFunc: func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
			return []data.Result{{ID: 1, TestID: 100, StatusID: 1}}, nil
		},
	}

	// Пересоздаем команду с нашим getClient чтобы использовать mock
	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	// Добавляем флаг save (как это делает Register)
	output.AddFlag(cmd)
	cmd.SetArgs([]string{"12345", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Дополнительные тесты для покрытия ====================

func TestAddBulkResults_ParseError(t *testing.T) {
	// Тест для покрытия ветки ошибки парсинга JSON в AddBulkResults
	mock := &client.MockClient{}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)

	// Передаем invalid JSON который не парсится ни в один формат
	invalidJSON := []byte(`{"invalid": "json"}`)

	result, err := svc.AddBulkResults(ctx, 12345, invalidJSON)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestPrintJSON_Error(t *testing.T) {
	// Тестируем ошибку в printJSON с несериализуемыми данными
	invalidData := make(chan int) // Канал нельзя сериализовать

	err := printJSON(invalidData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serialization")
}

// ==================== Дополнительные тесты для service_wrapper ====================

func TestResultServiceWrapper_AddBulkResults_EmptyArray(t *testing.T) {
	// Тест для покрытия ветки с empty массивом в JSON
	mock := &client.MockClient{}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)

	// Пустой массив
	emptyJSON := []byte(`[]`)

	result, err := svc.AddBulkResults(ctx, 12345, emptyJSON)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestResultServiceWrapper_AddBulkResults_InvalidJSON(t *testing.T) {
	// Тест для покрытия ветки с невалидным JSON
	mock := &client.MockClient{}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)

	// Невалидный JSON
	invalidJSON := []byte(`{invalid json`)

	result, err := svc.AddBulkResults(ctx, 12345, invalidJSON)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestResultServiceWrapper_AllMethods(t *testing.T) {
	mock := &client.MockClient{
		GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
			return []data.Result{{ID: 1, TestID: testID}}, nil
		},
		GetResultsForCaseFunc: func(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
			return []data.Result{{ID: 1, TestID: 100}}, nil
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)

	// Test GetForTest
	results, err := svc.GetForTest(ctx, 123)
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	// Test GetForCase
	results, err = svc.GetForCase(ctx, 1, 100)
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	// Test GetForRun
	mock.GetResultsForRunFunc = func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
		return []data.Result{{ID: 1, TestID: 200}}, nil
	}
	results, err = svc.GetForRun(ctx, 456)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestResultServiceWrapper_ParseID(t *testing.T) {
	mock := &client.MockClient{}
	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)

	// Test valid ID
	id, err := svc.ParseID(ctx, []string{"123"}, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)

	// Test invalid ID
	_, err = svc.ParseID(ctx, []string{"abc"}, 0)
	assert.Error(t, err)
}
