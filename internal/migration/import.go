// internal/migration/import.go
package migration

import (
	"fmt"
	"gotr/internal/models/data"
	"sync"
)

// ImportSharedSteps — импортирует отфильтрованные shared steps параллельно
// Обновляет mapping (AddPair для новых ID, status "created")
// Добавляет записи в лог из горутин (успех/ошибки)
func (m *Migration) ImportSharedSteps(filtered data.GetSharedStepsResponse, dryRun bool) error {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run или нет данных — импорт shared steps пропущен", "count", len(filtered))
		return nil
	}

	m.logger.Infow("Начало импорта shared steps", "count", len(filtered))

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, step := range filtered {
		wg.Add(1)
		go func(s data.SharedStep) {
			defer wg.Done()

			// Подготовка request (глубокое копирование шагов)
			req := &data.AddSharedStepRequest{
				Title:                s.Title,
				CustomStepsSeparated: make([]data.Step, len(s.CustomStepsSeparated)),
			}

			for i, orig := range s.CustomStepsSeparated {
				req.CustomStepsSeparated[i] = data.Step{
					Content:        orig.Content,
					AdditionalInfo: orig.AdditionalInfo,
					Expected:       orig.Expected,
					Refs:           orig.Refs,
				}
			}

			// Создание в target проекте
			created, err := m.Client.AddSharedStep(m.dstProject, req)
			if err != nil {
				mu.Lock()
				m.logger.Errorw("Ошибка импорта shared step", "title", s.Title, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			m.mapping.AddPair(s.ID, created.ID, "created")
			m.importedCases++
			m.logger.Infow("Успешно создан shared step", "old_id", s.ID, "new_id", created.ID, "title", s.Title)
			mu.Unlock()
		}(step)
	}
	wg.Wait()

	m.logger.Infow("Импорт shared steps завершён", "imported", m.importedCases)
	return nil
}

// ImportSuites — импортирует отфильтрованные suites параллельно
// Обновляет mapping (AddPair для новых ID, status "created")
// Добавляет записи в лог
func (m *Migration) ImportSuites(filtered data.GetSuitesResponse, dryRun bool) error {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run или нет данных — импорт suites пропущен", "count", len(filtered))
		return nil
	}

	m.logger.Infow("Начало импорта suites", "count", len(filtered))

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, suite := range filtered {
		wg.Add(1)
		go func(s data.Suite) {
			defer wg.Done()

			// Подготовка request
			req := &data.AddSuiteRequest{
				Name:        s.Name,
				Description: s.Description,
			}

			// Создание в target проекте
			created, err := m.Client.AddSuite(m.dstProject, req)
			if err != nil {
				mu.Lock()
				m.logger.Errorw("Ошибка импорта suite", "name", s.Name, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			m.mapping.AddPair(s.ID, created.ID, "created")
			m.importedCases++
			m.logger.Infow("Успешно создан suite", "old_id", s.ID, "new_id", created.ID, "name", s.Name)
			mu.Unlock()
		}(suite)
	}
	wg.Wait()

	m.logger.Infow("Импорт suites завершён", "imported", m.importedCases)
	return nil
}

// ImportSections — импортирует отфильтрованные sections параллельно
// Обновляет mapping (AddPair для новых ID, status "created")
// Добавляет записи в лог
func (m *Migration) ImportSections(filtered data.GetSectionsResponse, dryRun bool) error {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run или нет данных — импорт sections пропущен", "count", len(filtered))
		return nil
	}

	m.logger.Infow("Начало импорта sections", "count", len(filtered))

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, section := range filtered {
		wg.Add(1)
		go func(s data.Section) {
			defer wg.Done()

			// Подготовка request
			req := &data.AddSectionRequest{
				Name:        s.Name,
				Description: s.Description,
				SuiteID:     s.SuiteID,
				ParentID:    s.ParentID,
			}

			// Создание в target проекте
			created, err := m.Client.AddSection(m.dstProject, req)
			if err != nil {
				mu.Lock()
				m.logger.Errorw("Ошибка импорта section", "name", s.Name, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			m.mapping.AddPair(s.ID, created.ID, "created")
			m.importedCases++
			m.logger.Infow("Успешно создан section", "old_id", s.ID, "new_id", created.ID, "name", s.Name)
			mu.Unlock()
		}(section)
	}
	wg.Wait()

	m.logger.Infow("Импорт sections завершён", "imported", m.importedCases)
	return nil
}

// ImportCases — импортирует отфильтрованные cases параллельно
// Заменяет SharedStepID по mapping
// Добавляет записи в лог
func (m *Migration) ImportCases(filtered data.GetCasesResponse, dryRun bool) error {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run или нет данных — импорт cases пропущен", "count", len(filtered))
		return nil
	}

	m.logger.Infow("Начало импорта cases", "count", len(filtered))

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, c := range filtered {
		wg.Add(1)
		go func(caseData data.Case) {
			defer wg.Done()

			// Подготовка request
			req := &data.AddCaseRequest{
				Title:                caseData.Title,
				TypeID:               caseData.TypeID,
				PriorityID:           caseData.PriorityID,
				TemplateID:           caseData.TemplateID,
				MilestoneID:          caseData.MilestoneID,
				Refs:                 caseData.Refs,
				CustomPreconds:       caseData.CustomPreconds,
				CustomStepsSeparated: make([]data.Step, len(caseData.CustomStepsSeparated)),
			}

			for i, orig := range caseData.CustomStepsSeparated {
				newStep := data.Step{
					Content:        orig.Content,
					AdditionalInfo: orig.AdditionalInfo,
					Expected:       orig.Expected,
					Refs:           orig.Refs,
					SharedStepID:   orig.SharedStepID,
				}

				if orig.SharedStepID != 0 {
					if newID, exists := m.mapping.GetTargetBySource(orig.SharedStepID); exists {
						newStep.SharedStepID = newID
						m.logger.Infow("Замена shared_step_id в кейсе", "title", caseData.Title, "old", orig.SharedStepID, "new", newID)
					} else {
						m.logger.Warnw("Shared step ID не найден в mapping", "old_id", orig.SharedStepID, "case_title", caseData.Title)
					}
				}

				req.CustomStepsSeparated[i] = newStep
			}

			// Создание в target suite
			created, err := m.Client.AddCase(m.dstSuite, req)
			if err != nil {
				mu.Lock()
				m.logger.Errorw("Ошибка импорта кейса", "title", caseData.Title, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			m.importedCases++
			m.logger.Infow("Успешно создан кейс", "old_id", caseData.ID, "new_id", created.ID, "title", caseData.Title)
			mu.Unlock()
		}(c)
	}
	wg.Wait()

	m.logger.Infow("Импорт cases завершён", "imported", m.importedCases)
	return nil
}

// ImportCasesReport — как ImportCases, но возвращает список созданных ID и ошибок для использования в CLI
func (m *Migration) ImportCasesReport(filtered data.GetCasesResponse, dryRun bool) ([]int64, []string, error) {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run или нет данных — импорт cases пропущен", "count", len(filtered))
		return nil, nil, nil
	}

	m.logger.Infow("Начало импорта cases (report)", "count", len(filtered))

	var mu sync.Mutex
	var wg sync.WaitGroup
	var createdIDs []int64
	var errors []string

	for _, c := range filtered {
		wg.Add(1)
		go func(caseData data.Case) {
			defer wg.Done()

			// Подготовка request
			req := &data.AddCaseRequest{
				Title:                caseData.Title,
				TypeID:               caseData.TypeID,
				PriorityID:           caseData.PriorityID,
				TemplateID:           caseData.TemplateID,
				MilestoneID:          caseData.MilestoneID,
				Refs:                 caseData.Refs,
				CustomPreconds:       caseData.CustomPreconds,
				CustomStepsSeparated: make([]data.Step, len(caseData.CustomStepsSeparated)),
			}

			for i, orig := range caseData.CustomStepsSeparated {
				newStep := data.Step{
					Content:        orig.Content,
					AdditionalInfo: orig.AdditionalInfo,
					Expected:       orig.Expected,
					Refs:           orig.Refs,
					SharedStepID:   orig.SharedStepID,
				}

				if orig.SharedStepID != 0 {
					if newID, exists := m.mapping.GetTargetBySource(orig.SharedStepID); exists {
						newStep.SharedStepID = newID
						m.logger.Infow("Замена shared_step_id в кейсе", "title", caseData.Title, "old", orig.SharedStepID, "new", newID)
					} else {
						m.logger.Warnw("Shared step ID не найден в mapping", "old_id", orig.SharedStepID, "case_title", caseData.Title)
					}
				}

				req.CustomStepsSeparated[i] = newStep
			}

			// Создание в target suite
			created, err := m.Client.AddCase(m.dstSuite, req)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Sprintf("кейс %q: %v", caseData.Title, err))
				m.logger.Errorw("Ошибка импорта кейса", "title", caseData.Title, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			createdIDs = append(createdIDs, created.ID)
			m.importedCases++
			m.logger.Infow("Успешно создан кейс (report)", "old_id", caseData.ID, "new_id", created.ID, "title", caseData.Title)
			mu.Unlock()
		}(c)
	}
	wg.Wait()

	m.logger.Infow("Импорт cases (report) завершён", "imported", m.importedCases)
	return createdIDs, errors, nil
}
