// internal/migration/migration.go
package migration

import (
	"context"
	"fmt"
)

// MigrateSharedSteps runs the full shared steps migration cycle: fetch, filter, import.
func (m *Migration) MigrateSharedSteps(ctx context.Context, dryRun bool) error {
	m.logger.Info("Starting shared steps migration")

	source, target, err := m.FetchSharedStepsData(ctx)
	if err != nil {
		return fmt.Errorf("MigrateSharedSteps: %w", err)
	}

	sourceCases, err := m.Client.GetCases(ctx, m.srcProject, m.srcSuite, 0)
	if err != nil {
		return fmt.Errorf("MigrateSharedSteps: %w", err)
	}
	caseIDsSet := make(map[int64]struct{})
	for _, c := range sourceCases {
		caseIDsSet[c.ID] = struct{}{}
	}

	filtered, _ := m.FilterSharedSteps(source, target, caseIDsSet)
	m.lastFilteredSteps = filtered

	return m.ImportSharedSteps(ctx, filtered, dryRun)
}

// MigrateSuites runs the full suites migration cycle: fetch, filter, import.
func (m *Migration) MigrateSuites(ctx context.Context, dryRun bool) error {
	m.logger.Info("Starting suites migration")

	source, target, err := m.FetchSuitesData(ctx)
	if err != nil {
		return fmt.Errorf("MigrateSuites: %w", err)
	}

	filtered, _ := m.FilterSuites(source, target)

	return m.ImportSuites(ctx, filtered, dryRun)
}

// MigrateCases runs the full cases migration cycle: fetch, filter, import.
func (m *Migration) MigrateCases(ctx context.Context, dryRun bool) error {
	m.logger.Info("Starting cases migration")

	source, target, err := m.FetchCasesData(ctx)
	if err != nil {
		return fmt.Errorf("MigrateCases: %w", err)
	}

	filtered, _ := m.FilterCases(source, target)

	return m.ImportCases(ctx, filtered, dryRun)
}

// MigrateCasesReport runs cases migration and returns created case IDs and per-case errors.
func (m *Migration) MigrateCasesReport(ctx context.Context, dryRun bool) (createdIDs []int64, errs []string, err error) {
	m.logger.Info("Starting cases migration (report)")

	source, target, err := m.FetchCasesData(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("MigrateCasesReport: %w", err)
	}

	filtered, _ := m.FilterCases(source, target)

	createdIDs, errs, err = m.ImportCasesReport(ctx, filtered, dryRun)
	if err != nil {
		return nil, nil, fmt.Errorf("MigrateCasesReport: %w", err)
	}

	return createdIDs, errs, nil
}

// MigrateSections runs the full sections migration cycle: fetch, filter, import.
func (m *Migration) MigrateSections(ctx context.Context, dryRun bool) error {
	m.logger.Info("Starting sections migration")

	source, target, err := m.FetchSectionsData(ctx)
	if err != nil {
		return fmt.Errorf("MigrateSections: %w", err)
	}

	filtered, _ := m.FilterSections(source, target)

	return m.ImportSections(ctx, filtered, dryRun) // reuses ImportSections from import.go
}

// MigrateFull runs the full migration in order: suites → sections → shared steps → cases.
func (m *Migration) MigrateFull(ctx context.Context, dryRun bool) error {
	m.logger.Info("Starting full migration")

	if err := m.MigrateSuites(ctx, dryRun); err != nil {
		m.logger.Errorw("Suites migration error — full migration aborted", "error", err)
		return fmt.Errorf("MigrateFull: %w", err)
	}

	if err := m.MigrateSections(ctx, dryRun); err != nil {
		m.logger.Errorw("Sections migration error — full migration aborted", "error", err)
		return fmt.Errorf("MigrateFull: %w", err)
	}

	if err := m.MigrateSharedSteps(ctx, dryRun); err != nil {
		m.logger.Errorw("Shared steps migration error — full migration aborted", "error", err)
		return fmt.Errorf("MigrateFull: %w", err)
	}

	if err := m.MigrateCases(ctx, dryRun); err != nil {
		m.logger.Errorw("Cases migration error — full migration aborted", "error", err)
		return fmt.Errorf("MigrateFull: %w", err)
	}

	m.logger.Info("Full migration completed successfully")
	return nil
}
