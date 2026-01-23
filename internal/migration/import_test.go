// internal/migration/import_test.go
package migration // white-box тесты — в том же пакете

import (
	"errors"
	"testing"

	"gotr/internal/models/data"

	"github.com/stretchr/testify/assert"
)

func TestMigration_ImportSharedSteps(t *testing.T) {
	// Используем named response-тип напрямую — это отражает реальный контракт клиента
	cases := getImportTestCases[data.GetSharedStepsResponse]()
	for i, base := range cases {
		tt := base // копия для модификации

		// Заполняем конкретные данные по индексу кейса
		switch i {
		case 0: // dry-run — проверяем безопасность (импорт не происходит, mapping не меняется)
			tt.filtered = data.GetSharedStepsResponse{{ID: 1, Title: "Test"}}
		case 1: // успешный импорт — проверяем обновление mapping.Count (API работает)
			tt.filtered = data.GetSharedStepsResponse{{ID: 1, Title: "Test"}}
			tt.mockFunc = func(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
				return &data.SharedStep{ID: 100}, nil // симулируем успешный ответ API
			}
		case 2: // ошибка импорта — проверяем устойчивость (не паникуем, mapping не обновляется)
			tt.filtered = data.GetSharedStepsResponse{{ID: 1, Title: "Test"}}
			tt.mockFunc = func(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
				return nil, errors.New("API error") // симулируем ошибку API
			}
		case 3: // concurrency — много элементов, проверяем параллельный импорт (race detector pass)
			tt.filtered = data.GetSharedStepsResponse{
				{ID: 1, Title: "A"},
				{ID: 2, Title: "B"},
				{ID: 3, Title: "C"},
			}
			tt.mockFunc = func(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
				return &data.SharedStep{ID: int64(len(req.Title)) + 100}, nil // заглушка по длине Title
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{}
			if tt.mockFunc != nil {
				mock.AddSharedStepFunc = tt.mockFunc.(func(int64, *data.AddSharedStepRequest) (*data.SharedStep, error))
			}

			m, err := NewMigration(mock, 1, 1, 2, 2, "title", logDir())
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()

			err = m.ImportSharedSteps(tt.filtered, tt.dryRun)
			if err != nil && i != 2 { // ошибка ожидаема только в кейсе ошибки
				t.Errorf("неожиданная ошибка: %v", err)
			}

			if m.mapping.Count != tt.wantCount {
				t.Errorf("mapping.Count = %d, ожидалось %d", m.mapping.Count, tt.wantCount)
			}
		})
	}
}

func TestMigration_ImportSuites(t *testing.T) {
	cases := getImportTestCases[data.GetSuitesResponse]()
	for i, base := range cases {
		tt := base
		switch i {
		case 0: // dry-run — проверяем безопасность
			tt.filtered = data.GetSuitesResponse{{ID: 1, Name: "Test Suite"}}
		case 1: // успешный импорт — проверяем обновление mapping.Count
			tt.filtered = data.GetSuitesResponse{{ID: 1, Name: "Test Suite"}}
			tt.mockFunc = func(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
				return &data.Suite{ID: 100}, nil
			}
		case 2: // ошибка импорта — проверяем устойчивость
			tt.filtered = data.GetSuitesResponse{{ID: 1, Name: "Test Suite"}}
			tt.mockFunc = func(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
				return nil, errors.New("API error")
			}
		case 3: // concurrency — проверяем параллельный импорт
			tt.filtered = data.GetSuitesResponse{
				{ID: 1, Name: "A"},
				{ID: 2, Name: "B"},
				{ID: 3, Name: "C"},
			}
			tt.mockFunc = func(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
				return &data.Suite{ID: int64(len(req.Name)) + 100}, nil
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{}
			if tt.mockFunc != nil {
				mock.AddSuiteFunc = tt.mockFunc.(func(int64, *data.AddSuiteRequest) (*data.Suite, error))
			}

			m, err := NewMigration(mock, 1, 1, 2, 2, "title", logDir())
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()

			err = m.ImportSuites(tt.filtered, tt.dryRun)
			if err != nil && i != 2 {
				t.Errorf("неожиданная ошибка: %v", err)
			}

			if m.mapping.Count != tt.wantCount {
				t.Errorf("mapping.Count = %d, ожидалось %d", m.mapping.Count, tt.wantCount)
			}
		})
	}
}

func TestMigration_ImportCases(t *testing.T) {
	cases := getImportTestCases[data.GetCasesResponse]()
	for i, base := range cases {
		tt := base
		switch i {
		case 0: // dry-run — проверяем безопасность
			tt.filtered = data.GetCasesResponse{{ID: 1, Title: "Test Case"}}
		case 1: // успешный импорт — проверяем imported (не меняет mapping)
			tt.filtered = data.GetCasesResponse{{ID: 1, Title: "Test Case"}}
			tt.mockFunc = func(suiteID int64, req *data.AddCaseRequest) (*data.Case, error) {
				return &data.Case{ID: 100}, nil
			}
		case 2: // ошибка импорта — проверяем устойчивость
			tt.filtered = data.GetCasesResponse{{ID: 1, Title: "Test Case"}}
			tt.mockFunc = func(suiteID int64, req *data.AddCaseRequest) (*data.Case, error) {
				return nil, errors.New("API error")
			}
		case 3: // concurrency — проверяем параллельный импорт
			tt.filtered = data.GetCasesResponse{
				{ID: 1, Title: "A"},
				{ID: 2, Title: "B"},
				{ID: 3, Title: "C"},
			}
			tt.mockFunc = func(suiteID int64, req *data.AddCaseRequest) (*data.Case, error) {
				return &data.Case{ID: int64(len(req.Title)) + 100}, nil
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{}
			if tt.mockFunc != nil {
				mock.AddCaseFunc = tt.mockFunc.(func(int64, *data.AddCaseRequest) (*data.Case, error))
			}

			m, err := NewMigration(mock, 1, 1, 2, 2, "title", logDir())
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()

			err = m.ImportCases(tt.filtered, tt.dryRun)
			if err != nil && i != 2 {
				t.Errorf("неожиданная ошибка: %v", err)
			}

			// Для Cases mapping.Count не меняется — проверяем importedCases (если добавишь counter)
			assert.Equal(t, tt.wantImported, m.importedCases)
		})
	}
}

func TestMigration_ImportSections(t *testing.T) {
	cases := getImportTestCases[data.GetSectionsResponse]()
	for i, base := range cases {
		tt := base
		switch i {
		case 0: // dry-run — проверяем безопасность
			tt.filtered = data.GetSectionsResponse{{ID: 1, Name: "Test Section"}}
		case 1: // успешный импорт — проверяем обновление mapping.Count
			tt.filtered = data.GetSectionsResponse{{ID: 1, Name: "Test Section"}}
			tt.mockFunc = func(projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
				return &data.Section{ID: 100}, nil
			}
		case 2: // ошибка импорта — проверяем устойчивость
			tt.filtered = data.GetSectionsResponse{{ID: 1, Name: "Test Section"}}
			tt.mockFunc = func(projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
				return nil, errors.New("API error")
			}
		case 3: // concurrency — проверяем параллельный импорт
			tt.filtered = data.GetSectionsResponse{
				{ID: 1, Name: "A"},
				{ID: 2, Name: "B"},
				{ID: 3, Name: "C"},
			}
			tt.mockFunc = func(projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
				return &data.Section{ID: int64(len(req.Name)) + 100}, nil
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{}
			if tt.mockFunc != nil {
				mock.AddSectionFunc = tt.mockFunc.(func(int64, *data.AddSectionRequest) (*data.Section, error))
			}

			m, err := NewMigration(mock, 1, 1, 2, 2, "title", logDir())
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()

			err = m.ImportSections(tt.filtered, tt.dryRun)
			if err != nil && i != 2 {
				t.Errorf("неожиданная ошибка: %v", err)
			}

			if m.mapping.Count != tt.wantCount {
				t.Errorf("mapping.Count = %d, ожидалось %d", m.mapping.Count, tt.wantCount)
			}
		})
	}
}
