// Package cleanup implements bulk-cleanup workflows for TestRail data
// kept reversible via the snap engine. The first supported workflow is
// attachments cleanup: filter by age and parent entity type, snapshot
// the binaries, then delete in parallel.
package cleanup

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/snap"
)

// AttachmentsAPI is the slice of the TestRail client surface used by
// the attachment-cleanup walker and executor. It is satisfied by
// *client.HTTPClient.
type AttachmentsAPI interface {
	GetProjects(ctx context.Context) (data.GetProjectsResponse, error)
	GetAttachmentsForProject(ctx context.Context, projectID int64) (data.GetAttachmentsResponse, error)
	DeleteAttachment(ctx context.Context, attachmentID int64) error
	snap.CleanupAttachmentsAPI
	// ReferenceFetchAPI is required so Execute can index markdown
	// references that point at the about-to-be-deleted attachments.
	// Disabled when ExecuteOptions.SkipReferences is true.
	snap.ReferenceFetchAPI
}

// AttachmentFilter narrows the candidate set for an attachments cleanup.
type AttachmentFilter struct {
	// OlderThan is the cutoff: attachments with CreatedOn < OlderThan are
	// selected. Use time.Time{} to disable the age filter.
	OlderThan time.Time
	// EntityTypes restricts the selection to specific attachment parents.
	// Allowed values: "case", "run", "plan", "plan_entry", "result",
	// "test". Empty set = all.
	EntityTypes map[string]struct{}
	// Limit caps the total number of selected attachments across all
	// projects (0 = no limit). Selection is deterministic by
	// (project_id, attachment_id).
	Limit int
}

// Allowed reports whether the given attachment passes the filter.
func (f AttachmentFilter) Allowed(att data.Attachment) bool {
	if !f.OlderThan.IsZero() {
		if att.CreatedOn == 0 {
			return false
		}
		if time.Unix(att.CreatedOn, 0).After(f.OlderThan) {
			return false
		}
	}
	if len(f.EntityTypes) > 0 {
		kind := att.InferredEntityType()
		if _, ok := f.EntityTypes[kind]; !ok {
			return false
		}
	}
	return true
}

// ProjectSelection groups the matched attachments belonging to a single
// project, with summary statistics for the pre-flight summary.
type ProjectSelection struct {
	ProjectID   int64
	ProjectName string
	Attachments []data.Attachment
	TotalBytes  int64
	OldestUnix  int64
}

// Plan is the outcome of BuildPlan: the projects scanned and the
// attachments selected for deletion. It is the input to Execute.
type Plan struct {
	Projects        []ProjectSelection
	TotalCount      int
	TotalBytes      int64
	OldestUnix      int64
	TruncatedByLimit bool
}

// ProjectLister is the slice of the TestRail client surface used to
// resolve the project set walked by BuildPlan.
type ProjectLister interface {
	GetProjects(ctx context.Context) (data.GetProjectsResponse, error)
}

// BuildPlan walks the given projects, lists their attachments via the
// project-level endpoint, applies the filter and returns a
// deterministic Plan. When projectIDs is empty the walker enumerates
// every project the API key can see.
//
// Use BuildPlanWithScanner when the caller has already resolved a
// scan strategy (project vs entities) — for example via ResolveScanner.
// BuildPlan is preserved for back-compat and assumes the project-level
// endpoint is available on the server.
func BuildPlan(
	ctx context.Context,
	api AttachmentsAPI,
	projectIDs []int64,
	filter AttachmentFilter,
	concurrency int,
) (*Plan, error) {
	return BuildPlanWithScanner(ctx, api, NewProjectScanner(api), projectIDs, filter, concurrency)
}

// BuildPlanWithScanner is BuildPlan parameterised by an explicit
// AttachmentScanner. The scanner decides HOW each project's
// attachments are listed (single bulk call vs entity walk).
func BuildPlanWithScanner(
	ctx context.Context,
	lister ProjectLister,
	scanner AttachmentScanner,
	projectIDs []int64,
	filter AttachmentFilter,
	concurrency int,
) (*Plan, error) {
	return BuildPlanWithScannerProgress(ctx, lister, scanner, projectIDs, filter, concurrency, NoProgress)
}

// BuildPlanWithScannerProgress is BuildPlanWithScanner with an
// explicit progress sink. nil sink → NoProgress.
func BuildPlanWithScannerProgress(
	ctx context.Context,
	lister ProjectLister,
	scanner AttachmentScanner,
	projectIDs []int64,
	filter AttachmentFilter,
	concurrency int,
	progress ScanProgress,
) (*Plan, error) {
	if progress == nil {
		progress = NoProgress
	}
	if r, ok := scanner.(ScanProgressReceiver); ok {
		r.SetProgress(progress)
		defer r.SetProgress(NoProgress)
	}
	projects, err := resolveProjectsViaLister(ctx, lister, projectIDs)
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return &Plan{}, nil
	}

	total := len(projects)
	var startMu sync.Mutex
	var startedIdx int
	results, _ := concurrent.ParallelMap(ctx, projects, concurrency, func(p data.Project, _ int) (ProjectSelection, error) {
		startMu.Lock()
		startedIdx++
		idx := startedIdx
		startMu.Unlock()
		progress.OnProjectStart(idx, total, p.ID, p.Name)
		start := time.Now()
		sel, scanErr := scanProject(ctx, scanner, p, filter, 0)
		elapsed := time.Since(start)
		if scanErr != nil {
			progress.OnError(p.ID, scanErr)
		} else {
			progress.OnProjectDone(p.ID, len(sel.Attachments), len(sel.Attachments), sel.TotalBytes, elapsed)
		}
		return sel, scanErr
	})

	plan := &Plan{}
	for _, r := range results {
		if r.Error != nil {
			return nil, r.Error
		}
		if len(r.Data.Attachments) == 0 {
			continue
		}
		plan.Projects = append(plan.Projects, r.Data)
		plan.TotalCount += len(r.Data.Attachments)
		plan.TotalBytes += r.Data.TotalBytes
		if plan.OldestUnix == 0 || (r.Data.OldestUnix != 0 && r.Data.OldestUnix < plan.OldestUnix) {
			plan.OldestUnix = r.Data.OldestUnix
		}
	}

	sort.SliceStable(plan.Projects, func(i, j int) bool {
		return plan.Projects[i].ProjectID < plan.Projects[j].ProjectID
	})

	if filter.Limit > 0 && plan.TotalCount > filter.Limit {
		applyLimit(plan, filter.Limit)
		plan.TruncatedByLimit = true
	}
	return plan, nil
}

func resolveProjectsViaLister(ctx context.Context, lister ProjectLister, projectIDs []int64) ([]data.Project, error) {
	if len(projectIDs) == 0 {
		all, err := lister.GetProjects(ctx)
		if err != nil {
			return nil, fmt.Errorf("list projects: %w", err)
		}
		return all, nil
	}
	out := make([]data.Project, 0, len(projectIDs))
	for _, id := range projectIDs {
		p, err := lookupProjectViaLister(ctx, lister, id)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// lookupProjectViaLister prefers a single GET when the lister supports
// it; we fall back to scanning the full project list for environments
// where get_project is restricted.
func lookupProjectViaLister(ctx context.Context, lister ProjectLister, id int64) (data.Project, error) {
	if g, ok := lister.(interface {
		GetProject(context.Context, int64) (*data.GetProjectResponse, error)
	}); ok {
		p, err := g.GetProject(ctx, id)
		if err == nil && p != nil {
			return data.Project{ID: p.ID, Name: p.Name}, nil
		}
	}
	all, err := lister.GetProjects(ctx)
	if err != nil {
		return data.Project{}, fmt.Errorf("list projects (lookup %d): %w", id, err)
	}
	for _, p := range all {
		if p.ID == id {
			return p, nil
		}
	}
	return data.Project{}, fmt.Errorf("project %d not found", id)
}

func applyLimit(plan *Plan, limit int) {
	remaining := limit
	plan.TotalCount = 0
	plan.TotalBytes = 0
	plan.OldestUnix = 0
	for i := range plan.Projects {
		sel := &plan.Projects[i]
		if remaining <= 0 {
			sel.Attachments = nil
			sel.TotalBytes = 0
			continue
		}
		if len(sel.Attachments) > remaining {
			sel.Attachments = sel.Attachments[:remaining]
			var tb int64
			var oldest int64
			for _, a := range sel.Attachments {
				tb += a.Size
				if oldest == 0 || (a.CreatedOn != 0 && a.CreatedOn < oldest) {
					oldest = a.CreatedOn
				}
			}
			sel.TotalBytes = tb
			sel.OldestUnix = oldest
		}
		remaining -= len(sel.Attachments)
		plan.TotalCount += len(sel.Attachments)
		plan.TotalBytes += sel.TotalBytes
		if plan.OldestUnix == 0 || (sel.OldestUnix != 0 && sel.OldestUnix < plan.OldestUnix) {
			plan.OldestUnix = sel.OldestUnix
		}
	}
}

// FlatAttachments returns every selected attachment from every project
// in the plan, in deterministic order.
func (p *Plan) FlatAttachments() []data.Attachment {
	out := make([]data.Attachment, 0, p.TotalCount)
	for _, sel := range p.Projects {
		out = append(out, sel.Attachments...)
	}
	return out
}

// ExecuteOptions configures Execute.
type ExecuteOptions struct {
	// DryRun reports what would happen without taking a snapshot or
	// performing any deletes.
	DryRun bool
	// SkipSnapshot disables the pre-delete snapshot. Strongly
	// discouraged; normally surfaced via --no-snapshot.
	SkipSnapshot bool
	// SnapshotLabel is the optional human-readable name for the
	// resulting snapshot.
	SnapshotLabel string
	// CompressBinaries enables gzip compression for stored binaries.
	CompressBinaries bool
	// Concurrency sets the worker count for the delete phase.
	Concurrency int
	// BackupConcurrency sets the worker count for the snapshot
	// download phase. 0 means auto: min(8, Concurrency).
	BackupConcurrency int
	// SkipReferences disables the reference scan (markdown bodies that
	// point at the deleted attachments will not be indexed and cannot
	// be rewritten on restore). Use only when the entity API is
	// unhealthy and you accept lossy restore.
	SkipReferences bool
	// CLIArgs is propagated into the snapshot meta for traceability.
	CLIArgs []string
	// ServerURL is propagated into the snapshot meta. Empty = use
	// snap.CurrentServerURL().
	ServerURL string
}

// ExecuteResult summarizes an Execute call.
type ExecuteResult struct {
	SnapshotID   string
	DryRun       bool
	BackedUp     int
	BackupBytes  int64
	Deleted      int
	DeleteErrors int
	// FreedBytes is the sum of sizes of attachments that the server
	// confirmed deleted (i.e. failures excluded). Equals BackupBytes
	// when DeleteErrors == 0. Always 0 on dry-run.
	FreedBytes int64
	Failures   []DeleteFailure
	// EntitiesScanned is the number of (entity_type, entity_id) tuples
	// walked during the reference scan. 0 when SkipReferences=true.
	EntitiesScanned int
	// RefsIndexed is the number of attachment URLs persisted into
	// references.json.
	RefsIndexed int
	// IntegrityRoot is the merkle-style root hash recorded in
	// integrity.json. Empty when no snapshot was taken.
	IntegrityRoot string
}

// DeleteFailure is a single failed DeleteAttachment call.
type DeleteFailure struct {
	AttachmentID int64
	ProjectID    int64
	Error        string
}

// Execute performs the cleanup workflow described by plan: optional
// snapshot, then parallel DeleteAttachment calls. The snapshot is
// always taken unless opts.SkipSnapshot or opts.DryRun is set.
//nolint:gocyclo // Two-phase executor (snapshot → parallel deletes) tracked with explicit branches keeps the rollback contract visible.
func Execute(
	ctx context.Context,
	api AttachmentsAPI,
	store *snap.Store,
	manifest *snap.Manifest,
	plan *Plan,
	opts ExecuteOptions,
) (*ExecuteResult, error) {
	res := &ExecuteResult{DryRun: opts.DryRun}
	atts := plan.FlatAttachments()
	if len(atts) == 0 {
		return res, nil
	}

	if opts.DryRun {
		res.BackedUp = len(atts)
		for _, a := range atts {
			res.BackupBytes += a.Size
		}
		return res, nil
	}

	// 1. Snapshot phase (default ON).
	if !opts.SkipSnapshot {
		entityIDs := make([]int64, 0, len(atts))
		for _, a := range atts {
			entityIDs = append(entityIDs, a.ID)
		}
		projectID := int64(0)
		if len(plan.Projects) == 1 {
			projectID = plan.Projects[0].ProjectID
		}
		meta := snap.BuildMeta(
			snap.OpDelete,
			snap.EntityTypeAttachments,
			entityIDs,
			snap.Tier2,
			projectID,
			0,
			opts.SnapshotLabel,
			opts.CLIArgs,
			opts.ServerURL,
		)
		meta.Timestamp = time.Now().UTC()
		meta.Status = snap.StatusAvailable
		meta.DataFile = "attachments.json"

		if err := store.SaveMeta(&meta); err != nil {
			return res, fmt.Errorf("snapshot: save meta: %w", err)
		}
		saved, bytes, err := snap.BackupAttachmentsForCleanup(ctx, api, store, meta.ID, atts, snap.BackupOptions{
			Compress:    opts.CompressBinaries,
			Concurrency: resolveBackupConcurrency(opts),
		})
		if err != nil {
			return res, fmt.Errorf("snapshot: backup binaries: %w", err)
		}
		meta.DataSizeBytes = bytes
		if err := store.SaveMeta(&meta); err != nil {
			return res, fmt.Errorf("snapshot: rewrite meta: %w", err)
		}
		if err := manifest.Add(&meta); err != nil {
			return res, fmt.Errorf("snapshot: manifest add: %w", err)
		}
		res.SnapshotID = meta.ID
		res.BackedUp = saved
		res.BackupBytes = bytes

		// 1b. Reference scan: walk every entity that owned a deleted
		// attachment and persist references.json next to attachments.json.
		// Best-effort — partial indexing is preferable to aborting the
		// cleanup. Empty array means "scanned, found nothing", which
		// is distinct from "scan was skipped" (no file written).
		if !opts.SkipReferences {
			refsList := snap.ScanReferencesForAttachments(ctx, api, atts)
			if err := snap.WriteReferencesSidecar(store, meta.ID, refsList); err != nil {
				return res, fmt.Errorf("snapshot: write references.json: %w", err)
			}
			res.EntitiesScanned = len(refsList)
			for _, e := range refsList {
				res.RefsIndexed += len(e.Refs)
			}
		}

		// 1c. Integrity index over the entire snapshot directory.
		idx, err := snap.WriteIntegrityIndex(store, meta.ID)
		if err != nil {
			return res, fmt.Errorf("snapshot: write integrity.json: %w", err)
		}
		res.IntegrityRoot = idx.Root
	}

	// 2. Delete phase (parallel).
	type item struct {
		ID        int64
		ProjectID int64
		Size      int64
	}
	deleteItems := make([]item, 0, len(atts))
	for _, sel := range plan.Projects {
		for _, a := range sel.Attachments {
			deleteItems = append(deleteItems, item{ID: a.ID, ProjectID: sel.ProjectID, Size: a.Size})
		}
	}

	delResults, _ := concurrent.ParallelMap(ctx, deleteItems, opts.Concurrency, func(it item, _ int) (int64, error) {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := api.DeleteAttachment(ctx, it.ID); err != nil {
			return it.ID, err
		}
		return it.ID, nil
	})
	for i, r := range delResults {
		if r.Error != nil {
			res.DeleteErrors++
			res.Failures = append(res.Failures, DeleteFailure{
				AttachmentID: deleteItems[i].ID,
				ProjectID:    deleteItems[i].ProjectID,
				Error:        r.Error.Error(),
			})
			continue
		}
		res.Deleted++
		res.FreedBytes += deleteItems[i].Size
	}
	return res, nil
}

// Compile-time assertion that *client.HTTPClient satisfies AttachmentsAPI.
var _ AttachmentsAPI = (*client.HTTPClient)(nil)

// resolveBackupConcurrency picks the snapshot-download worker count.
// 0 means "auto = min(8, Concurrency)" so tiny --concurrency values
// don't accidentally enable a thundering herd of downloads.
func resolveBackupConcurrency(opts ExecuteOptions) int {
	if opts.BackupConcurrency > 0 {
		return opts.BackupConcurrency
	}
	auto := opts.Concurrency
	if auto > 8 {
		auto = 8
	}
	if auto <= 0 {
		auto = 1
	}
	return auto
}
