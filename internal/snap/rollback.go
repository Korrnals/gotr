package snap

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/models/data"
)

// concurrentThreshold is the minimum entity count to trigger parallel processing.
const concurrentThreshold = 10

// defaultParallelism is the default number of concurrent workers.
const defaultParallelism = 4

// CasesAPI defines the API methods needed for rollback operations.
// client.ClientInterface satisfies this interface.
type CasesAPI interface {
	GetCase(ctx context.Context, caseID int64) (*data.Case, error)
	UpdateCase(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error)
	AddCase(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error)
	DeleteCase(ctx context.Context, caseID int64) error

	// Section operations for cascade rollback.
	GetSection(ctx context.Context, sectionID int64) (*data.Section, error)
	AddSection(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error)
	DeleteSection(ctx context.Context, sectionID int64) error

	// Sync rollback operations.
	DeleteSharedStep(ctx context.Context, stepID int64, keepInCases int) error
	DeleteSuite(ctx context.Context, suiteID int64) error
}

// RollbackOpts configures rollback behavior.
type RollbackOpts struct {
	// EntityIDs limits rollback to specific entity IDs (nil = all).
	EntityIDs []int64
	// DryRun previews changes without applying them.
	DryRun bool
}

// RollbackResult holds the outcome of a rollback operation.
type RollbackResult struct {
	SnapID      string
	Operation   Operation
	EntityType  string
	Success     bool
	NewEntityID int64  // non-zero if a new entity was created (e.g. delete rollback)
	Message     string
	DryRun      bool

	// Preview contains field-level diff entries (populated in dry-run or preview mode).
	Preview []DiffEntry
}

// DiffEntry represents a single field difference for rollback preview.
type DiffEntry struct {
	EntityID int64
	Field    string
	Current  string
	Saved    string
}

// Rollback reverses a mutation using the saved snapshot data.
// Accepts optional RollbackOpts for entity filtering and dry-run mode.
func Rollback(ctx context.Context, api CasesAPI, store *Store, manifest *Manifest, snapID string, opts ...RollbackOpts) (*RollbackResult, error) {
	var opt RollbackOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	meta, err := store.LoadMeta(snapID)
	if err != nil {
		return nil, fmt.Errorf("load snapshot: %w", err)
	}

	// Allow resume: rollback_partial snapshots can be retried.
	if meta.Status != StatusAvailable && meta.Status != StatusRollbackPartial {
		return nil, fmt.Errorf("snapshot %q status is %q, expected %q", snapID, meta.Status, StatusAvailable)
	}

	result := &RollbackResult{
		SnapID:     snapID,
		Operation:  meta.Operation,
		EntityType: meta.EntityType,
		DryRun:     opt.DryRun,
	}

	switch {
	case meta.EntityType == "case":
		err = rollbackCase(ctx, api, store, meta, result, opt)
	case meta.EntityType == "section":
		err = rollbackSectionCascade(ctx, api, store, meta, result, opt)
	case meta.IsSyncOp():
		err = rollbackSync(ctx, api, store, meta, result, opt)
	default:
		return nil, fmt.Errorf("unsupported entity type for rollback: %q", meta.EntityType)
	}

	if err != nil {
		meta.Status = StatusRollbackPartial
		_ = store.SaveMeta(meta)
		_ = manifest.UpdateStatus(snapID, StatusRollbackPartial)
		return result, err
	}

	if !opt.DryRun {
		result.Success = true
		meta.Status = StatusRolledBack
		_ = store.SaveMeta(meta)
		_ = manifest.UpdateStatus(snapID, StatusRolledBack)
	} else {
		result.Success = true
	}
	return result, nil
}

func rollbackCase(ctx context.Context, api CasesAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	switch meta.Operation {
	case OpUpdate:
		return rollbackCaseUpdate(ctx, api, store, meta, result, opt)
	case OpDelete:
		return rollbackCaseDelete(ctx, api, store, meta, result, opt)
	case OpAdd:
		return rollbackCaseAdd(ctx, api, store, meta, result, opt)
	default:
		return fmt.Errorf("unsupported operation for case rollback: %q", meta.Operation)
	}
}

// entityAllowed checks if a given entity ID is in the allowed set (empty = all allowed).
func entityAllowed(id int64, filter []int64) bool {
	if len(filter) == 0 {
		return true
	}
	for _, fid := range filter {
		if fid == id {
			return true
		}
	}
	return false
}

// logEntry finds or creates a RollbackLogEntry for the given entity in meta.
func logEntry(meta *Meta, entityType string, id int64) *RollbackLogEntry {
	for i := range meta.RollbackLog {
		if meta.RollbackLog[i].ID == id && meta.RollbackLog[i].Type == entityType {
			return &meta.RollbackLog[i]
		}
	}
	meta.RollbackLog = append(meta.RollbackLog, RollbackLogEntry{
		Type:   entityType,
		ID:     id,
		Status: RBPending,
	})
	return &meta.RollbackLog[len(meta.RollbackLog)-1]
}

// rollbackCaseUpdate restores the pre-update state of a case.
func rollbackCaseUpdate(ctx context.Context, api CasesAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	var saved data.Case
	if err := store.LoadData(meta.ID, meta.DataFile, &saved); err != nil {
		return fmt.Errorf("load saved case data: %w", err)
	}

	if len(meta.EntityIDs) == 0 {
		return fmt.Errorf("no entity ID in snapshot meta")
	}
	caseID := meta.EntityIDs[0]

	if !entityAllowed(caseID, opt.EntityIDs) {
		result.Message = fmt.Sprintf("Case %d skipped (not in --entity-ids filter)", caseID)
		result.Success = true
		return nil
	}

	entry := logEntry(meta, "case", caseID)
	if entry.Status == RBRestored {
		result.Message = fmt.Sprintf("Case %d already restored (resume skip)", caseID)
		result.Success = true
		return nil
	}

	// Build diff preview.
	if opt.DryRun {
		current, err := api.GetCase(ctx, caseID)
		if err != nil {
			return fmt.Errorf("get current case %d for preview: %w", caseID, err)
		}
		if current != nil {
			result.Preview = buildCaseDiff(caseID, current, &saved)
		}
		result.Message = fmt.Sprintf("Dry-run: Case %d would be restored to pre-update state", caseID)
		return nil
	}

	req := caseToUpdateRequest(&saved)
	_, err := api.UpdateCase(ctx, caseID, req)
	if err != nil {
		entry.Status = RBFailed
		entry.Error = err.Error()
		return fmt.Errorf("API update_case %d: %w", caseID, err)
	}

	entry.Status = RBRestored
	result.Message = fmt.Sprintf("Case %d restored to pre-update state", caseID)
	return nil
}

// rollbackCaseDelete re-creates a deleted case from snapshot data.
func rollbackCaseDelete(ctx context.Context, api CasesAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	var saved data.Case
	if err := store.LoadData(meta.ID, meta.DataFile, &saved); err != nil {
		return fmt.Errorf("load saved case data: %w", err)
	}

	if !entityAllowed(saved.ID, opt.EntityIDs) {
		result.Message = fmt.Sprintf("Case %d skipped (not in --entity-ids filter)", saved.ID)
		result.Success = true
		return nil
	}

	entry := logEntry(meta, "case", saved.ID)
	if entry.Status == RBRestored {
		result.Message = fmt.Sprintf("Case %d already restored (resume skip)", saved.ID)
		result.Success = true
		return nil
	}

	if opt.DryRun {
		result.Preview = buildCaseAddPreview(saved.ID, &saved)
		result.Message = fmt.Sprintf("Dry-run: Case %d would be re-created (Tier 2 — new ID)", saved.ID)
		return nil
	}

	req := caseToAddRequest(&saved)
	created, err := api.AddCase(ctx, saved.SectionID, req)
	if err != nil {
		entry.Status = RBFailed
		entry.Error = err.Error()
		return fmt.Errorf("API add_case (re-create): %w", err)
	}

	entry.Status = RBRestored
	result.NewEntityID = created.ID
	result.Message = fmt.Sprintf("Case re-created as ID %d (original: %d, Tier 2: new ID)", created.ID, saved.ID)
	return nil
}

// rollbackCaseAdd deletes a case that was created by an add operation.
func rollbackCaseAdd(ctx context.Context, api CasesAPI, _ *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	if len(meta.EntityIDs) == 0 {
		return fmt.Errorf("no created entity ID in snapshot meta (FinalizeAdd may not have been called)")
	}
	caseID := meta.EntityIDs[0]

	if !entityAllowed(caseID, opt.EntityIDs) {
		result.Message = fmt.Sprintf("Case %d skipped (not in --entity-ids filter)", caseID)
		result.Success = true
		return nil
	}

	entry := logEntry(meta, "case", caseID)
	if entry.Status == RBRestored {
		result.Message = fmt.Sprintf("Case %d already rolled back (resume skip)", caseID)
		result.Success = true
		return nil
	}

	if opt.DryRun {
		result.Preview = []DiffEntry{{EntityID: caseID, Field: "action", Current: "exists", Saved: "DELETE"}}
		result.Message = fmt.Sprintf("Dry-run: Case %d would be deleted (undo add)", caseID)
		return nil
	}

	if err := api.DeleteCase(ctx, caseID); err != nil {
		entry.Status = RBFailed
		entry.Error = err.Error()
		return fmt.Errorf("API delete_case %d: %w", caseID, err)
	}

	entry.Status = RBRestored
	result.Message = fmt.Sprintf("Case %d deleted (undo add)", caseID)
	return nil
}

// ---------------------------------------------------------------------------
// diff helpers
// ---------------------------------------------------------------------------

// buildCaseDiff compares current remote state with saved snapshot state.
func buildCaseDiff(caseID int64, current, saved *data.Case) []DiffEntry {
	var diffs []DiffEntry
	add := func(field, cur, sav string) {
		if cur != sav {
			diffs = append(diffs, DiffEntry{EntityID: caseID, Field: field, Current: cur, Saved: sav})
		}
	}

	add("title", current.Title, saved.Title)
	add("type_id", fmt.Sprintf("%d", current.TypeID), fmt.Sprintf("%d", saved.TypeID))
	add("priority_id", fmt.Sprintf("%d", current.PriorityID), fmt.Sprintf("%d", saved.PriorityID))
	add("estimate", current.Estimate, saved.Estimate)
	add("refs", current.Refs, saved.Refs)
	add("milestone_id", fmt.Sprintf("%d", current.MilestoneID), fmt.Sprintf("%d", saved.MilestoneID))
	add("template_id", fmt.Sprintf("%d", current.TemplateID), fmt.Sprintf("%d", saved.TemplateID))
	add("section_id", fmt.Sprintf("%d", current.SectionID), fmt.Sprintf("%d", saved.SectionID))
	add("custom_preconds", current.CustomPreconds, saved.CustomPreconds)
	add("custom_steps", current.CustomSteps, saved.CustomSteps)
	add("custom_expected", current.CustomExpected, saved.CustomExpected)

	return diffs
}

// buildCaseAddPreview shows fields that would be restored for a deleted case.
func buildCaseAddPreview(caseID int64, saved *data.Case) []DiffEntry {
	var diffs []DiffEntry
	diffs = append(diffs, DiffEntry{EntityID: caseID, Field: "action", Current: "DELETED", Saved: "RE-CREATE"})
	diffs = append(diffs, DiffEntry{EntityID: caseID, Field: "title", Current: "—", Saved: saved.Title})
	diffs = append(diffs, DiffEntry{EntityID: caseID, Field: "section_id", Current: "—", Saved: fmt.Sprintf("%d", saved.SectionID)})
	diffs = append(diffs, DiffEntry{EntityID: caseID, Field: "priority_id", Current: "—", Saved: fmt.Sprintf("%d", saved.PriorityID)})
	return diffs
}

// caseToUpdateRequest converts a saved Case to an UpdateCaseRequest.
func caseToUpdateRequest(c *data.Case) *data.UpdateCaseRequest {
	req := &data.UpdateCaseRequest{}
	if c.Title != "" {
		req.Title = &c.Title
	}
	if c.TypeID != 0 {
		req.TypeID = &c.TypeID
	}
	if c.PriorityID != 0 {
		req.PriorityID = &c.PriorityID
	}
	if c.Estimate != "" {
		req.Estimate = &c.Estimate
	}
	if c.CustomPreconds != "" {
		req.CustomPreconds = &c.CustomPreconds
	}
	if c.CustomSteps != "" {
		req.CustomSteps = &c.CustomSteps
	}
	if c.CustomExpected != "" {
		req.CustomExpected = &c.CustomExpected
	}
	if len(c.CustomStepsSeparated) > 0 {
		req.CustomStepsSeparated = c.CustomStepsSeparated
	}
	if c.Refs != "" {
		req.Refs = &c.Refs
	}
	if c.MilestoneID != 0 {
		req.MilestoneID = &c.MilestoneID
	}
	if c.TemplateID != 0 {
		req.TemplateID = &c.TemplateID
	}
	if c.SectionID != 0 {
		req.SectionID = &c.SectionID
	}
	return req
}

// caseToAddRequest converts a saved Case to an AddCaseRequest for re-creation.
func caseToAddRequest(c *data.Case) *data.AddCaseRequest {
	return &data.AddCaseRequest{
		Title:                c.Title,
		SectionID:            c.SectionID,
		TypeID:               c.TypeID,
		PriorityID:           c.PriorityID,
		Estimate:             c.Estimate,
		TemplateID:           c.TemplateID,
		Refs:                 c.Refs,
		MilestoneID:          c.MilestoneID,
		CustomPreconds:       c.CustomPreconds,
		CustomSteps:          c.CustomSteps,
		CustomExpected:       c.CustomExpected,
		CustomStepsSeparated: c.CustomStepsSeparated,
	}
}

// ---------------------------------------------------------------------------
// Cascade rollback: section delete → re-create section + child cases
// ---------------------------------------------------------------------------

// CascadeData stores section + child cases for cascade snapshots.
type CascadeData struct {
	Section data.Section `json:"section"`
	Cases   []data.Case  `json:"cases"`
}

// rollbackSectionCascade re-creates a deleted section and its child cases.
func rollbackSectionCascade(ctx context.Context, api CasesAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	if meta.Operation != OpDelete {
		return fmt.Errorf("section cascade rollback only supports delete, got %q", meta.Operation)
	}

	var cascade CascadeData
	if err := store.LoadData(meta.ID, meta.DataFile, &cascade); err != nil {
		return fmt.Errorf("load cascade data: %w", err)
	}

	sectionID := cascade.Section.ID

	// Dry-run: preview section + cases.
	if opt.DryRun {
		result.Preview = buildCascadePreview(&cascade, opt.EntityIDs)
		result.Message = fmt.Sprintf("Dry-run: Section %d + %d cases would be re-created", sectionID, len(cascade.Cases))
		return nil
	}

	sectionEntry := logEntry(meta, "section", sectionID)
	var newSectionID int64

	if sectionEntry.Status != RBRestored {
		req := &data.AddSectionRequest{
			Name:        cascade.Section.Name,
			Description: cascade.Section.Description,
			SuiteID:     cascade.Section.SuiteID,
			ParentID:    cascade.Section.ParentID,
		}

		created, err := api.AddSection(ctx, meta.ProjectID, req)
		if err != nil {
			sectionEntry.Status = RBFailed
			sectionEntry.Error = err.Error()
			return fmt.Errorf("API add_section (re-create %d): %w", sectionID, err)
		}
		sectionEntry.Status = RBRestored
		newSectionID = created.ID
	}

	// Re-create child cases in the new section.
	// Filter applicable cases first.
	var toRestore []data.Case
	for _, c := range cascade.Cases {
		if !entityAllowed(c.ID, opt.EntityIDs) {
			continue
		}
		entry := logEntry(meta, "case", c.ID)
		if entry.Status == RBRestored {
			continue
		}
		toRestore = append(toRestore, c)
	}

	type caseResult struct {
		CaseID    int64
		CreatedID int64
	}

	restoreOne := func(c data.Case, _ int) (caseResult, error) {
		req := caseToAddRequest(&c)
		if newSectionID > 0 {
			req.SectionID = newSectionID
		}
		created, err := api.AddCase(ctx, req.SectionID, req)
		if err != nil {
			return caseResult{CaseID: c.ID}, err
		}
		return caseResult{CaseID: c.ID, CreatedID: created.ID}, nil
	}

	var results []concurrent.Result[caseResult]
	if len(toRestore) >= concurrentThreshold {
		results, _ = concurrent.ParallelMap(ctx, toRestore, defaultParallelism, restoreOne)
	} else {
		for i, c := range toRestore {
			r, err := restoreOne(c, i)
			results = append(results, concurrent.Result[caseResult]{Data: r, Error: err, Index: i})
		}
	}

	// Update rollback log from results.
	restored := 0
	var lastErr error
	for _, r := range results {
		entry := logEntry(meta, "case", r.Data.CaseID)
		if r.Error != nil {
			entry.Status = RBFailed
			entry.Error = r.Error.Error()
			lastErr = r.Error
		} else {
			entry.Status = RBRestored
			restored++
		}
	}
	// Count pre-restored (skipped by resume).
	for _, c := range cascade.Cases {
		if entityAllowed(c.ID, opt.EntityIDs) {
			e := logEntry(meta, "case", c.ID)
			if e.Status == RBRestored {
				restored++
			}
		}
	}

	if newSectionID > 0 {
		result.NewEntityID = newSectionID
	}
	result.Message = fmt.Sprintf("Section re-created as ID %d, %d/%d cases restored", newSectionID, restored, len(cascade.Cases))

	if lastErr != nil {
		return fmt.Errorf("cascade rollback partial: %w", lastErr)
	}
	return nil
}

// buildCascadePreview builds diff entries for a section cascade rollback.
func buildCascadePreview(cascade *CascadeData, filterIDs []int64) []DiffEntry {
	var diffs []DiffEntry

	diffs = append(diffs, DiffEntry{
		EntityID: cascade.Section.ID,
		Field:    "action",
		Current:  "DELETED",
		Saved:    "RE-CREATE SECTION: " + cascade.Section.Name,
	})

	for _, c := range cascade.Cases {
		if !entityAllowed(c.ID, filterIDs) {
			continue
		}
		diffs = append(diffs, DiffEntry{
			EntityID: c.ID,
			Field:    "action",
			Current:  "DELETED",
			Saved:    "RE-CREATE CASE: " + c.Title,
		})
	}
	return diffs
}

// ---------------------------------------------------------------------------
// Sync rollback: undo sync by deleting all created entities
// ---------------------------------------------------------------------------

// SyncData stores the result of a sync operation for rollback.
type SyncData struct {
	SrcProject int64               `json:"src_project"`
	DstProject int64               `json:"dst_project"`
	SrcSuite   int64               `json:"src_suite"`
	DstSuite   int64               `json:"dst_suite"`
	Created    []SyncCreatedEntity `json:"created"`
}

// SyncCreatedEntity records a single entity created during sync.
type SyncCreatedEntity struct {
	Type     string `json:"type"` // "case", "section", "shared_step", "suite"
	SourceID int64  `json:"source_id"`
	TargetID int64  `json:"target_id"`
}

// rollbackSync deletes all entities created by a sync operation.
// Deletion order: cases → sections → shared_steps → suites (reverse dependency).
func rollbackSync(ctx context.Context, api CasesAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	var syncData SyncData
	if err := store.LoadData(meta.ID, meta.DataFile, &syncData); err != nil {
		return fmt.Errorf("load sync data: %w", err)
	}

	if opt.DryRun {
		result.Preview = buildSyncPreview(&syncData, opt.EntityIDs)
		result.Message = fmt.Sprintf("Dry-run: %d created entities would be deleted", len(syncData.Created))
		return nil
	}

	// Group by type for ordered deletion.
	deleteOrder := []string{"case", "section", "shared_step", "suite"}
	byType := make(map[string][]SyncCreatedEntity)
	for _, e := range syncData.Created {
		byType[e.Type] = append(byType[e.Type], e)
	}

	deleted := 0
	var lastErr error

	for _, typ := range deleteOrder {
		entities := byType[typ]
		for _, e := range entities {
			if !entityAllowed(e.TargetID, opt.EntityIDs) {
				continue
			}

			entry := logEntry(meta, typ, e.TargetID)
			if entry.Status == RBRestored {
				deleted++
				continue
			}

			err := deleteSyncEntity(ctx, api, typ, e.TargetID)
			if err != nil {
				entry.Status = RBFailed
				entry.Error = err.Error()
				lastErr = err
				continue
			}
			entry.Status = RBRestored
			deleted++
		}
	}

	result.Message = fmt.Sprintf("Sync rollback: %d/%d entities deleted", deleted, len(syncData.Created))

	if lastErr != nil {
		return fmt.Errorf("sync rollback partial: %w", lastErr)
	}
	return nil
}

// deleteSyncEntity performs deletion based on entity type.
func deleteSyncEntity(ctx context.Context, api CasesAPI, entityType string, targetID int64) error {
	switch entityType {
	case "case":
		return api.DeleteCase(ctx, targetID)
	case "section":
		return api.DeleteSection(ctx, targetID)
	case "shared_step":
		return api.DeleteSharedStep(ctx, targetID, 0)
	case "suite":
		return api.DeleteSuite(ctx, targetID)
	default:
		return fmt.Errorf("unknown sync entity type: %q", entityType)
	}
}

// buildSyncPreview builds diff entries for sync rollback.
func buildSyncPreview(sd *SyncData, filterIDs []int64) []DiffEntry {
	var diffs []DiffEntry
	for _, e := range sd.Created {
		if !entityAllowed(e.TargetID, filterIDs) {
			continue
		}
		diffs = append(diffs, DiffEntry{
			EntityID: e.TargetID,
			Field:    "action",
			Current:  fmt.Sprintf("%s (synced from %d)", e.Type, e.SourceID),
			Saved:    "DELETE",
		})
	}
	return diffs
}
