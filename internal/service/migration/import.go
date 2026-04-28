// internal/migration/import.go
package migration

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/service/casefields"
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
				m.failedImports++
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
				m.failedImports++
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

// ImportSections imports filtered sections level by level (by Depth), so parent sections are
// always created before their children. Within each depth level sections are imported in parallel.
// Parent IDs are resolved through the mapping (src→dst) so cross-project references are correct.
// Updates the mapping (AddPair with status "created" for new IDs).
func (m *Migration) ImportSections(ctx context.Context, filtered data.GetSectionsResponse, dryRun bool) error {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run or no data — sections import skipped", "count", len(filtered))
		return nil
	}

	m.logger.Infow("Starting sections import", "count", len(filtered))

	// Group sections by depth so parents are created before children.
	byDepth := make(map[int64][]data.Section)
	var depths []int64
	seen := make(map[int64]bool)
	for _, s := range filtered {
		if !seen[s.Depth] {
			seen[s.Depth] = true
			depths = append(depths, s.Depth)
		}
		byDepth[s.Depth] = append(byDepth[s.Depth], s)
	}
	// Sort depths ascending so root (depth 0) comes first.
	sort.Slice(depths, func(i, j int) bool { return depths[i] < depths[j] })

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxImportConcurrency)

	for _, depth := range depths {
		for _, section := range byDepth[depth] {
			wg.Add(1)
			sem <- struct{}{}
			go func(s data.Section) {
				defer func() { <-sem }()
				defer wg.Done()

				dstSuiteID := s.SuiteID
				if dstSuiteID != 0 {
					if mappedSuiteID, ok := m.mapping.GetTargetBySource(s.SuiteID); ok {
						dstSuiteID = mappedSuiteID
					} else if m.dstSuite != 0 {
						dstSuiteID = m.dstSuite
					}
				}

				// Resolve parent ID: map source parent section ID to destination parent section ID.
				// If the parent cannot be resolved, this section cannot be placed correctly —
				// we refuse to silently create it at root (which would corrupt the tree) and
				// record it as a failure so the final summary reflects the real state.
				dstParentID := int64(0)
				if s.ParentID != 0 {
					if mappedParentID, ok := m.mapping.GetTargetBySource(s.ParentID); ok {
						dstParentID = mappedParentID
					} else {
						mu.Lock()
						m.failedImports++
						m.logger.Errorw("Parent section not found in mapping — section skipped",
							"name", s.Name, "src_parent_id", s.ParentID)
						mu.Unlock()
						return
					}
				}

				// Prepare request
				req := &data.AddSectionRequest{
					Name:        s.Name,
					Description: s.Description,
					SuiteID:     dstSuiteID,
					ParentID:    dstParentID,
				}

				// Create in target project
				created, err := m.Client.AddSection(ctx, m.dstProject, req)
				if err != nil {
					mu.Lock()
					m.failedImports++
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
		wg.Wait() // wait for all sections at this depth before processing the next level
	}

	m.logger.Infow("Sections import completed", "imported", m.importedCases)
	return nil
}

// ImportCases imports filtered cases in parallel.
// Replaces SharedStepID references using the mapping.
//
// After parallel creation finishes, ImportCases restores the source order
// inside each destination section by calling move_cases_to_section with the
// case IDs in the same relative order they appeared in `filtered`. TestRail
// renders cases in the order they are received by add_case, which is
// non-deterministic when calls are issued concurrently — see
// reorderCreatedCases for the rationale.
func (m *Migration) ImportCases(ctx context.Context, filtered data.GetCasesResponse, dryRun bool) error {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run or no data — cases import skipped", "count", len(filtered))
		return nil
	}

	m.logger.Infow("Starting cases import", "count", len(filtered))

	sectionMap, err := m.resolveSectionMapByName(ctx)
	if err != nil {
		return err
	}

	autotestDefault, err := m.resolveAutotestOnDefault(ctx)
	if err != nil {
		return err
	}

	extraFields, err := m.resolveRequiredExtraFields(ctx)
	if err != nil {
		return err
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxImportConcurrency)
	perSection := make(map[int64][]createdCase)

	for i, c := range filtered {
		wg.Add(1)
		sem <- struct{}{}
		go func(srcOrder int, caseData data.Case) {
			defer func() { <-sem }()
			defer wg.Done()

			// Prepare request
			dstSectionID := m.resolveDestinationSectionID(caseData.SectionID, sectionMap)
			if dstSectionID == 0 {
				mu.Lock()
				m.failedImports++
				m.logger.Errorw("Error importing case", "title", caseData.Title, "error", "unable to resolve destination section_id")
				mu.Unlock()
				return
			}

			req := &data.AddCaseRequest{
				Title:          caseData.Title,
				SectionID:      dstSectionID,
				TypeID:         caseData.TypeID,
				PriorityID:     caseData.PriorityID,
				TemplateID:     caseData.TemplateID,
				MilestoneID:    caseData.MilestoneID,
				Refs:           caseData.Refs,
				CustomPreconds: caseData.CustomPreconds,
			}
			applyAutotestOnValue(req, caseData.CustomAutotestOn, autotestDefault)
			req.ExtraFields = extraFields
			req.CustomStepsSeparated = m.buildImportSteps(caseData, &mu)

			// Create in target section
			created, err := m.Client.AddCase(ctx, dstSectionID, req)
			if err != nil {
				mu.Lock()
				m.failedImports++
				m.logger.Errorw("Error importing case", "title", caseData.Title, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			perSection[dstSectionID] = append(perSection[dstSectionID], createdCase{srcOrder: srcOrder, newID: created.ID})
			m.importedCases++
			m.logger.Infow("Successfully created case", "old_id", caseData.ID, "new_id", created.ID, "title", caseData.Title)
			mu.Unlock()
		}(i, c)
	}
	wg.Wait()

	m.reorderCreatedCases(ctx, perSection)

	m.logger.Infow("Cases import completed", "imported", m.importedCases)
	return nil
}

// buildImportSteps builds the CustomStepsSeparated payload for AddCase.
//
// TestRail's get_case expands referenced shared steps inline: every sub-step of
// a shared step is returned as its own Step object carrying the same
// shared_step_id. When reimporting, we must collapse each contiguous run of
// identical shared_step_id rows into a single {shared_step_id:newID} reference;
// sending N duplicates causes add_case to reject the payload with
// "Field :shared_step_id is not a valid shared test step".
//
// Steps whose shared_step_id is unmapped fall back to inline content (best
// effort); steps without a shared_step_id are copied as inline content.
func (m *Migration) buildImportSteps(caseData data.Case, mu *sync.Mutex) []data.Step {
	if len(caseData.CustomStepsSeparated) == 0 {
		return nil
	}
	out := make([]data.Step, 0, len(caseData.CustomStepsSeparated))
	var lastSharedID int64 // last shared_step_id pushed into out (0 = none)
	for _, orig := range caseData.CustomStepsSeparated {
		if orig.SharedStepID != 0 {
			if newID, exists := m.mapping.GetTargetBySource(orig.SharedStepID); exists {
				if newID == lastSharedID {
					// Skip duplicate expansion row from the same shared step.
					continue
				}
				out = append(out, data.Step{SharedStepID: newID})
				lastSharedID = newID
				continue
			}
			// Source case carries an orphan reference: the parent shared step
			// no longer exists in the source project (deleted upstream) but the
			// case payload retains the shared_step_id and the inline expansion.
			// Fall back to inline content; aggregate the occurrence so the CLI
			// can surface a single concise note instead of one WARN per case.
			mu.Lock()
			m.logger.Debugw("Orphan shared_step_id in source case — importing as inline content",
				"case_title", caseData.Title, "unmapped_shared_step_id", orig.SharedStepID)
			mu.Unlock()
			m.unmappedMu.Lock()
			if m.unmappedSharedStepIDs == nil {
				m.unmappedSharedStepIDs = make(map[int64]struct{})
			}
			m.unmappedSharedStepIDs[orig.SharedStepID] = struct{}{}
			m.unmappedSharedStepHits++
			m.unmappedMu.Unlock()
		}
		out = append(out, data.Step{
			Content:        orig.Content,
			AdditionalInfo: orig.AdditionalInfo,
			Expected:       orig.Expected,
			Refs:           orig.Refs,
		})
		lastSharedID = 0
	}
	return out
}

// ImportCasesReport is like ImportCases but returns lists of created IDs and errors for CLI reporting.
//
// Like ImportCases, this restores source order within each destination section
// after the parallel import phase via move_cases_to_section.
func (m *Migration) ImportCasesReport(ctx context.Context, filtered data.GetCasesResponse, dryRun bool) (createdIDs []int64, errs []string, err error) {
	if dryRun || len(filtered) == 0 {
		m.logger.Infow("Dry-run or no data — cases import skipped", "count", len(filtered))
		return nil, nil, nil
	}

	m.logger.Infow("Starting cases import (report)", "count", len(filtered))

	sectionMap, err := m.resolveSectionMapByName(ctx)
	if err != nil {
		return nil, nil, err
	}

	autotestDefault, err := m.resolveAutotestOnDefault(ctx)
	if err != nil {
		return nil, nil, err
	}

	extraFields, err := m.resolveRequiredExtraFields(ctx)
	if err != nil {
		return nil, nil, err
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxImportConcurrency)
	perSection := make(map[int64][]createdCase)

	for i, c := range filtered {
		wg.Add(1)
		sem <- struct{}{}
		go func(srcOrder int, caseData data.Case) {
			defer func() { <-sem }()
			defer wg.Done()

			// Prepare request
			dstSectionID := m.resolveDestinationSectionID(caseData.SectionID, sectionMap)
			if dstSectionID == 0 {
				mu.Lock()
				m.failedImports++
				errs = append(errs, fmt.Sprintf("case %q: unable to resolve destination section_id", caseData.Title))
				m.logger.Errorw("Error importing case", "title", caseData.Title, "error", "unable to resolve destination section_id")
				mu.Unlock()
				return
			}

			req := &data.AddCaseRequest{
				Title:          caseData.Title,
				SectionID:      dstSectionID,
				TypeID:         caseData.TypeID,
				PriorityID:     caseData.PriorityID,
				TemplateID:     caseData.TemplateID,
				MilestoneID:    caseData.MilestoneID,
				Refs:           caseData.Refs,
				CustomPreconds: caseData.CustomPreconds,
			}
			applyAutotestOnValue(req, caseData.CustomAutotestOn, autotestDefault)
			req.ExtraFields = extraFields
			req.CustomStepsSeparated = m.buildImportSteps(caseData, &mu)

			// Create in target section
			created, err := m.Client.AddCase(ctx, dstSectionID, req)
			if err != nil {
				mu.Lock()
				m.failedImports++
				errs = append(errs, fmt.Sprintf("case %q: %v", caseData.Title, err))
				m.logger.Errorw("Error importing case", "title", caseData.Title, "error", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			createdIDs = append(createdIDs, created.ID)
			perSection[dstSectionID] = append(perSection[dstSectionID], createdCase{srcOrder: srcOrder, newID: created.ID})
			m.importedCases++
			m.logger.Infow("Successfully created case (report)", "old_id", caseData.ID, "new_id", created.ID, "title", caseData.Title)
			mu.Unlock()
		}(i, c)
	}
	wg.Wait()

	m.reorderCreatedCases(ctx, perSection)

	m.logger.Infow("Cases import (report) completed", "imported", m.importedCases)
	return createdIDs, errs, nil
}

func (m *Migration) resolveAutotestOnDefault(ctx context.Context) (*int64, error) {
	fields, err := m.Client.GetCaseFields(ctx)
	if err != nil {
		return nil, fmt.Errorf("get case fields: %w", err)
	}

	resolved, err := casefields.ResolveCustomAutotestOn(fields, m.dstProject)
	if err != nil {
		return nil, fmt.Errorf("resolve custom_autotest_on for destination project %d: %w", m.dstProject, err)
	}
	return resolved, nil
}

// resolveRequiredExtraFields returns a map of system_name→value for all DST project
// required custom fields (that have a configured default) except custom_autotest_on
// (which is handled separately). The map is used as AddCaseRequest.ExtraFields.
func (m *Migration) resolveRequiredExtraFields(ctx context.Context) (map[string]interface{}, error) {
	fields, err := m.Client.GetCaseFields(ctx)
	if err != nil {
		return nil, fmt.Errorf("get case fields for extra required fields: %w", err)
	}
	return casefields.ResolveRequiredCustomFields(fields, m.dstProject, []string{"custom_autotest_on"}), nil
}

func applyAutotestOnValue(req *data.AddCaseRequest, sourceValue int64, defaultValue *int64) {
	if sourceValue != 0 {
		value := sourceValue
		req.CustomAutotestOn = &value
		return
	}
	if defaultValue != nil {
		value := *defaultValue
		req.CustomAutotestOn = &value
	}
}

func (m *Migration) resolveSectionMapByName(ctx context.Context) (map[int64]int64, error) {
	sourceSections, err := m.Client.GetSections(ctx, m.srcProject, m.srcSuite)
	if err != nil {
		return nil, fmt.Errorf("get source sections: %w", err)
	}

	targetSections, err := m.Client.GetSections(ctx, m.dstProject, m.dstSuite)
	if err != nil {
		return nil, fmt.Errorf("get destination sections: %w", err)
	}

	targetByName := make(map[string]int64, len(targetSections))
	for _, section := range targetSections {
		if section.Name == "" {
			continue
		}
		if _, exists := targetByName[section.Name]; !exists {
			targetByName[section.Name] = section.ID
		}
	}

	sectionMap := make(map[int64]int64, len(sourceSections))
	for _, source := range sourceSections {
		if mapped, ok := m.mapping.GetTargetBySource(source.ID); ok {
			sectionMap[source.ID] = mapped
			continue
		}
		if targetID, ok := targetByName[source.Name]; ok {
			sectionMap[source.ID] = targetID
			// Register section resolution in the global mapping so downstream
			// consumers (e.g. coverage gate via resolveDstSectionIDForFilter)
			// see the same scope resolution as ImportCases. Without this the
			// post-import coverage verification treats every case as missing
			// because it only consults m.mapping.
			m.mapping.AddPair(source.ID, targetID, "existing")
		}
	}

	return sectionMap, nil
}

func (m *Migration) resolveDestinationSectionID(sourceSectionID int64, sectionMap map[int64]int64) int64 {
	if sourceSectionID == 0 {
		return m.dstSuite
	}

	if mapped, ok := m.mapping.GetTargetBySource(sourceSectionID); ok {
		return mapped
	}

	if mapped, ok := sectionMap[sourceSectionID]; ok {
		return mapped
	}

	return 0
}

// createdCase pairs a newly imported case ID with its original index in the
// source `filtered` slice, so that ImportCases/ImportCasesReport can restore
// per-section ordering after parallel creation.
type createdCase struct {
	srcOrder int
	newID    int64
}

// reorderCreatedCases restores the source-side ordering of cases inside each
// destination section by issuing move_cases_to_section with case IDs sorted by
// their original index in the source slice.
//
// Background: ImportCases creates cases concurrently (maxImportConcurrency=10)
// for throughput. TestRail records cases in the order add_case requests arrive,
// so concurrent goroutines produce a non-deterministic on-screen ordering that
// does not match the source suite. move_cases_to_section accepts the new order
// in a single call per section; cases stay in the same section (no cross-suite
// move), so the call is cheap and idempotent.
//
// Errors are logged but never abort the import: ordering is a UX concern, the
// underlying data is already migrated correctly. Sections with 0 or 1 created
// cases are skipped (no work to do).
func (m *Migration) reorderCreatedCases(ctx context.Context, perSection map[int64][]createdCase) {
	if len(perSection) == 0 {
		return
	}

	sectionIDs := make([]int64, 0, len(perSection))
	for sid := range perSection {
		sectionIDs = append(sectionIDs, sid)
	}
	sort.Slice(sectionIDs, func(i, j int) bool { return sectionIDs[i] < sectionIDs[j] })

	reordered := 0
	for _, sectionID := range sectionIDs {
		entries := perSection[sectionID]
		if len(entries) < 2 {
			continue
		}
		sort.SliceStable(entries, func(i, j int) bool {
			return entries[i].srcOrder < entries[j].srcOrder
		})
		ids := make([]int64, len(entries))
		for i, e := range entries {
			ids[i] = e.newID
		}
		if err := m.Client.MoveCasesToSection(ctx, sectionID, &data.MoveCasesRequest{CaseIDs: ids}); err != nil {
			m.logger.Warnw("Failed to restore case order in section",
				"section_id", sectionID, "case_count", len(ids), "error", err)
			continue
		}
		reordered++
		m.logger.Infow("Restored case order in section",
			"section_id", sectionID, "case_count", len(ids))
	}

	if reordered > 0 {
		m.logger.Infow("Case order restoration completed",
			"sections_reordered", reordered, "sections_total", len(sectionIDs))
	}
}
