// internal/migration/import.go
package migration

import (
	"context"
	"fmt"
	"sync"

	"github.com/Korrnals/gotr/internal/models/data"
)

// maxImportConcurrency limits the number of parallel API calls during import
// to avoid overwhelming the TestRail server.
const maxImportConcurrency = 10

// ImportSharedSteps imports filtered shared steps in parallel.
// Updates the mapping (AddPair with status "created" for new IDs).
// Logs success/error entries from goroutines.
func (m *Migration) ImportSharedSteps(ctx context.Context, filtered data.GetSharedStepsResponse, dryRun bool) error {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run or no data — shared steps import skipped", "count", len(filtered))
		return nil
	}

	m.logger.Infow("Starting shared steps import", "count", len(filtered))

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxImportConcurrency)

	for _, step := range filtered {
		wg.Add(1)
		sem <- struct{}{}
		go func(s data.SharedStep) {
			defer func() { <-sem }()
			defer wg.Done()

			// Prepare request (deep copy steps)
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

			// Create in target project
			created, err := m.Client.AddSharedStep(ctx, m.dstProject, req)
			if err != nil {
				mu.Lock()
				m.logger.Errorw("Error importing shared step", "title", s.Title, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			m.mapping.AddPair(s.ID, created.ID, "created")
			m.importedCases++
			m.logger.Infow("Successfully created shared step", "old_id", s.ID, "new_id", created.ID, "title", s.Title)
			mu.Unlock()
		}(step)
	}
	wg.Wait()

	m.logger.Infow("Shared steps import completed", "imported", m.importedCases)
	return nil
}

// ImportSuites imports filtered suites in parallel.
// Updates the mapping (AddPair with status "created" for new IDs).
func (m *Migration) ImportSuites(ctx context.Context, filtered data.GetSuitesResponse, dryRun bool) error {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run or no data — suites import skipped", "count", len(filtered))
		return nil
	}

	m.logger.Infow("Starting suites import", "count", len(filtered))

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxImportConcurrency)

	for _, suite := range filtered {
		wg.Add(1)
		sem <- struct{}{}
		go func(s data.Suite) {
			defer func() { <-sem }()
			defer wg.Done()

			// Prepare request
			req := &data.AddSuiteRequest{
				Name:        s.Name,
				Description: s.Description,
			}

			// Create in target project
			created, err := m.Client.AddSuite(ctx, m.dstProject, req)
			if err != nil {
				mu.Lock()
				m.logger.Errorw("Error importing suite", "name", s.Name, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			m.mapping.AddPair(s.ID, created.ID, "created")
			m.importedCases++
			m.logger.Infow("Successfully created suite", "old_id", s.ID, "new_id", created.ID, "name", s.Name)
			mu.Unlock()
		}(suite)
	}
	wg.Wait()

	m.logger.Infow("Suites import completed", "imported", m.importedCases)
	return nil
}

// ImportSections imports filtered sections in parallel.
// Updates the mapping (AddPair with status "created" for new IDs).
func (m *Migration) ImportSections(ctx context.Context, filtered data.GetSectionsResponse, dryRun bool) error {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run or no data — sections import skipped", "count", len(filtered))
		return nil
	}

	m.logger.Infow("Starting sections import", "count", len(filtered))

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxImportConcurrency)

	for _, section := range filtered {
		wg.Add(1)
		sem <- struct{}{}
		go func(s data.Section) {
			defer func() { <-sem }()
			defer wg.Done()

			// Prepare request
			req := &data.AddSectionRequest{
				Name:        s.Name,
				Description: s.Description,
				SuiteID:     s.SuiteID,
				ParentID:    s.ParentID,
			}

			// Create in target project
			created, err := m.Client.AddSection(ctx, m.dstProject, req)
			if err != nil {
				mu.Lock()
				m.logger.Errorw("Error importing section", "name", s.Name, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			m.mapping.AddPair(s.ID, created.ID, "created")
			m.importedCases++
			m.logger.Infow("Successfully created section", "old_id", s.ID, "new_id", created.ID, "name", s.Name)
			mu.Unlock()
		}(section)
	}
	wg.Wait()

	m.logger.Infow("Sections import completed", "imported", m.importedCases)
	return nil
}

// ImportCases imports filtered cases in parallel.
// Replaces SharedStepID references using the mapping.
func (m *Migration) ImportCases(ctx context.Context, filtered data.GetCasesResponse, dryRun bool) error {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run or no data — cases import skipped", "count", len(filtered))
		return nil
	}

	m.logger.Infow("Starting cases import", "count", len(filtered))

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxImportConcurrency)

	for _, c := range filtered {
		wg.Add(1)
		sem <- struct{}{}
		go func(caseData data.Case) {
			defer func() { <-sem }()
			defer wg.Done()

			// Prepare request
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
					}
				}

				req.CustomStepsSeparated[i] = newStep
			}

			// Create in target suite
			created, err := m.Client.AddCase(ctx, m.dstSuite, req)
			if err != nil {
				mu.Lock()
				m.logger.Errorw("Error importing case", "title", caseData.Title, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			m.importedCases++
			m.logger.Infow("Successfully created case", "old_id", caseData.ID, "new_id", created.ID, "title", caseData.Title)
			mu.Unlock()
		}(c)
	}
	wg.Wait()

	m.logger.Infow("Cases import completed", "imported", m.importedCases)
	return nil
}

// ImportCasesReport is like ImportCases but returns lists of created IDs and errors for CLI reporting.
func (m *Migration) ImportCasesReport(ctx context.Context, filtered data.GetCasesResponse, dryRun bool) (createdIDs []int64, errs []string, err error) {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run or no data — cases import skipped", "count", len(filtered))
		return nil, nil, nil
	}

	m.logger.Infow("Starting cases import (report)", "count", len(filtered))

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxImportConcurrency)

	for _, c := range filtered {
		wg.Add(1)
		sem <- struct{}{}
		go func(caseData data.Case) {
			defer func() { <-sem }()
			defer wg.Done()

			// Prepare request
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
					}
				}

				req.CustomStepsSeparated[i] = newStep
			}

			// Create in target suite
			created, err := m.Client.AddCase(ctx, m.dstSuite, req)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("case %q: %v", caseData.Title, err))
				m.logger.Errorw("Error importing case", "title", caseData.Title, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			createdIDs = append(createdIDs, created.ID)
			m.importedCases++
			m.logger.Infow("Successfully created case (report)", "old_id", caseData.ID, "new_id", created.ID, "title", caseData.Title)
			mu.Unlock()
		}(c)
	}
	wg.Wait()

	m.logger.Infow("Cases import (report) completed", "imported", m.importedCases)
	return createdIDs, errs, nil
}
