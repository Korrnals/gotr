// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

// Package cleanup builds a deletion-audit report for `gotr attachments
// cleanup`. The report is emitted in four formats — Markdown, JSON,
// CSV, and PDF — under
//
//	~/.gotr/reports/cleanup-attachments/<label>/<YYYY-MM>/cleanup-<id>-<ts>.<ext>
//
// The Markdown rendition is the human-readable artefact, JSON is the
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

	// Aggregate counters.
	Summary Summary `json:"summary"`

	// Per-project breakdown.
	Projects []ProjectGroup `json:"projects"`

	// Failures (only set when DeleteErrors > 0).
	Failures []Failure `json:"failures,omitempty"`
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
