// Package cleanup implements bulk-cleanup workflows for TestRail data
// kept reversible via the snap engine. The first supported workflow is
// attachments cleanup: filter by age and parent entity type, snapshot
// the binaries, then delete in parallel.
package cleanup

import (
	"context"
	"fmt"
	"sort"
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
//nolint:gocyclo // Plan walker: project resolution + parallel fetch + filter + limit are best read top-to-bottom.
func BuildPlanWithScanner(
	ctx context.Context,
	lister ProjectLister,
	scanner AttachmentScanner,
	projectIDs []int64,
	filter AttachmentFilter,
	concurrency int,
) (*Plan, error) {
	projects, err := resolveProjectsViaLister(ctx, lister, projectIDs)
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return &Plan{}, nil
	}

	results, _ := concurrent.ParallelMap(ctx, projects, concurrency, func(p data.Project, _ int) (ProjectSelection, error) {
		atts, err := scanner.Scan(ctx, p.ID)
		if err != nil {
			return ProjectSelection{}, fmt.Errorf("project %d: %w", p.ID, err)
		}
		sel := ProjectSelection{ProjectID: p.ID, ProjectName: p.Name}
		for _, att := range atts {
			if !filter.Allowed(att) {
				continue
			}
			sel.Attachments = append(sel.Attachments, att)
			sel.TotalBytes += att.Size
			if sel.OldestUnix == 0 || (att.CreatedOn != 0 && att.CreatedOn < sel.OldestUnix) {
				sel.OldestUnix = att.CreatedOn
			}
		}
		sort.SliceStable(sel.Attachments, func(i, j int) bool {
			return sel.Attachments[i].ID < sel.Attachments[j].ID
		})
		return sel, nil
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
	Failures     []DeleteFailure
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
		meta.DataFile = "data.json"

		if err := store.SaveMeta(&meta); err != nil {
			return res, fmt.Errorf("snapshot: save meta: %w", err)
		}
		saved, bytes, err := snap.BackupAttachmentsForCleanup(ctx, api, store, meta.ID, atts, opts.CompressBinaries)
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
	}

	// 2. Delete phase (parallel).
	type item struct {
		ID        int64
		ProjectID int64
	}
	deleteItems := make([]item, 0, len(atts))
	for _, sel := range plan.Projects {
		for _, a := range sel.Attachments {
			deleteItems = append(deleteItems, item{ID: a.ID, ProjectID: sel.ProjectID})
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
	}
	return res, nil
}

// Compile-time assertion that *client.HTTPClient satisfies AttachmentsAPI.
var _ AttachmentsAPI = (*client.HTTPClient)(nil)
