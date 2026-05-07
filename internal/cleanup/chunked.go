// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync/atomic"
	"time"

	"github.com/Korrnals/gotr/internal/cleanup/checkpoint"
	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/models/data"
)

// ChunkConfig drives BuildPlanChunked.
type ChunkConfig struct {
	// ChunkSize is the maximum number of projects scanned in parallel
	// before flushing to the checkpoint store. Defaulted to 10 when
	// non-positive.
	ChunkSize int
	// ScanTimeoutPerProject caps a single project's Scan call. Zero
	// disables the per-project timeout. CLI default is 10m.
	ScanTimeoutPerProject time.Duration
	// RunID identifies this run on disk. Empty → a new ID is generated.
	RunID string
	// Resume causes BuildPlanChunked to load the existing checkpoint
	// for RunID and continue from where it stopped.
	Resume bool
	// Store is the checkpoint persistence layer. MUST be non-nil.
	Store *checkpoint.Store
	// Filter snapshot recorded in the checkpoint. Used for resume
	// integrity verification.
	FilterSnapshot checkpoint.FilterSnapshot
	// AllProjects is captured in the checkpoint for resume verification
	// and audit.
	AllProjects bool
	// CLIArgs is captured in the checkpoint for audit only.
	CLIArgs []string
	// OnChunkComplete fires after each chunk has been persisted. May
	// be nil. partial is the cumulative plan; idx is 1-based, total is
	// the number of chunks.
	OnChunkComplete func(idx, total int, partial *Plan)
	// Progress receives per-project lifecycle and per-phase events.
	// nil → NoProgress. Scanners that implement ScanProgressReceiver
	// have their sink installed for the duration of the call.
	Progress ScanProgress
}

// ErrCheckpointMismatch is returned when a --resume invocation cannot
// safely continue an earlier run because the project set, filter or
// scan strategy has diverged.
var ErrCheckpointMismatch = errors.New("resume: checkpoint mismatch")

// scanProject is the per-project scan body shared by BuildPlanWithScanner
// and BuildPlanChunked. It applies AttachmentFilter to the scanner
// output and returns a deterministic ProjectSelection. When timeout > 0
// the scan is wrapped in a context.WithTimeout; the caller distinguishes
// timeout errors via errors.Is(err, context.DeadlineExceeded).
func scanProject(ctx context.Context, scanner AttachmentScanner, p data.Project, filter AttachmentFilter, timeout time.Duration) (ProjectSelection, error) {
	scanCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		scanCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	atts, err := scanner.Scan(scanCtx, p.ID)
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
}

// BuildPlanChunked scans projects in chunks, persisting a checkpoint
// after every chunk and returning the cumulative Plan plus the final
// checkpoint state. Per-project timeouts isolate "stuck" projects so
// the rest of the run still progresses.
//
// When cfg.Resume is true the caller's project set, filter and
// concurrency MUST match the checkpoint on disk; otherwise
// BuildPlanChunked returns ErrCheckpointMismatch and refuses to
// silently override the persisted state.
//
// When the run completes with NO failed/timeout projects the
// checkpoint directory is auto-deleted. Otherwise it is preserved so
// the caller can rerun with --resume.
//
//nolint:gocyclo // Chunked driver: project resolution + resume verification + per-chunk scan + persist + summarize is best read top-to-bottom.
func BuildPlanChunked(
	ctx context.Context,
	lister ProjectLister,
	scanner AttachmentScanner,
	projectIDs []int64,
	filter AttachmentFilter,
	concurrency int,
	cfg ChunkConfig,
) (*Plan, *checkpoint.Checkpoint, error) {
	if cfg.Store == nil {
		return nil, nil, errors.New("BuildPlanChunked: ChunkConfig.Store is required")
	}
	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = 10
	}
	if cfg.Progress == nil {
		cfg.Progress = NoProgress
	}
	if r, ok := scanner.(ScanProgressReceiver); ok {
		r.SetProgress(cfg.Progress)
		defer r.SetProgress(NoProgress)
	}

	projects, err := resolveProjectsViaLister(ctx, lister, projectIDs)
	if err != nil {
		return nil, nil, err
	}

	cp, plan, err := initOrResumeCheckpoint(projects, projectIDs, filter, concurrency, cfg)
	if err != nil {
		return nil, nil, err
	}

	// Build a workset: projects whose state still requires scanning.
	type pendingProject struct {
		project data.Project
		idx     int // index into cp.Projects
	}
	var workset []pendingProject
	for i, ps := range cp.Projects {
		if !cfg.Resume && ps.State != "" && ps.State != checkpoint.StatePending && ps.State != checkpoint.StateRetryPending {
			// In a fresh run every project is pending — but defensive
			// guard for callers that pass an existing checkpoint.
			continue
		}
		switch ps.State {
		case checkpoint.StateDone:
			continue
		case checkpoint.StateFailed, checkpoint.StateTimeout:
			if !cfg.Resume {
				continue
			}
			// On resume: failed/timeout left intact unless explicitly
			// flipped to retry_pending by the operator. Treated as
			// terminal here.
			continue
		}
		// Find the corresponding data.Project (project list and
		// checkpoint are co-ordered by construction).
		if i < len(projects) {
			workset = append(workset, pendingProject{project: projects[i], idx: i})
		}
	}

	totalChunks := chunkCount(len(workset), cfg.ChunkSize)
	totalProjects := len(workset)
	var startCounter int64

	for chunkIdx, start := 0, 0; start < len(workset); start += cfg.ChunkSize {
		end := start + cfg.ChunkSize
		if end > len(workset) {
			end = len(workset)
		}
		chunkIdx++
		batch := workset[start:end]

		// Stamp StartedAt for each project in the batch before scanning,
		// so an interrupted run still has a paper trail.
		now := time.Now().UTC()
		for _, w := range batch {
			cp.Projects[w.idx].State = checkpoint.StatePending
			cp.Projects[w.idx].StartedAt = now
			cp.Projects[w.idx].Reason = ""
		}

		results, _ := concurrent.ParallelMap(ctx, batch, concurrency, func(w pendingProject, _ int) (ProjectSelection, error) {
			idx := int(atomic.AddInt64(&startCounter, 1))
			cfg.Progress.OnProjectStart(idx, totalProjects, w.project.ID, w.project.Name)
			projStart := time.Now()
			sel, scanErr := scanProject(ctx, scanner, w.project, filter, cfg.ScanTimeoutPerProject)
			elapsed := time.Since(projStart)
			if scanErr != nil {
				cfg.Progress.OnError(w.project.ID, scanErr)
			} else {
				cfg.Progress.OnProjectDone(w.project.ID, len(sel.Attachments), len(sel.Attachments), sel.TotalBytes, elapsed)
			}
			return sel, scanErr
		})

		for i, r := range results {
			w := batch[i]
			finishedAt := time.Now().UTC()
			cp.Projects[w.idx].EndedAt = finishedAt
			if r.Error != nil {
				if errors.Is(r.Error, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) && errors.Is(r.Error, context.Canceled) {
					cp.Projects[w.idx].State = checkpoint.StateTimeout
				} else if errors.Is(r.Error, context.DeadlineExceeded) {
					cp.Projects[w.idx].State = checkpoint.StateTimeout
				} else {
					cp.Projects[w.idx].State = checkpoint.StateFailed
				}
				cp.Projects[w.idx].Reason = r.Error.Error()
				continue
			}
			cp.Projects[w.idx].State = checkpoint.StateDone
			cp.Projects[w.idx].Found = len(r.Data.Attachments)
			cp.Projects[w.idx].Eligible = len(r.Data.Attachments)
			cp.Projects[w.idx].Bytes = r.Data.TotalBytes
			cp.Projects[w.idx].Reason = ""

			if len(r.Data.Attachments) > 0 {
				plan.Projects = append(plan.Projects, r.Data)
				plan.TotalCount += len(r.Data.Attachments)
				plan.TotalBytes += r.Data.TotalBytes
				if plan.OldestUnix == 0 || (r.Data.OldestUnix != 0 && r.Data.OldestUnix < plan.OldestUnix) {
					plan.OldestUnix = r.Data.OldestUnix
				}
			}
		}

		sort.SliceStable(plan.Projects, func(i, j int) bool {
			return plan.Projects[i].ProjectID < plan.Projects[j].ProjectID
		})

		cp.UpdatedAt = time.Now().UTC()
		if err := cfg.Store.Save(cp.RunID, cp); err != nil {
			return plan, cp, fmt.Errorf("checkpoint save: %w", err)
		}
		if err := cfg.Store.SavePartialPlan(cp.RunID, plan); err != nil {
			return plan, cp, fmt.Errorf("partial-plan save: %w", err)
		}
		if cfg.OnChunkComplete != nil {
			cfg.OnChunkComplete(chunkIdx, totalChunks, plan)
		}

		if err := ctx.Err(); err != nil {
			return plan, cp, err
		}
	}

	// Apply --limit AFTER all chunks have been gathered, so the global
	// limit is deterministic regardless of chunk boundaries.
	if filter.Limit > 0 && plan.TotalCount > filter.Limit {
		applyLimit(plan, filter.Limit)
		plan.TruncatedByLimit = true
	}

	// Auto-delete checkpoint on perfectly clean completion.
	if isCleanCompletion(cp) {
		if err := cfg.Store.Delete(cp.RunID); err != nil {
			// Non-fatal: surface as a warning via the returned plan.
			return plan, cp, fmt.Errorf("auto-delete checkpoint: %w", err)
		}
	}

	return plan, cp, nil
}

// initOrResumeCheckpoint either loads + verifies an existing checkpoint
// (Resume=true) or builds a fresh one (Resume=false) for the supplied
// project list.
func initOrResumeCheckpoint(projects []data.Project, projectIDs []int64, filter AttachmentFilter, concurrency int, cfg ChunkConfig) (*checkpoint.Checkpoint, *Plan, error) {
	now := time.Now().UTC()
	if cfg.Resume {
		if cfg.RunID == "" {
			return nil, nil, errors.New("BuildPlanChunked: Resume requires RunID")
		}
		cp, err := cfg.Store.Load(cfg.RunID)
		if err != nil {
			return nil, nil, err
		}
		if err := verifyCheckpointMatches(cp, projects, projectIDs, cfg, concurrency); err != nil {
			return nil, nil, err
		}
		// Restore the partial plan accumulated by previous chunks.
		plan := &Plan{}
		if err := cfg.Store.LoadPartialPlan(cfg.RunID, plan); err != nil && !errors.Is(err, checkpoint.ErrCheckpointNotFound) {
			return nil, nil, err
		}
		return cp, plan, nil
	}

	runID := cfg.RunID
	if runID == "" {
		runID = checkpoint.NewRunID(now)
	}
	cp := &checkpoint.Checkpoint{
		Schema:      checkpoint.Schema,
		RunID:       runID,
		StartedAt:   now,
		UpdatedAt:   now,
		CLIArgs:     cfg.CLIArgs,
		Filter:      cfg.FilterSnapshot,
		AllProjects: cfg.AllProjects,
		ProjectIDs:  append([]int64(nil), projectIDs...),
		ChunkSize:   cfg.ChunkSize,
		Projects:    make([]checkpoint.ProjectStatus, 0, len(projects)),
	}
	if cp.Filter.Concurrency == 0 {
		cp.Filter.Concurrency = concurrency
	}
	for _, p := range projects {
		cp.Projects = append(cp.Projects, checkpoint.ProjectStatus{ID: p.ID, Name: p.Name, State: checkpoint.StatePending})
	}
	if err := cfg.Store.Save(runID, cp); err != nil {
		return nil, nil, fmt.Errorf("checkpoint save: %w", err)
	}
	return cp, &Plan{}, nil
}

func verifyCheckpointMatches(cp *checkpoint.Checkpoint, projects []data.Project, projectIDs []int64, cfg ChunkConfig, concurrency int) error {
	// Project set must match exactly (by ID, in order).
	if len(cp.Projects) != len(projects) {
		return fmt.Errorf("%w: project count differs (checkpoint=%d, current=%d)", ErrCheckpointMismatch, len(cp.Projects), len(projects))
	}
	for i, p := range projects {
		if cp.Projects[i].ID != p.ID {
			return fmt.Errorf("%w: project[%d].id checkpoint=%d current=%d", ErrCheckpointMismatch, i, cp.Projects[i].ID, p.ID)
		}
	}
	if !reflect.DeepEqual(cp.ProjectIDs, projectIDs) && len(cp.ProjectIDs) != 0 {
		// ProjectIDs may be empty for --all-projects; only compare
		// when both sides are explicit.
		if len(projectIDs) != 0 {
			return fmt.Errorf("%w: explicit project IDs differ", ErrCheckpointMismatch)
		}
	}
	if cp.AllProjects != cfg.AllProjects {
		return fmt.Errorf("%w: --all-projects flag differs", ErrCheckpointMismatch)
	}
	if !filterSnapshotEqual(cp.Filter, cfg.FilterSnapshot) {
		return fmt.Errorf("%w: filter snapshot differs (checkpoint=%+v current=%+v)", ErrCheckpointMismatch, cp.Filter, cfg.FilterSnapshot)
	}
	if cp.ChunkSize != cfg.ChunkSize {
		return fmt.Errorf("%w: chunk size differs (checkpoint=%d current=%d)", ErrCheckpointMismatch, cp.ChunkSize, cfg.ChunkSize)
	}
	_ = concurrency
	return nil
}

// filterSnapshotEqual performs the comparison that controls resume
// gating: ScanStrategy + EntityTypes + OlderThan + Limit must match.
// Concurrency and ScanTimeoutPerProject are recorded for audit but do
// NOT participate in the equality test (they are operational tuning).
func filterSnapshotEqual(a, b checkpoint.FilterSnapshot) bool {
	if a.OlderThanRaw != b.OlderThanRaw {
		return false
	}
	if a.ScanStrategy != b.ScanStrategy {
		return false
	}
	if a.Limit != b.Limit {
		return false
	}
	if !sortedSliceEqual(a.EntityTypes, b.EntityTypes) {
		return false
	}
	return true
}

func sortedSliceEqual(a, b []string) bool {
	aa := append([]string(nil), a...)
	bb := append([]string(nil), b...)
	sort.Strings(aa)
	sort.Strings(bb)
	return reflect.DeepEqual(aa, bb)
}

func isCleanCompletion(cp *checkpoint.Checkpoint) bool {
	for _, ps := range cp.Projects {
		switch ps.State {
		case checkpoint.StateFailed, checkpoint.StateTimeout, checkpoint.StatePending, checkpoint.StateRetryPending, "":
			if ps.State != checkpoint.StateDone {
				return false
			}
		}
	}
	return true
}

func chunkCount(n, size int) int {
	if size <= 0 || n <= 0 {
		return 0
	}
	return (n + size - 1) / size
}
