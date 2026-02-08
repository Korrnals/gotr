package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/utils"
)

// MockClient — алиас для client.MockClient для обратной совместимости
type MockClient = client.MockClient

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
