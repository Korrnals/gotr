package snap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/log"
	"github.com/Korrnals/gotr/internal/models/data"
	"go.uber.org/zap"
)

// errUnsupportedEntity signals that the entity type has no rollback handler.
// It is never wrapped, so callers can use errors.Is.
var errUnsupportedEntity = errors.New("unsupported entity type for rollback")

// concurrentThreshold is the minimum entity count to trigger parallel processing.
const concurrentThreshold = 10

// defaultParallelism is the default number of concurrent workers.
const defaultParallelism = 4

// RollbackAPI defines the API methods needed for rollback operations.
// client.ClientInterface satisfies this interface.
type RollbackAPI interface {
	// Case operations.
	GetCase(ctx context.Context, caseID int64) (*data.Case, error)
	UpdateCase(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error)
	AddCase(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error)
	DeleteCase(ctx context.Context, caseID int64) error

	// Section operations.
	GetSection(ctx context.Context, sectionID int64) (*data.Section, error)
	AddSection(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error)
	DeleteSection(ctx context.Context, sectionID int64) error

	// Suite operations.
	AddSuite(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error)
	DeleteSuite(ctx context.Context, suiteID int64) error

	// Shared step operations.
	DeleteSharedStep(ctx context.Context, stepID int64, keepInCases int) error

	// Run operations.
	AddRun(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error)
	DeleteRun(ctx context.Context, runID int64) error

	// Milestone operations.
	AddMilestone(ctx context.Context, projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error)
	DeleteMilestone(ctx context.Context, milestoneID int64) error

	// Plan operations.
	AddPlan(ctx context.Context, projectID int64, req *data.AddPlanRequest) (*data.Plan, error)
	DeletePlan(ctx context.Context, planID int64) error

	// Project operations.
	AddProject(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error)
	DeleteProject(ctx context.Context, projectID int64) error

	// Configuration operations.
	AddConfigGroup(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error)
	AddConfig(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error)
	DeleteConfigGroup(ctx context.Context, groupID int64) error
	DeleteConfig(ctx context.Context, configID int64) error

	// Group operations.
	AddGroup(ctx context.Context, projectID int64, name string, userIDs []int64) (*data.Group, error)
	DeleteGroup(ctx context.Context, groupID int64) error

	// Dataset operations.
	AddDataset(ctx context.Context, projectID int64, name string) (*data.Dataset, error)
	DeleteDataset(ctx context.Context, datasetID int64) error

	// Variable operations.
	AddVariable(ctx context.Context, datasetID int64, name string) (*data.Variable, error)
	DeleteVariable(ctx context.Context, variableID int64) error

	// Attachment operations (used by cleanup-attachments rollback).
	DownloadAttachment(ctx context.Context, attachmentID int64) (io.ReadCloser, error)
	AddAttachmentToCase(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlan(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlanEntry(ctx context.Context, planID int64, entryID, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToResult(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToRun(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error)
}

// CasesAPI is an alias for backward compatibility.
//
// Deprecated: use RollbackAPI.
type CasesAPI = RollbackAPI

// RollbackOpts configures rollback behavior.
type RollbackOpts struct {
	// EntityIDs limits rollback to specific entity IDs (nil = all).
	EntityIDs []int64
	// DryRun previews changes without applying them.
	DryRun bool
	// SkipReferences disables markdown-reference rewrite during the
	// attachments-cleanup rollback. References indexed in
	// references.json are left untouched on TestRail and the restore
	// report records them as not-rewritten.
	SkipReferences bool
	// RewriteAPI carries the optional surface required to GET/UPDATE
	// case/run/plan/milestone bodies for reference rewrite. When nil,
	// the rewrite phase is skipped (regardless of SkipReferences). The
	// CLI passes *client.HTTPClient which already implements every
	// Update* method on this interface.
	RewriteAPI ReferenceRewriteAPI
	// VerifyIntegrity enables a SHA-256 round-trip check against
	// integrity.json before any restore call. A mismatch is logged
	// as a warning and does not abort the rollback (operators can
	// still recover even if one binary is corrupt).
	VerifyIntegrity bool
}

// RollbackResult holds the outcome of a rollback operation.
type RollbackResult struct {
	SnapID      string
	Operation   Operation
	EntityType  string
	Success     bool
	NewEntityID int64 // non-zero if a new entity was created (e.g. delete rollback)
	Message     string
	DryRun      bool

	// Preview contains field-level diff entries (populated in dry-run or preview mode).
	Preview []DiffEntry

	// Stats holds detailed sync rollback statistics (populated for sync rollbacks).
	Stats *SyncRollbackStats
}

// SyncRollbackStats provides per-type breakdown of sync rollback outcomes.
// Deleted: removed by this run. Skipped: already absent on server. Failed: hard errors.
// PreRestored: previously restored entries skipped on resume.
type SyncRollbackStats struct {
	Total       int
	Deleted     int
	Skipped     int
	Failed      int
	PreRestored int
	ByType      map[string]*SyncRollbackTypeStats
	Failures    []SyncRollbackFailure
}

// SyncRollbackTypeStats holds per-type counters.
type SyncRollbackTypeStats struct {
	Total       int
	Deleted     int
	Skipped     int
	Failed      int
	PreRestored int
}

// SyncRollbackFailure records a single failed entity for reporting.
type SyncRollbackFailure struct {
	Type     string
	TargetID int64
	Error    string
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

	err = dispatchRollback(ctx, api, store, meta, result, opt)
	if errors.Is(err, errUnsupportedEntity) {
		return nil, err
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

// dispatchRollback routes rollback to the correct entity handler.
// Returns errUnsupportedEntity (unwrapped) when no handler exists for the entity type.
func dispatchRollback(ctx context.Context, api CasesAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	switch meta.EntityType {
	case "case":
		return rollbackCase(ctx, api, store, meta, result, opt)
	case "section":
		return rollbackSection(ctx, api, store, meta, result, opt)
	case "project":
		return rollbackProject(ctx, api, store, meta, result, opt)
	case "run", "milestone", "plan", "suite", "group", "variable", "dataset", "configuration":
		return rollbackSimpleEntity(ctx, api, store, meta, result, opt)
	case EntityTypeAttachments:
		return rollbackAttachmentsCleanup(ctx, api, store, meta, result, opt)
	}
	if meta.IsSyncOp() {
		return rollbackSync(ctx, api, store, meta, result, opt)
	}
	return fmt.Errorf("%w: %q", errUnsupportedEntity, meta.EntityType)
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

// isGoneError returns true if the error indicates entity was already deleted (400/404).
func isGoneError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "400") || strings.Contains(msg, "404") ||
		strings.Contains(msg, "not found") || strings.Contains(msg, "no longer exists")
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
		if isGoneError(err) {
			entry.Status = RBFailed
			entry.Error = fmt.Sprintf("section %d not found, cannot re-create case", saved.SectionID)
			result.Message = fmt.Sprintf("Case %d: section %d no longer exists, skipping re-create", saved.ID, saved.SectionID)
			return nil
		}
		entry.Status = RBFailed
		entry.Error = err.Error()
		return fmt.Errorf("API add_case (re-create): %w", err)
	}

	entry.Status = RBRestored
	entry.NewID = created.ID
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
		if isGoneError(err) {
			entry.Status = RBRestored
			result.Message = fmt.Sprintf("Case %d already deleted, skipping (undo add)", caseID)
			return nil
		}
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
	return []DiffEntry{
		{EntityID: caseID, Field: "action", Current: "DELETED", Saved: "RE-CREATE"},
		{EntityID: caseID, Field: "title", Current: "—", Saved: saved.Title},
		{EntityID: caseID, Field: "section_id", Current: "—", Saved: fmt.Sprintf("%d", saved.SectionID)},
		{EntityID: caseID, Field: "priority_id", Current: "—", Saved: fmt.Sprintf("%d", saved.PriorityID)},
	}
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
// Section rollback: routes by operation
// ---------------------------------------------------------------------------

// rollbackSection routes section rollback by operation.
func rollbackSection(ctx context.Context, api RollbackAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	switch meta.Operation {
	case OpDelete:
		return rollbackSectionCascade(ctx, api, store, meta, result, opt)
	case OpAdd, OpCopy:
		return rollbackSectionAdd(ctx, api, meta, result, opt)
	default:
		return fmt.Errorf("unsupported operation for section rollback: %q", meta.Operation)
	}
}

// rollbackSectionAdd deletes a section that was created by an add/copy operation.
func rollbackSectionAdd(ctx context.Context, api RollbackAPI, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	if len(meta.EntityIDs) == 0 {
		return fmt.Errorf("no created entity ID in snapshot meta")
	}
	sectionID := meta.EntityIDs[0]

	if !entityAllowed(sectionID, opt.EntityIDs) {
		result.Message = fmt.Sprintf("Section %d skipped (not in --entity-ids filter)", sectionID)
		result.Success = true
		return nil
	}

	entry := logEntry(meta, "section", sectionID)
	if entry.Status == RBRestored {
		result.Message = fmt.Sprintf("Section %d already rolled back (resume skip)", sectionID)
		result.Success = true
		return nil
	}

	if opt.DryRun {
		result.Preview = []DiffEntry{{EntityID: sectionID, Field: "action", Current: "exists", Saved: "DELETE"}}
		result.Message = fmt.Sprintf("Dry-run: Section %d would be deleted (undo %s)", sectionID, meta.Operation)
		return nil
	}

	if err := api.DeleteSection(ctx, sectionID); err != nil {
		if isGoneError(err) {
			entry.Status = RBRestored
			result.Message = fmt.Sprintf("Section %d already deleted (undo %s)", sectionID, meta.Operation)
			return nil
		}
		entry.Status = RBFailed
		entry.Error = err.Error()
		return fmt.Errorf("API delete_section %d: %w", sectionID, err)
	}

	entry.Status = RBRestored
	result.Message = fmt.Sprintf("Section %d deleted (undo %s)", sectionID, meta.Operation)
	return nil
}

// ---------------------------------------------------------------------------
// Project rollback
// ---------------------------------------------------------------------------

// rollbackProject handles project-level rollback.
func rollbackProject(ctx context.Context, api RollbackAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	switch meta.Operation {
	case OpDelete:
		return rollbackProjectDelete(ctx, api, store, meta, result, opt)
	case OpAdd:
		return rollbackProjectAdd(ctx, api, meta, result, opt)
	default:
		return fmt.Errorf("unsupported operation for project rollback: %q", meta.Operation)
	}
}

// ProjectData stores project state for rollback of deletion.
type ProjectData struct {
	Project data.Project `json:"project"`
}

// rollbackProjectDelete re-creates a deleted project from snapshot data.
func rollbackProjectDelete(ctx context.Context, api RollbackAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	var saved ProjectData
	if err := store.LoadData(meta.ID, meta.DataFile, &saved); err != nil {
		return fmt.Errorf("load project data: %w", err)
	}

	projectID := saved.Project.ID

	entry := logEntry(meta, "project", projectID)
	if entry.Status == RBRestored {
		result.Message = fmt.Sprintf("Project %d already restored (resume skip)", projectID)
		result.Success = true
		return nil
	}

	if opt.DryRun {
		result.Preview = []DiffEntry{
			{EntityID: projectID, Field: "action", Current: "DELETED", Saved: "RE-CREATE"},
			{EntityID: projectID, Field: "name", Current: "—", Saved: saved.Project.Name},
			{EntityID: projectID, Field: "suite_mode", Current: "—", Saved: fmt.Sprintf("%d", saved.Project.SuiteMode)},
		}
		result.Message = fmt.Sprintf("Dry-run: Project %d (%s) would be re-created (Tier 2 — new ID)", projectID, saved.Project.Name)
		return nil
	}

	req := &data.AddProjectRequest{
		Name:             saved.Project.Name,
		Announcement:     saved.Project.Announcement,
		ShowAnnouncement: saved.Project.ShowAnnouncement,
		SuiteMode:        saved.Project.SuiteMode,
	}

	created, err := api.AddProject(ctx, req)
	if err != nil {
		entry.Status = RBFailed
		entry.Error = err.Error()
		return fmt.Errorf("API add_project (re-create %d): %w", projectID, err)
	}

	entry.Status = RBRestored
	entry.NewID = created.ID
	result.NewEntityID = created.ID
	result.Message = fmt.Sprintf("Project re-created as ID %d (original: %d, Tier 2: new ID)", created.ID, projectID)
	return nil
}

// rollbackProjectAdd deletes a project that was created by an add operation.
func rollbackProjectAdd(ctx context.Context, api RollbackAPI, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	if len(meta.EntityIDs) == 0 {
		return fmt.Errorf("no created entity ID in snapshot meta")
	}
	projectID := meta.EntityIDs[0]

	entry := logEntry(meta, "project", projectID)
	if entry.Status == RBRestored {
		result.Message = fmt.Sprintf("Project %d already rolled back (resume skip)", projectID)
		result.Success = true
		return nil
	}

	if opt.DryRun {
		result.Preview = []DiffEntry{{EntityID: projectID, Field: "action", Current: "exists", Saved: "DELETE"}}
		result.Message = fmt.Sprintf("Dry-run: Project %d would be deleted (undo add)", projectID)
		return nil
	}

	if err := api.DeleteProject(ctx, projectID); err != nil {
		if isGoneError(err) {
			entry.Status = RBRestored
			result.Message = fmt.Sprintf("Project %d already deleted (undo add)", projectID)
			return nil
		}
		entry.Status = RBFailed
		entry.Error = err.Error()
		return fmt.Errorf("API delete_project %d: %w", projectID, err)
	}

	entry.Status = RBRestored
	result.Message = fmt.Sprintf("Project %d deleted (undo add)", projectID)
	return nil
}

// ---------------------------------------------------------------------------
// Generic simple entity rollback (add → delete)
// Covers: run, milestone, plan, suite, group, variable, dataset, configuration
// ---------------------------------------------------------------------------

// rollbackSimpleEntity handles undo-add for entities with a simple delete API.
func rollbackSimpleEntity(ctx context.Context, api RollbackAPI, _ *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	if meta.Operation != OpAdd {
		return fmt.Errorf("unsupported operation for %s rollback: %q (only add is supported)", meta.EntityType, meta.Operation)
	}

	if len(meta.EntityIDs) == 0 {
		return fmt.Errorf("no created entity ID in snapshot meta for %s", meta.EntityType)
	}
	entityID := meta.EntityIDs[0]

	if !entityAllowed(entityID, opt.EntityIDs) {
		result.Message = fmt.Sprintf("%s %d skipped (not in --entity-ids filter)", meta.EntityType, entityID)
		result.Success = true
		return nil
	}

	entry := logEntry(meta, meta.EntityType, entityID)
	if entry.Status == RBRestored {
		result.Message = fmt.Sprintf("%s %d already rolled back (resume skip)", meta.EntityType, entityID)
		result.Success = true
		return nil
	}

	if opt.DryRun {
		result.Preview = []DiffEntry{{EntityID: entityID, Field: "action", Current: "exists", Saved: "DELETE"}}
		result.Message = fmt.Sprintf("Dry-run: %s %d would be deleted (undo add)", meta.EntityType, entityID)
		return nil
	}

	deleteFn, err := resolveDeleteFn(api, meta.EntityType)
	if err != nil {
		return fmt.Errorf("rollbackSimpleEntity: %w", err)
	}

	if err := deleteFn(ctx, entityID); err != nil {
		if isGoneError(err) {
			entry.Status = RBRestored
			result.Message = fmt.Sprintf("%s %d already deleted (undo add)", meta.EntityType, entityID)
			return nil
		}
		entry.Status = RBFailed
		entry.Error = err.Error()
		return fmt.Errorf("API delete_%s %d: %w", meta.EntityType, entityID, err)
	}

	entry.Status = RBRestored
	result.Message = fmt.Sprintf("%s %d deleted (undo add)", meta.EntityType, entityID)
	return nil
}

// deleteFn is a function that deletes an entity by ID.
type deleteFn func(ctx context.Context, id int64) error

// resolveDeleteFn returns the appropriate delete function for the entity type.
func resolveDeleteFn(api RollbackAPI, entityType string) (deleteFn, error) {
	switch entityType {
	case "run":
		return api.DeleteRun, nil
	case "milestone":
		return api.DeleteMilestone, nil
	case "plan":
		return api.DeletePlan, nil
	case "suite":
		return api.DeleteSuite, nil
	case "group":
		return api.DeleteGroup, nil
	case "variable":
		return api.DeleteVariable, nil
	case "dataset":
		return api.DeleteDataset, nil
	case "configuration":
		return func(ctx context.Context, id int64) error {
			// configuration snapshots may be config or config_group;
			// try config_group first (it cascades).
			err := api.DeleteConfigGroup(ctx, id)
			if err != nil && isGoneError(err) {
				return api.DeleteConfig(ctx, id)
			}
			return err
		}, nil
	default:
		return nil, fmt.Errorf("no delete handler for entity type %q", entityType)
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
func rollbackSectionCascade(ctx context.Context, api RollbackAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {

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

	newSectionID, err := restoreCascadeSection(ctx, api, meta, cascade.Section)
	if err != nil {
		return err
	}

	toRestore := filterCasesForCascade(meta, cascade.Cases, opt.EntityIDs)
	restored, lastErr := runCascadeCases(ctx, api, meta, toRestore, newSectionID)
	restored += countPreRestored(meta, cascade.Cases, opt.EntityIDs)

	if newSectionID > 0 {
		result.NewEntityID = newSectionID
	}
	result.Message = fmt.Sprintf("Section re-created as ID %d, %d/%d cases restored", newSectionID, restored, len(cascade.Cases))

	if lastErr != nil {
		return fmt.Errorf("cascade rollback partial: %w", lastErr)
	}
	return nil
}

// restoreCascadeSection creates a section from snapshot data (skips if already restored).
// Returns the new section ID (0 if the section was already restored earlier).
func restoreCascadeSection(ctx context.Context, api RollbackAPI, meta *Meta, section data.Section) (int64, error) {
	sectionEntry := logEntry(meta, "section", section.ID)
	if sectionEntry.Status == RBRestored {
		return 0, nil
	}

	req := &data.AddSectionRequest{
		Name:        section.Name,
		Description: section.Description,
		SuiteID:     section.SuiteID,
		ParentID:    section.ParentID,
	}
	created, err := api.AddSection(ctx, meta.ProjectID, req)
	if err != nil {
		sectionEntry.Status = RBFailed
		sectionEntry.Error = err.Error()
		return 0, fmt.Errorf("API add_section (re-create %d): %w", section.ID, err)
	}
	sectionEntry.Status = RBRestored
	sectionEntry.NewID = created.ID
	return created.ID, nil
}

// filterCasesForCascade returns cases that are allowed and not yet restored.
func filterCasesForCascade(meta *Meta, cases []data.Case, filterIDs []int64) []data.Case {
	out := make([]data.Case, 0, len(cases))
	for _, c := range cases {
		if !entityAllowed(c.ID, filterIDs) {
			continue
		}
		if logEntry(meta, "case", c.ID).Status == RBRestored {
			continue
		}
		out = append(out, c)
	}
	return out
}

// cascadeCaseResult holds the outcome of a single case re-create attempt.
type cascadeCaseResult struct {
	CaseID    int64
	CreatedID int64
}

// runCascadeCases re-creates cases in the new section, updating the rollback log.
// Returns (number restored, last error encountered).
func runCascadeCases(ctx context.Context, api RollbackAPI, meta *Meta, toRestore []data.Case, newSectionID int64) (int, error) {
	restoreOne := func(c data.Case, _ int) (cascadeCaseResult, error) {
		req := caseToAddRequest(&c)
		if newSectionID > 0 {
			req.SectionID = newSectionID
		}
		created, err := api.AddCase(ctx, req.SectionID, req)
		if err != nil {
			return cascadeCaseResult{CaseID: c.ID}, err
		}
		return cascadeCaseResult{CaseID: c.ID, CreatedID: created.ID}, nil
	}

	var results []concurrent.Result[cascadeCaseResult]
	if len(toRestore) >= concurrentThreshold {
		results, _ = concurrent.ParallelMap(ctx, toRestore, defaultParallelism, restoreOne)
	} else {
		for i, c := range toRestore {
			r, err := restoreOne(c, i)
			results = append(results, concurrent.Result[cascadeCaseResult]{Data: r, Error: err, Index: i})
		}
	}

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
			entry.NewID = r.Data.CreatedID
			restored++
		}
	}
	return restored, lastErr
}

// countPreRestored counts cases that were already restored (resume/skip).
func countPreRestored(meta *Meta, cases []data.Case, filterIDs []int64) int {
	n := 0
	for _, c := range cases {
		if entityAllowed(c.ID, filterIDs) && logEntry(meta, "case", c.ID).Status == RBRestored {
			n++
		}
	}
	return n
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

	stats := &SyncRollbackStats{
		ByType: make(map[string]*SyncRollbackTypeStats),
	}
	bumpType := func(typ string) *SyncRollbackTypeStats {
		t, ok := stats.ByType[typ]
		if !ok {
			t = &SyncRollbackTypeStats{}
			stats.ByType[typ] = t
		}
		return t
	}

	var lastErr error

	for _, typ := range deleteOrder {
		entities := byType[typ]
		for _, e := range entities {
			if !entityAllowed(e.TargetID, opt.EntityIDs) {
				continue
			}

			ts := bumpType(typ)
			ts.Total++
			stats.Total++

			entry := logEntry(meta, typ, e.TargetID)
			if entry.Status == RBRestored {
				ts.PreRestored++
				stats.PreRestored++
				continue
			}
			if entry.Status == RBSkipped {
				ts.Skipped++
				stats.Skipped++
				continue
			}

			err := deleteSyncEntity(ctx, api, typ, e.TargetID)
			switch {
			case err == nil:
				entry.Status = RBRestored
				entry.Error = ""
				ts.Deleted++
				stats.Deleted++
			case isAlreadyDeletedErr(err):
				entry.Status = RBSkipped
				entry.Error = ""
				ts.Skipped++
				stats.Skipped++
			default:
				entry.Status = RBFailed
				entry.Error = err.Error()
				ts.Failed++
				stats.Failed++
				stats.Failures = append(stats.Failures, SyncRollbackFailure{
					Type:     typ,
					TargetID: e.TargetID,
					Error:    err.Error(),
				})
				lastErr = err
			}
		}
	}

	result.Stats = stats
	result.Message = formatSyncRollbackMessage(stats)

	if lastErr != nil {
		return fmt.Errorf("sync rollback partial: %w", lastErr)
	}
	return nil
}

// isAlreadyDeletedErr returns true when the API rejects deletion because the
// entity does not exist anymore. TestRail signals this with HTTP 400 and a
// message like "is not a valid <entity>" / "не является допустимым". Treating
// these as a successful skip lets rollback be idempotent across overlapping
// snapshots.
func isAlreadyDeletedErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "400") {
		return false
	}
	switch {
	case strings.Contains(msg, "is not a valid"),
		strings.Contains(msg, "не является допустимым"),
		strings.Contains(msg, "is not a valid test case"),
		strings.Contains(msg, "is not a valid shared test step"),
		strings.Contains(msg, "is not a valid section"),
		strings.Contains(msg, "is not a valid test suite"):
		return true
	}
	return false
}

// formatSyncRollbackMessage produces a human-readable summary line including a
// per-type breakdown, e.g.:
//
//	"Sync rollback: 12 deleted, 1 skipped, 0 failed (cases: 12 deleted; shared_steps: 1 skipped)".
func formatSyncRollbackMessage(stats *SyncRollbackStats) string {
	if stats == nil || stats.Total == 0 {
		return "Sync rollback: nothing to do"
	}
	parts := []string{
		fmt.Sprintf("%d deleted", stats.Deleted),
		fmt.Sprintf("%d skipped (already absent)", stats.Skipped),
		fmt.Sprintf("%d failed", stats.Failed),
	}
	if stats.PreRestored > 0 {
		parts = append(parts, fmt.Sprintf("%d already rolled back", stats.PreRestored))
	}
	header := fmt.Sprintf("Sync rollback: %s of %d total", strings.Join(parts, ", "), stats.Total)

	typeOrder := []string{"case", "section", "shared_step", "suite"}
	var details []string
	for _, t := range typeOrder {
		ts, ok := stats.ByType[t]
		if !ok || ts.Total == 0 {
			continue
		}
		var sub []string
		if ts.Deleted > 0 {
			sub = append(sub, fmt.Sprintf("%d deleted", ts.Deleted))
		}
		if ts.Skipped > 0 {
			sub = append(sub, fmt.Sprintf("%d skipped", ts.Skipped))
		}
		if ts.Failed > 0 {
			sub = append(sub, fmt.Sprintf("%d failed", ts.Failed))
		}
		if ts.PreRestored > 0 {
			sub = append(sub, fmt.Sprintf("%d prior", ts.PreRestored))
		}
		if len(sub) == 0 {
			continue
		}
		details = append(details, fmt.Sprintf("%ss: %s", t, strings.Join(sub, ", ")))
	}
	if len(details) > 0 {
		header += " (" + strings.Join(details, "; ") + ")"
	}
	return header
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

// ---------------------------------------------------------------------------
// Attachments cleanup rollback (re-upload deleted attachments)
// ---------------------------------------------------------------------------

// rollbackAttachmentsCleanup re-uploads every attachment recorded in the
// snapshot's data.json. The TestRail API assigns a new ID to each
// re-upload, so the per-entry mapping is recorded in the rollback log.
// Attachments bound to a "test" entity cannot be restored — TestRail has
// no add_attachment_to_test endpoint — and are reported as skipped.
func rollbackAttachmentsCleanup(ctx context.Context, api RollbackAPI, store *Store, meta *Meta, result *RollbackResult, opt RollbackOpts) error {
	if meta.Operation != OpDelete {
		return fmt.Errorf("unsupported operation for attachments rollback: %q (only delete is supported)", meta.Operation)
	}

	// 1. Optional pre-flight integrity check. Best-effort: a missing
	// integrity.json (legacy v1 snapshot) is fine; a mismatch is a
	// warning, not a hard stop.
	if opt.VerifyIntegrity {
		if err := VerifyIntegrityIndex(store, meta.ID); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Warn("integrity verify failed; continuing with restore",
				zap.String("snap_id", meta.ID), zap.Error(err))
		}
	}

	rb, err := RestoreCleanupAttachments(ctx, api, store, meta.ID, opt.DryRun)
	if err != nil {
		return fmt.Errorf("rollbackAttachmentsCleanup: %w", err)
	}

	for oldID, newID := range rb.Mapping {
		entry := logEntry(meta, "attachment", oldID)
		entry.Status = RBRestored
		entry.NewID = newID
	}
	for _, f := range rb.Failures {
		entry := logEntry(meta, "attachment", f.OriginalID)
		if errors.Is(errors.New(f.Error), ErrCleanupRollbackUnsupportedEntity) || strings.Contains(f.Error, ErrCleanupRollbackUnsupportedEntity.Error()) {
			entry.Status = RBSkipped
		} else {
			entry.Status = RBFailed
		}
		entry.Error = f.Error
	}

	// 2. Reference rewrite phase. Skipped on dry-run, when explicitly
	// disabled, when no rewrite API is wired, or when references.json
	// is absent (legacy v1 snapshot or empty index).
	var rwSummary string
	if !opt.DryRun && !opt.SkipReferences && opt.RewriteAPI != nil && len(rb.Mapping) > 0 {
		entries, err := LoadReferencesSidecar(store, meta.ID)
		switch {
		case err != nil && errors.Is(err, os.ErrNotExist):
			// legacy snapshot or scan was skipped: leave references untouched.
		case err != nil:
			log.Warn("references.json unreadable; skipping rewrite",
				zap.String("snap_id", meta.ID), zap.Error(err))
		case len(entries) == 0:
			// scanned but found nothing — nothing to do.
		default:
			rw, rerr := RewriteReferences(ctx, opt.RewriteAPI, entries, rb.Mapping)
			if rerr != nil {
				log.Warn("reference rewrite aborted",
					zap.String("snap_id", meta.ID), zap.Error(rerr))
			}
			if rw != nil {
				rwSummary = fmt.Sprintf(", rewrote %d references across %d entities (skipped %d, failed %d)",
					rw.RefsRewritten, rw.EntitiesRewritten, rw.RefsSkipped, rw.EntitiesFailed)
				for _, f := range rw.Failures {
					e := logEntry(meta, f.EntityType, f.EntityID)
					e.Status = RBFailed
					e.Error = f.Error
				}
			}
		}
	}

	if opt.DryRun {
		result.Message = fmt.Sprintf("Dry-run: %d attachments would be re-uploaded", rb.Restored)
		return nil
	}

	switch {
	case rb.Failed == 0 && rb.Skipped == 0:
		result.Success = true
		result.Message = fmt.Sprintf("Restored %d attachments (new IDs assigned by TestRail)%s", rb.Restored, rwSummary)
	case rb.Failed == 0:
		result.Success = true
		result.Message = fmt.Sprintf("Restored %d, skipped %d (entity type without add API)%s", rb.Restored, rb.Skipped, rwSummary)
	default:
		result.Message = fmt.Sprintf("Restored %d, skipped %d, failed %d%s", rb.Restored, rb.Skipped, rb.Failed, rwSummary)
		return fmt.Errorf("attachments rollback completed with %d failures", rb.Failed)
	}
	return nil
}
