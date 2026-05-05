// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

// Package cleanup builds a deletion-audit report for `gotr attachments
// cleanup`. The report is emitted in four formats — Markdown, JSON,
// CSV, and PDF — under
//
//	~/.gotr/reports/cleanup-attachments/<label>/<YYYY-MM>/cleanup-<id>-<ts>.<ext>
//
// The Markdown rendition is the human-readable artifact, JSON is the
// machine-readable mirror, CSV is a flat per-attachment table, and PDF
// is a fixed-format document for ticket attachments.
package cleanup

import (
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
)

// Report is the canonical, format-independent summary of a single
// `gotr attachments cleanup` invocation.
type Report struct {
	// Header.
	ID         string    `json:"id"`
	RunID      string    `json:"run_id,omitempty"`
	SnapshotID string    `json:"snapshot_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Server     string    `json:"server,omitempty"`
	GotrVer    string    `json:"gotr_version,omitempty"`
	Label      string    `json:"label,omitempty"`
	User       string    `json:"user,omitempty"`
	CLIArgs    []string  `json:"cli_args,omitempty"`
	DryRun     bool      `json:"dry_run"`

	// Filters that were applied.
	Filters Filters `json:"filters"`

	// Chunked execution + concurrency context (v3.6.0).
	Chunking *ChunkingInfo `json:"chunking,omitempty"`

	// Aggregate counters.
	Summary Summary `json:"summary"`

	// Per-project breakdown.
	Projects []ProjectGroup `json:"projects"`

	// Per-project × entity-type matrix (v3.6.0).
	EntityBreakdown []ProjectEntityRow `json:"entity_breakdown,omitempty"`

	// Snapshot v2 artifacts (mapping/references/integrity) — v3.6.0.
	Snapshot *SnapshotInfo `json:"snapshot,omitempty"`

	// Failures (only set when DeleteErrors > 0).
	Failures []Failure `json:"failures,omitempty"`

	// File-system locations of every artifact written by this run
	// (audit reports, snapshot dir, checkpoint cache). v3.6.0.
	Artifacts *Artifacts `json:"artifacts,omitempty"`
}

// ChunkingInfo describes the chunked execution & concurrency profile
// of a `gotr attachments cleanup` invocation. Mirrors v3.6.0 flags.
type ChunkingInfo struct {
	ChunkSize             int    `json:"chunk_size"`
	ChunksTotal           int    `json:"chunks_total,omitempty"`
	ChunksCompleted       int    `json:"chunks_completed,omitempty"`
	ScanTimeoutPerProject string `json:"scan_timeout_per_project,omitempty"`
	DeleteConcurrency     int    `json:"delete_concurrency"`
	BackupConcurrency     int    `json:"backup_concurrency,omitempty"`
	ResumedFrom           string `json:"resumed_from,omitempty"`
	SkipReferences        bool   `json:"skip_references,omitempty"`
	Compress              bool   `json:"compress,omitempty"`
}

// ProjectEntityRow is one row of the per-project × entity-type matrix.
// Counts is keyed by entity-type ("case","run","plan","plan_entry","result","test").
type ProjectEntityRow struct {
	ProjectID   int64          `json:"project_id"`
	ProjectName string         `json:"project_name,omitempty"`
	Counts      map[string]int `json:"counts"`
	Total       int            `json:"total"`
	Bytes       int64          `json:"bytes"`
}

// SnapshotInfo summarizes the contents of the snapshot directory
// produced by the cleanup run. Paths are absolute when known.
type SnapshotInfo struct {
	Path                 string `json:"path,omitempty"`
	MetaPath             string `json:"meta_path,omitempty"`
	MappingPath          string `json:"mapping_path,omitempty"`
	MappingSchemaVersion int    `json:"mapping_schema_version,omitempty"`
	MappingTotal         int    `json:"mapping_total,omitempty"`
	MappingRestorable    int    `json:"mapping_restorable,omitempty"`
	ReferencesPath       string `json:"references_path,omitempty"`
	EntitiesScanned      int    `json:"entities_scanned"`
	RefsIndexed          int    `json:"refs_indexed"`
	ReferencesSkipped    bool   `json:"references_skipped,omitempty"`
	IntegrityPath        string `json:"integrity_path,omitempty"`
	IntegrityRoot        string `json:"integrity_merkle_root,omitempty"`
	IntegrityFiles       int    `json:"integrity_files,omitempty"`
	FilesDir             string `json:"files_dir,omitempty"`
	FilesCount           int    `json:"files_count,omitempty"`
	FilesBytes           int64  `json:"files_bytes,omitempty"`
	// ReferencesByEntity counts indexed markdown URL refs grouped by
	// entity type (case/run/plan/milestone/...). Populated from
	// references.json. Helpful for visualizing the rewrite gap (see
	// "Known limitations" in the docs): these refs are recorded but
	// NOT rewritten in v3.6.0 — gotr does not modify external entity
	// bodies on cleanup or rollback.
	ReferencesByEntity map[string]int `json:"references_by_entity,omitempty"`
}

// Artifacts captures the absolute file-system locations of every
// artifact produced by the cleanup run (audit reports, snapshot dir,
// checkpoint cache). Useful for support hand-offs.
type Artifacts struct {
	ReportPaths   []string `json:"report_paths,omitempty"`
	SnapshotPath  string   `json:"snapshot_path,omitempty"`
	CheckpointDir string   `json:"checkpoint_dir,omitempty"`
}

// Filters captures the filters in effect for this run.
type Filters struct {
	ProjectIDs   []int64  `json:"project_ids,omitempty"`
	AllProjects  bool     `json:"all_projects"`
	OlderThan    string   `json:"older_than,omitempty"` // human form, e.g. "3M"
	CutoffUnix   int64    `json:"cutoff_unix,omitempty"`
	EntityTypes  []string `json:"entity_types,omitempty"`
	ScanStrategy string   `json:"scan_strategy,omitempty"`
	Limit        int      `json:"limit,omitempty"`
}

// Summary holds aggregate counters.
type Summary struct {
	TotalSelected int   `json:"total_selected"`
	BackedUp      int   `json:"backed_up"`
	BackupBytes   int64 `json:"backup_bytes"`
	Deleted       int   `json:"deleted"`
	Failed        int   `json:"failed"`
	// FreedBytes is the on-server space reclaimed: sum of sizes of
	// attachments the server confirmed deleted. Equals BackupBytes
	// when Failed == 0. Always 0 on dry-run.
	FreedBytes int64 `json:"freed_bytes"`
}

// ProjectGroup holds per-project deletion info.
type ProjectGroup struct {
	ProjectID   int64    `json:"project_id"`
	ProjectName string   `json:"project_name,omitempty"`
	Count       int      `json:"count"`
	TotalBytes  int64    `json:"total_bytes"`
	OldestUnix  int64    `json:"oldest_unix,omitempty"`
	Items       []Record `json:"items,omitempty"`
}

// Record is a single attachment included in the deletion plan.
type Record struct {
	AttachmentID int64  `json:"attachment_id"`
	Name         string `json:"name,omitempty"`
	Size         int64  `json:"size"`
	ParentKind   string `json:"parent_kind,omitempty"`
	ParentID     string `json:"parent_id,omitempty"`
	CreatedUnix  int64  `json:"created_unix,omitempty"`
}

// Failure is a single failed delete call.
type Failure struct {
	AttachmentID int64  `json:"attachment_id"`
	ProjectID    int64  `json:"project_id"`
	Error        string `json:"error"`
}

// FromAttachment converts a TestRail attachment into a Record. ParentKind
// is taken from Attachment.InferredEntityType(); ParentID is the matching
// id (string-formatted to preserve the cloud-format EntityID when set).
func FromAttachment(a data.Attachment) Record {
	r := Record{
		AttachmentID: a.ID,
		Name:         firstNonEmpty(a.Name, a.Filename),
		Size:         a.Size,
		ParentKind:   a.InferredEntityType(),
		CreatedUnix:  a.CreatedOn,
	}
	if a.EntityID != "" {
		r.ParentID = a.EntityID
		return r
	}
	switch r.ParentKind {
	case "case":
		r.ParentID = itoa(a.CaseID)
	case "run":
		r.ParentID = itoa(a.RunID)
	case "plan":
		r.ParentID = itoa(a.PlanID)
	case "plan_entry":
		r.ParentID = a.EntryID
	case "result":
		r.ParentID = itoa(a.ResultID)
	case "test":
		r.ParentID = itoa(a.TestID)
	}
	return r
}
