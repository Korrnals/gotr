// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"sort"
	"time"

	cleanupsvc "github.com/Korrnals/gotr/internal/cleanup"
)

// BuildInput aggregates everything needed to assemble a Report from a
// completed (or dry-run) `gotr attachments cleanup` invocation.
type BuildInput struct {
	Plan   *cleanupsvc.Plan
	Result *cleanupsvc.ExecuteResult

	// Header context.
	Timestamp time.Time
	Server    string
	GotrVer   string
	Label     string
	User      string
	CLIArgs   []string

	// Filters context (mirrors the parsed CLI flags).
	ProjectIDs   []int64
	AllProjects  bool
	OlderThanRaw string
	CutoffUnix   int64
	EntityTypes  []string
	ScanStrategy string
	Limit        int

	// Chunked-execution + concurrency context (v3.6.0).
	RunID                 string
	ChunkSize             int
	ChunksTotal           int
	ChunksCompleted       int
	ScanTimeoutPerProject string
	DeleteConcurrency     int
	BackupConcurrency     int
	ResumedFrom           string
	SkipReferences        bool
	Compress              bool

	// Snapshot artifact paths and counters (v3.6.0). All paths
	// should be absolute when known. Empty values are omitted.
	SnapshotPath         string
	MetaPath             string
	MappingPath          string
	MappingSchemaVersion int
	MappingTotal         int
	MappingRestorable    int
	ReferencesPath       string
	IntegrityPath        string
	IntegrityFiles       int
	FilesDir             string
	FilesCount           int
	FilesBytes           int64
	// ReferencesByEntity is a precomputed entity_type→count map of
	// indexed markdown URL refs (from references.json). Used by the
	// "Known limitations" callout in the report.
	ReferencesByEntity map[string]int

	// Generated artifacts on disk (v3.6.0).
	ReportPaths   []string
	CheckpointDir string

	// ID is the report identifier; if empty, NewID() is generated.
	ID string
}

// Build assembles a Report from a completed (or dry-run) cleanup
// invocation. The returned report is safe to render via any of the
// Render* helpers in this package.
func Build(in BuildInput) *Report {
	if in.Timestamp.IsZero() {
		in.Timestamp = time.Now().UTC()
	}
	rep := &Report{
		ID:        firstNonEmpty(in.ID, NewID(in.Timestamp)),
		RunID:     in.RunID,
		Timestamp: in.Timestamp.UTC(),
		Server:    in.Server,
		GotrVer:   in.GotrVer,
		Label:     in.Label,
		User:      in.User,
		CLIArgs:   in.CLIArgs,
		Filters: Filters{
			ProjectIDs:   append([]int64(nil), in.ProjectIDs...),
			AllProjects:  in.AllProjects,
			OlderThan:    in.OlderThanRaw,
			CutoffUnix:   in.CutoffUnix,
			EntityTypes:  append([]string(nil), in.EntityTypes...),
			ScanStrategy: in.ScanStrategy,
			Limit:        in.Limit,
		},
	}
	sort.Strings(rep.Filters.EntityTypes)

	if in.Result != nil {
		rep.SnapshotID = in.Result.SnapshotID
		rep.DryRun = in.Result.DryRun
		rep.Summary = Summary{
			BackedUp:      in.Result.BackedUp,
			BackupSkipped: in.Result.BackupSkipped,
			BackupBytes:   in.Result.BackupBytes,
			Deleted:       in.Result.Deleted,
			Failed:        in.Result.DeleteErrors,
			FreedBytes:    in.Result.FreedBytes,
		}
		for _, f := range in.Result.Failures {
			rep.Failures = append(rep.Failures, Failure{
				AttachmentID: f.AttachmentID,
				ProjectID:    f.ProjectID,
				Error:        f.Error,
			})
		}
	}

	if in.Plan != nil {
		rep.Summary.TotalSelected = in.Plan.TotalCount
		// Fill BackupBytes from plan when result didn't compute it (dry-run).
		if rep.Summary.BackupBytes == 0 {
			rep.Summary.BackupBytes = in.Plan.TotalBytes
		}
		for _, sel := range in.Plan.Projects {
			if len(sel.Attachments) == 0 {
				continue
			}
			pg := ProjectGroup{
				ProjectID:   sel.ProjectID,
				ProjectName: sel.ProjectName,
				Count:       len(sel.Attachments),
				TotalBytes:  sel.TotalBytes,
				OldestUnix:  sel.OldestUnix,
				Items:       make([]Record, 0, len(sel.Attachments)),
			}
			for _, a := range sel.Attachments {
				pg.Items = append(pg.Items, FromAttachment(a))
			}
			rep.Projects = append(rep.Projects, pg)
		}
		rep.EntityBreakdown = buildEntityBreakdown(in.Plan)
	}

	rep.Chunking = buildChunkingInfo(in)
	rep.Snapshot = buildSnapshotInfo(in, rep)
	rep.Artifacts = buildArtifacts(in)

	return rep
}

// buildChunkingInfo returns a populated ChunkingInfo, or nil when no
// chunking-related field was provided (e.g. in older callers / tests).
func buildChunkingInfo(in BuildInput) *ChunkingInfo {
	if in.ChunkSize == 0 && in.ChunksTotal == 0 && in.DeleteConcurrency == 0 &&
		in.BackupConcurrency == 0 && in.ResumedFrom == "" &&
		in.ScanTimeoutPerProject == "" && !in.SkipReferences && !in.Compress {
		return nil
	}
	return &ChunkingInfo{
		ChunkSize:             in.ChunkSize,
		ChunksTotal:           in.ChunksTotal,
		ChunksCompleted:       in.ChunksCompleted,
		ScanTimeoutPerProject: in.ScanTimeoutPerProject,
		DeleteConcurrency:     in.DeleteConcurrency,
		BackupConcurrency:     in.BackupConcurrency,
		ResumedFrom:           in.ResumedFrom,
		SkipReferences:        in.SkipReferences,
		Compress:              in.Compress,
	}
}

// buildSnapshotInfo returns a SnapshotInfo populated from BuildInput
// and Result counters. Returns nil when no snapshot was produced.
func buildSnapshotInfo(in BuildInput, rep *Report) *SnapshotInfo {
	if rep.SnapshotID == "" && in.SnapshotPath == "" {
		return nil
	}
	si := &SnapshotInfo{
		Path:                 in.SnapshotPath,
		MetaPath:             in.MetaPath,
		MappingPath:          in.MappingPath,
		MappingSchemaVersion: in.MappingSchemaVersion,
		MappingTotal:         in.MappingTotal,
		MappingRestorable:    in.MappingRestorable,
		ReferencesPath:       in.ReferencesPath,
		ReferencesSkipped:    in.SkipReferences,
		IntegrityPath:        in.IntegrityPath,
		IntegrityFiles:       in.IntegrityFiles,
		FilesDir:             in.FilesDir,
		FilesCount:           in.FilesCount,
		FilesBytes:           in.FilesBytes,
		ReferencesByEntity:   in.ReferencesByEntity,
	}
	if in.Result != nil {
		si.EntitiesScanned = in.Result.EntitiesScanned
		si.RefsIndexed = in.Result.RefsIndexed
		si.IntegrityRoot = in.Result.IntegrityRoot
	}
	return si
}

// buildArtifacts returns nil when no artifact paths are known.
func buildArtifacts(in BuildInput) *Artifacts {
	if len(in.ReportPaths) == 0 && in.SnapshotPath == "" && in.CheckpointDir == "" {
		return nil
	}
	return &Artifacts{
		ReportPaths:   append([]string(nil), in.ReportPaths...),
		SnapshotPath:  in.SnapshotPath,
		CheckpointDir: in.CheckpointDir,
	}
}

// buildEntityBreakdown aggregates plan.Projects into a per-project ×
// entity-type matrix using the canonical entity-type column order.
func buildEntityBreakdown(plan *cleanupsvc.Plan) []ProjectEntityRow {
	if plan == nil || len(plan.Projects) == 0 {
		return nil
	}
	out := make([]ProjectEntityRow, 0, len(plan.Projects))
	for _, sel := range plan.Projects {
		if len(sel.Attachments) == 0 {
			continue
		}
		row := ProjectEntityRow{
			ProjectID:   sel.ProjectID,
			ProjectName: sel.ProjectName,
			Counts:      map[string]int{},
		}
		for _, a := range sel.Attachments {
			kind := a.InferredEntityType()
			row.Counts[kind]++
			row.Total++
			row.Bytes += a.Size
		}
		out = append(out, row)
	}
	return out
}

// NewID returns a deterministic report ID derived from the timestamp.
func NewID(ts time.Time) string {
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	return "cleanup-" + ts.UTC().Format("20060102T150405Z")
}
