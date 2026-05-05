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
			BackedUp:    in.Result.BackedUp,
			BackupBytes: in.Result.BackupBytes,
			Deleted:     in.Result.Deleted,
			Failed:      in.Result.DeleteErrors,
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
	}

	return rep
}

// NewID returns a deterministic report ID derived from the timestamp.
func NewID(ts time.Time) string {
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	return "cleanup-" + ts.UTC().Format("20060102T150405Z")
}
