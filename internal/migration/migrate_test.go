package migration

import (
	"errors"
	"testing"

	"gotr/internal/models/data"

	"github.com/stretchr/testify/assert"
)

// TestMigration_MigrateSuites проверяет поведение миграции для сущностей suites
func TestMigration_MigrateSuites(t *testing.T) {
	t.Run("Успешная миграция suites", func(t *testing.T) {
		mock := &MockClient{
			GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
				if projectID == 1 {
					return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
				}
				return data.GetSuitesResponse{}, nil
			},
			AddSuiteFunc: func(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
				return &data.Suite{ID: 100, Name: req.Name}, nil
			},
		}

		m := setupTestMigration(t, mock)
		m.compareField = "Name" // Переопределяем поле для теста

		err := m.MigrateSuites(false)
		assert.NoError(t, err)

		id, exists := m.mapping.GetTargetBySource(10)
		assert.True(t, exists)
		assert.Equal(t, int64(100), id)
	})
}

// TestMigration_MigrateSharedSteps проверяет миграцию общих шагов (shared steps)
func TestMigration_MigrateSharedSteps(t *testing.T) {
	t.Run("Миграция только неиспользуемых шагов", func(t *testing.T) {
		mock := &MockClient{
			GetSharedStepsFunc: func(p int64) (data.GetSharedStepsResponse, error) {
				if p == 1 { // source
					return data.GetSharedStepsResponse{
						{ID: 1, Title: "Used", CaseIDs: []int64{10}},
						{ID: 2, Title: "Free", CaseIDs: []int64{}},
					}, nil
				}
				// target — ВОТ ЗДЕСЬ ИЗМЕНИ 2 на 200
				return data.GetSharedStepsResponse{
					{ID: 200, Title: "Free", CaseIDs: []int64{}},
				}, nil
			},
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateSharedSteps(false)

		assert.NoError(t, err)
		// Проверяем маппинг: шаг 2 (свободный) должен быть там
		id, exists := m.mapping.GetTargetBySource(2)
		assert.True(t, exists)
		assert.Equal(t, int64(200), id)
	})
}

// TestMigration_MigrateSections проверяет поведение миграции для разделов (sections)
func TestMigration_MigrateSections(t *testing.T) {
	t.Run("Ошибка при получении данных sections", func(t *testing.T) {
		mock := &MockClient{
			GetSectionsFunc: func(p, s int64) (data.GetSectionsResponse, error) {
				return data.GetSectionsResponse{}, errors.New("fail to fetch sections")
			},
		}
		m := setupTestMigration(t, mock)
		err := m.MigrateSections(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fail to fetch sections")
	})
}

// TestMigration_MigrateCases проверяет миграцию тест-кейсов (cases)
func TestMigration_MigrateCases(t *testing.T) {
	t.Run("Успешная миграция кейсов", func(t *testing.T) {
		mock := &MockClient{
			GetCasesFunc: func(p, s, sec int64) (data.GetCasesResponse, error) {
				if p == 1 {
					return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
				}
				return data.GetCasesResponse{}, nil
			},
			AddCaseFunc: func(suiteID int64, req *data.AddCaseRequest) (*data.Case, error) {
				return &data.Case{ID: 100, Title: req.Title}, nil
			},
		}
		m := setupTestMigration(t, mock)
		err := m.MigrateCases(false)
		assert.NoError(t, err)
	})
}

// TestMigration_MigrateFull проверяет последовательную полную миграцию (full flow)
func TestMigration_MigrateFull(t *testing.T) {
	t.Run("Остановка при ошибке на первом этапе (Suites)", func(t *testing.T) {
		mock := &MockClient{
			GetSuitesFunc: func(p int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{}, errors.New("suites_critical_error")
			},
			// Важно: заглушки для остальных, чтобы не падало
			GetSectionsFunc:    func(p, s int64) (data.GetSectionsResponse, error) { return data.GetSectionsResponse{}, nil },
			GetSharedStepsFunc: func(p int64) (data.GetSharedStepsResponse, error) { return data.GetSharedStepsResponse{}, nil },
			GetCasesFunc:       func(p, s, sec int64) (data.GetCasesResponse, error) { return nil, nil },
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateFull(false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "suites_critical_error")
	})

	t.Run("Остановка при ошибке в sections — cases не запускаются", func(t *testing.T) {
		calledCases := false

		mock := &MockClient{
			GetSuitesFunc: func(p int64) (data.GetSuitesResponse, error) { return data.GetSuitesResponse{}, nil },
			GetSectionsFunc: func(p, s int64) (data.GetSectionsResponse, error) {
				return data.GetSectionsResponse{}, errors.New("sections_error")
			},
			GetCasesFunc: func(p, s, sec int64) (data.GetCasesResponse, error) {
				calledCases = true
				return data.GetCasesResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateFull(false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sections_error")
		assert.False(t, calledCases, "MigrateCases не должен был вызываться")
	})

	t.Run("DryRun не вызывает создание сущностей", func(t *testing.T) {
		addSuiteCalled := false
		addSectionCalled := false
		addSharedStepCalled := false
		addCaseCalled := false

		mock := &MockClient{
			GetSuitesFunc: func(p int64) (data.GetSuitesResponse, error) { return data.GetSuitesResponse{{ID: 1}}, nil },
			AddSuiteFunc: func(p int64, r *data.AddSuiteRequest) (*data.Suite, error) {
				addSuiteCalled = true
				return nil, nil
			},
			GetSectionsFunc: func(p, s int64) (data.GetSectionsResponse, error) { return data.GetSectionsResponse{}, nil },
			AddSectionFunc: func(p int64, r *data.AddSectionRequest) (*data.Section, error) {
				addSectionCalled = true
				return nil, nil
			},
			GetSharedStepsFunc: func(p int64) (data.GetSharedStepsResponse, error) { return data.GetSharedStepsResponse{}, nil },
			AddSharedStepFunc: func(p int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
				addSharedStepCalled = true
				return nil, nil
			},
			GetCasesFunc: func(p, s, sec int64) (data.GetCasesResponse, error) { return data.GetCasesResponse{}, nil },
			AddCaseFunc: func(s int64, r *data.AddCaseRequest) (*data.Case, error) {
				addCaseCalled = true
				return nil, nil
			},
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateFull(true)

		assert.NoError(t, err)
		assert.False(t, addSuiteCalled, "AddSuite не должен вызываться в dryRun")
		assert.False(t, addSectionCalled, "AddSection не должен вызываться в dryRun")
		assert.False(t, addSharedStepCalled, "AddSharedStep не должен вызываться в dryRun")
		assert.False(t, addCaseCalled, "AddCase не должен вызываться в dryRun")
	})

	t.Run("Успешная полная миграция", func(t *testing.T) {
		mock := &MockClient{
			GetSuitesFunc:      func(p int64) (data.GetSuitesResponse, error) { return data.GetSuitesResponse{{ID: 10}}, nil },
			AddSuiteFunc:       func(p int64, r *data.AddSuiteRequest) (*data.Suite, error) { return &data.Suite{ID: 100}, nil },
			GetSectionsFunc:    func(p, s int64) (data.GetSectionsResponse, error) { return data.GetSectionsResponse{{ID: 20}}, nil },
			AddSectionFunc:     func(p int64, r *data.AddSectionRequest) (*data.Section, error) { return &data.Section{ID: 200}, nil },
			GetSharedStepsFunc: func(p int64) (data.GetSharedStepsResponse, error) { return data.GetSharedStepsResponse{{ID: 30}}, nil },
			AddSharedStepFunc: func(p int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
				return &data.SharedStep{ID: 300}, nil
			},
			GetCasesFunc: func(p, s, sec int64) (data.GetCasesResponse, error) { return data.GetCasesResponse{{ID: 40}}, nil },
			AddCaseFunc:  func(s int64, r *data.AddCaseRequest) (*data.Case, error) { return &data.Case{ID: 400}, nil },
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateFull(false)

		assert.NoError(t, err)
		// Можно проверить mapping (если добавишь counter — проверить imported*)
	})
}
