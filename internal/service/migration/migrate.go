// internal/migration/migration.go
package migration

// MigrateSharedSteps — полный цикл миграции shared steps
func (m *Migration) MigrateSharedSteps(dryRun bool) error {
	m.logger.Info("Начало миграции shared steps")

	source, target, err := m.FetchSharedStepsData()
	if err != nil {
		return err
	}

	sourceCases, err := m.Client.GetCases(m.srcProject, m.srcSuite, 0)
	if err != nil {
		return err
	}
	caseIDsSet := make(map[int64]struct{})
	for _, c := range sourceCases {
		caseIDsSet[c.ID] = struct{}{}
	}

	filtered, err := m.FilterSharedSteps(source, target, caseIDsSet)
	if err != nil {
		return err
	}

	return m.ImportSharedSteps(filtered, dryRun)
}

// MigrateSuites — полный цикл миграции suites
func (m *Migration) MigrateSuites(dryRun bool) error {
	m.logger.Info("Начало миграции suites")

	source, target, err := m.FetchSuitesData()
	if err != nil {
		return err
	}

	filtered, err := m.FilterSuites(source, target)
	if err != nil {
		return err
	}

	return m.ImportSuites(filtered, dryRun)
}

// MigrateCases — полный цикл миграции cases
func (m *Migration) MigrateCases(dryRun bool) error {
	m.logger.Info("Начало миграции cases")

	source, target, err := m.FetchCasesData()
	if err != nil {
		return err
	}

	filtered, err := m.FilterCases(source, target)
	if err != nil {
		return err
	}

	return m.ImportCases(filtered, dryRun)
}

// MigrateSections — полный цикл миграции sections (переиспользуем методы)
func (m *Migration) MigrateSections(dryRun bool) error {
	m.logger.Info("Начало миграции sections")

	source, target, err := m.FetchSectionsData()
	if err != nil {
		return err
	}

	filtered, err := m.FilterSections(source, target)
	if err != nil {
		return err
	}

	return m.ImportSections(filtered, dryRun) // переиспользуем ImportSections из import.go
}

// MigrateFull — полная миграция (порядок: suites → sections → shared steps → cases)
func (m *Migration) MigrateFull(dryRun bool) error {
	m.logger.Info("Начало полной миграции")

	if err := m.MigrateSuites(dryRun); err != nil {
		m.logger.Errorw("Ошибка миграции suites — полная миграция прервана", "error", err)
		return err
	}

	if err := m.MigrateSections(dryRun); err != nil {
		m.logger.Errorw("Ошибка миграции sections — полная миграция прервана", "error", err)
		return err
	}

	if err := m.MigrateSharedSteps(dryRun); err != nil {
		m.logger.Errorw("Ошибка миграции shared steps — полная миграция прервана", "error", err)
		return err
	}

	if err := m.MigrateCases(dryRun); err != nil {
		m.logger.Errorw("Ошибка миграции cases — полная миграция прервана", "error", err)
		return err
	}

	m.logger.Info("Полная миграция завершена успешно")
	return nil
}
