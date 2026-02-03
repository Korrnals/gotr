// internal/migration/filter.go
package migration

import (
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/utils"
)

// FilterSharedSteps — фильтрация shared steps по использованию в suite и дубликатам в target
// Кандидаты — shared steps, не используемые в source suite (по CaseIDs)
// Дубликаты — добавляются в mapping (status "existing")
// Новые — возвращаются для импорта
func (m *Migration) FilterSharedSteps(source, target data.GetSharedStepsResponse, sourceCaseIDs map[int64]struct{}) (filtered data.GetSharedStepsResponse, err error) {
	m.logger.Info("Начало фильтрации shared steps по использованию в suite")

	var candidates data.GetSharedStepsResponse
	for _, step := range source {
		used := false
		for _, id := range step.CaseIDs {
			if _, ok := sourceCaseIDs[id]; ok {
				used = true
				break
			}
		}
		if !used {
			candidates = append(candidates, step)
		}
	}
	m.logger.Infow("Найдено кандидатов на перенос (не используются в suite)", "count", len(candidates))

	m.logger.Info("Проверка дубликатов в target проекте")
	targetMap := make(map[string]int64)
	for _, t := range target {
		val := utils.GetFieldValue(t, m.compareField)
		if val != "" {
			targetMap[val] = t.ID
		}
	}

	for _, step := range candidates {
		val := utils.GetFieldValue(step, m.compareField)
		if existingID, ok := targetMap[val]; ok {
			m.mapping.AddPair(step.ID, existingID, "existing")
			m.logger.Infow("Дубликат найден — добавлен в mapping", "title", step.Title, "old_id", step.ID, "existing_id", existingID)
		} else {
			filtered = append(filtered, step)
		}
	}

	m.logger.Infow("Готово к импорту новых shared steps", "count", len(filtered))
	return filtered, nil
}

// FilterSuites — фильтрация suites по дубликатам (по name)
// Дубликаты — добавляются в mapping (status "existing")
// Новые — возвращаются для импорта
func (m *Migration) FilterSuites(source, target data.GetSuitesResponse) (filtered data.GetSuitesResponse, err error) {
	m.logger.Info("Начало фильтрации suites по дубликатам (по name)")

	targetMap := make(map[string]int64)
	for _, t := range target {
		if t.Name != "" {
			targetMap[t.Name] = t.ID
		}
	}

	for _, s := range source {
		if existingID, ok := targetMap[s.Name]; ok {
			m.mapping.AddPair(s.ID, existingID, "existing")
			m.logger.Infow("Дубликат suite найден — добавлен в mapping", "name", s.Name, "old_id", s.ID, "existing_id", existingID)
		} else {
			filtered = append(filtered, s)
		}
	}

	m.logger.Infow("Готово к импорту новых suites", "count", len(filtered))
	return filtered, nil
}

// FilterCases — фильтрация cases по дубликатам (по compareField)
func (m *Migration) FilterCases(source, target data.GetCasesResponse) (filtered data.GetCasesResponse, err error) {
	m.logger.Info("Начало фильтрации кейсов по дубликатам")

	targetMap := make(map[string]int64)
	for _, t := range target {
		val := utils.GetFieldValue(t, m.compareField)
		if val != "" {
			targetMap[val] = t.ID
		}
	}

	for _, c := range source {
		val := utils.GetFieldValue(c, m.compareField)
		if _, exists := targetMap[val]; !exists {
			filtered = append(filtered, c)
		} else {
			m.logger.Infow("Дубликат кейса найден — пропущен", "title", c.Title)
		}
	}

	m.logger.Infow("Готово к импорту новых кейсов", "count", len(filtered))
	return filtered, nil
}

// FilterSections — фильтрация sections по дубликатам в target suite (по name)
func (m *Migration) FilterSections(source, target data.GetSectionsResponse) (filtered data.GetSectionsResponse, err error) {
	m.logger.Info("Начало фильтрации sections по дубликатам (по name в suite)")

	targetMap := make(map[string]int64)
	for _, t := range target {
		if t.Name != "" {
			targetMap[t.Name] = t.ID
		}
	}

	for _, s := range source {
		if existingID, ok := targetMap[s.Name]; ok {
			m.mapping.AddPair(s.ID, existingID, "existing")
			m.logger.Infow("Дубликат section найден — mapping добавлен", "name", s.Name, "old_id", s.ID, "existing_id", existingID)
		} else {
			filtered = append(filtered, s)
		}
	}

	m.logger.Infow("Готово к импорту новых sections", "count", len(filtered))
	return filtered, nil
}
