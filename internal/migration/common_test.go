package migration

import (
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/utils"
	"os"
	"path/filepath"
	"testing"
)

// MockClient — mock для ClientInterface
type MockClient struct {
	AddSharedStepFunc func(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error)
	AddSuiteFunc      func(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error)
	AddCaseFunc       func(suiteID int64, req *data.AddCaseRequest) (*data.Case, error)
	AddSectionFunc    func(projectID int64, req *data.AddSectionRequest) (*data.Section, error)

	GetSharedStepsFunc func(projectID int64) (data.GetSharedStepsResponse, error)
	GetSuitesFunc      func(projectID int64) (data.GetSuitesResponse, error)
	GetCasesFunc       func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error)
	GetSectionsFunc    func(projectID, suiteID int64) (data.GetSectionsResponse, error)
}

// Реализация интерфейса (заглушки + func fields)
func (m *MockClient) AddSharedStep(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
	if m.AddSharedStepFunc != nil {
		return m.AddSharedStepFunc(projectID, req)
	}
	return &data.SharedStep{ID: 999}, nil
}

func (m *MockClient) AddSuite(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
	if m.AddSuiteFunc != nil {
		return m.AddSuiteFunc(projectID, req)
	}
	return &data.Suite{ID: 999}, nil
}

func (m *MockClient) AddSection(projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
	if m.AddSectionFunc != nil {
		return m.AddSectionFunc(projectID, req)
	}
	return &data.Section{ID: 999}, nil
}

func (m *MockClient) AddCase(sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
	if m.AddCaseFunc != nil {
		return m.AddCaseFunc(sectionID, req)
	}
	return &data.Case{ID: 999}, nil
}

func (m *MockClient) GetSharedSteps(projectID int64) (data.GetSharedStepsResponse, error) {
	if m.GetSharedStepsFunc != nil {
		return m.GetSharedStepsFunc(projectID)
	}
	return data.GetSharedStepsResponse{}, nil
}

func (m *MockClient) GetSuites(projectID int64) (data.GetSuitesResponse, error) {
	if m.GetSuitesFunc != nil {
		return m.GetSuitesFunc(projectID)
	}
	return data.GetSuitesResponse{}, nil
}

func (m *MockClient) GetCases(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
	if m.GetCasesFunc != nil {
		return m.GetCasesFunc(projectID, suiteID, sectionID)
	}
	return data.GetCasesResponse{}, nil
}

func (m *MockClient) GetSections(projectID, suiteID int64) (data.GetSectionsResponse, error) {
	if m.GetSectionsFunc != nil {
		return m.GetSectionsFunc(projectID, suiteID)
	}
	return data.GetSectionsResponse{}, nil
}

// ImportTestCase — универсальная структура для тестов импорта (generics)
// `T` здесь — именно тип *ответа* от Client (например, data.GetSharedStepsResponse),
// чтобы тесты использовали именованные response-типы напрямую.
type ImportTestCase[T any] struct {
	name         string
	filtered     T
	dryRun       bool
	mockFunc     interface{} // func для mock (type assertion в тестах)
	wantCount    int         // для mapping.Count (SharedSteps/Suites)
	wantImported int         // для imported counter (Cases)
}

// getImportTestCases — универсальная функция кейсов (generics)
func getImportTestCases[T any]() []ImportTestCase[T] {
	return []ImportTestCase[T]{
		{
			name:         "dry-run — импорт пропущен",
			dryRun:       true,
			wantCount:    0,
			wantImported: 0,
		},
		{
			name:         "успешный импорт",
			dryRun:       false,
			wantCount:    1,
			wantImported: 1,
		},
		{
			name:         "ошибка импорта",
			dryRun:       false,
			wantCount:    0,
			wantImported: 0,
		},
		{
			name:         "много элементов — concurrency работает",
			dryRun:       false,
			wantCount:    3,
			wantImported: 3,
		},
	}
}

// logDir — директория для логов в тестах (создается, если не существует)
// Использует `utils.LogDir()` и создаёт подпапку `test_runs` для изоляции логов тестов.
func logDir() string {
	base := utils.LogDir()
	testPath := filepath.Join(base, "test_runs")
	if err := os.MkdirAll(testPath, 0755); err != nil {
		panic(err)
	}
	return testPath
}

// setupTestMigration использует оригинальный конструктор из types.go
func setupTestMigration(t *testing.T, mock *MockClient) *Migration {
	m, err := NewMigration(mock, 1, 10, 2, 20, "Title", logDir())
	if err != nil {
		t.Fatalf("Failed to setup migration: %v", err)
	}

	// Можно добавить автоматический вызов Close по завершению теста
	t.Cleanup(func() {
		m.Close()
	})

	return m
}
