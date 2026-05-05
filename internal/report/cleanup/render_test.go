// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	cleanupsvc "github.com/Korrnals/gotr/internal/cleanup"
	"github.com/Korrnals/gotr/internal/models/data"
)

// fixedReport returns a deterministic Report covering every section the
// renderers exercise.
func fixedReport(t *testing.T) *Report {
	t.Helper()
	ts := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	created := time.Date(2026, 1, 15, 9, 30, 0, 0, time.UTC).Unix()
	older := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC).Unix()
	plan := &cleanupsvc.Plan{
		TotalCount: 3,
		TotalBytes: 4096,
		Projects: []cleanupsvc.ProjectSelection{{
			ProjectID:   30,
			ProjectName: "Acme",
			TotalBytes:  3072,
			OldestUnix:  older,
			Attachments: []data.Attachment{
				{ID: 100, Name: "screen.png", Size: 1024, ResultID: 555, CreatedOn: created, ProjectID: 30},
				{ID: 101, Name: "trace.log", Size: 2048, ResultID: 556, CreatedOn: created, ProjectID: 30},
			},
		}, {
			ProjectID:   31,
			ProjectName: "Beta | Pipe", // tests pipe-escape in markdown
			TotalBytes:  1024,
			OldestUnix:  older,
			Attachments: []data.Attachment{
				{ID: 200, Name: "x.bin", Size: 1024, EntityType: "case", EntityID: "c-42", CreatedOn: created, ProjectID: 31},
			},
		}},
	}
	res := &cleanupsvc.ExecuteResult{
		SnapshotID:   "snap_abc",
		BackedUp:     3,
		BackupBytes:  4096,
		Deleted:      2,
		DeleteErrors: 1,
		Failures: []cleanupsvc.DeleteFailure{
			{AttachmentID: 101, ProjectID: 30, Error: "boom"},
		},
	}
	return Build(BuildInput{
		Plan:         plan,
		Result:       res,
		Timestamp:    ts,
		Server:       "https://tr.example.com",
		GotrVer:      "3.5.1",
		Label:        "audit-2026-05",
		User:         "alice",
		CLIArgs:      []string{"attachments", "cleanup", "--all-projects=true"},
		AllProjects:  true,
		OlderThanRaw: "3M",
		CutoffUnix:   time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC).Unix(),
		EntityTypes:  []string{"result", "case"},
		ScanStrategy: "auto",
		Limit:        100,
		ID:           "cleanup-fixed",
	})
}

func TestRenderMarkdown_StableSections(t *testing.T) {
	r := fixedReport(t)
	got := RenderMarkdown(r)

	mustContain := []string{
		"# Attachments Cleanup Report",
		"## Run",
		"| Report ID | `cleanup-fixed` |",
		"| Snapshot ID | `snap_abc` |",
		"## Filters",
		"| Older than | 3M |",
		"| Entity types | case, result |", // sorted by Build
		"## Summary",
		"| Total selected | 3 |",
		"| Deleted | 2 |",
		"| Failed | 1 |",
		"## Per-project breakdown",
		"| 30 | Acme | 2 | 3.00 KB |",
		`| 31 | Beta \| Pipe | 1 | 1.00 KB |`, // pipe escape
		"## Deleted attachments",
		"| 30 | 100 | screen.png |",
		"| 31 | 200 | x.bin | 1.00 KB | case:c-42 |",
		"## Failures",
		"| 101 | 30 | boom |",
		"## Rollback",
		"gotr snap rollback snap_abc",
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Errorf("markdown missing %q\n--- output ---\n%s", want, got)
		}
	}
}

func TestRenderMarkdown_DryRunMarker(t *testing.T) {
	r := fixedReport(t)
	r.DryRun = true
	r.SnapshotID = ""
	got := RenderMarkdown(r)
	if !strings.Contains(got, "# Attachments Cleanup Report (DRY-RUN)") {
		t.Errorf("expected DRY-RUN title; got %q", firstLine(got))
	}
	if !strings.Contains(got, "**DRY-RUN**") {
		t.Error("expected DRY-RUN callout")
	}
	if strings.Contains(got, "gotr snap rollback") {
		t.Error("dry-run markdown must not advertise a rollback command")
	}
}

func TestRenderJSON_RoundTrip(t *testing.T) {
	r := fixedReport(t)
	b, err := RenderJSON(r)
	if err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	if !strings.HasSuffix(string(b), "\n") {
		t.Error("JSON output must end with newline")
	}
	var back Report
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.SnapshotID != r.SnapshotID {
		t.Errorf("snapshot id mismatch: %q vs %q", back.SnapshotID, r.SnapshotID)
	}
	if back.Summary.Deleted != 2 {
		t.Errorf("deleted=%d want 2", back.Summary.Deleted)
	}
	if len(back.Projects) != 2 {
		t.Errorf("projects=%d want 2", len(back.Projects))
	}
	if len(back.Failures) != 1 || back.Failures[0].Error != "boom" {
		t.Errorf("failures lost: %+v", back.Failures)
	}
}

func TestRenderCSV_HeaderAndRows(t *testing.T) {
	r := fixedReport(t)
	b, err := RenderCSV(r)
	if err != nil {
		t.Fatalf("RenderCSV: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	wantHeader := "project_id,project_name,attachment_id,name,size_bytes,parent_kind,parent_id,created_unix,created_utc,deleted,dry_run,snapshot_id"
	if lines[0] != wantHeader {
		t.Errorf("csv header mismatch:\n got: %s\nwant: %s", lines[0], wantHeader)
	}
	// 1 header + 3 attachment rows
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d:\n%s", len(lines), b)
	}
	if !strings.Contains(lines[1], ",100,screen.png,1024,result,") {
		t.Errorf("row 1 unexpected: %s", lines[1])
	}
	if !strings.Contains(lines[3], ",200,x.bin,1024,case,c-42,") {
		t.Errorf("row 3 unexpected: %s", lines[3])
	}
	// deleted=true (not dry-run)
	if !strings.Contains(lines[1], ",true,false,snap_abc") {
		t.Errorf("row 1 missing deleted/dry-run/snap suffix: %s", lines[1])
	}
}

func TestRenderCSV_DryRunDeletedFalse(t *testing.T) {
	r := fixedReport(t)
	r.DryRun = true
	r.SnapshotID = ""
	b, _ := RenderCSV(r)
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n")[1:] {
		if !strings.HasSuffix(line, ",false,true,") {
			t.Errorf("dry-run row must end with ,false,true,<empty> — got %q", line)
		}
	}
}

func TestBuild_DryRunPropagates(t *testing.T) {
	plan := &cleanupsvc.Plan{TotalCount: 1, TotalBytes: 10, Projects: []cleanupsvc.ProjectSelection{{
		ProjectID:   1,
		Attachments: []data.Attachment{{ID: 1, Size: 10, ResultID: 2}},
	}}}
	res := &cleanupsvc.ExecuteResult{DryRun: true, BackedUp: 1, BackupBytes: 10}
	rep := Build(BuildInput{Plan: plan, Result: res})
	if !rep.DryRun {
		t.Error("DryRun not propagated")
	}
	if rep.Summary.TotalSelected != 1 {
		t.Errorf("TotalSelected=%d want 1", rep.Summary.TotalSelected)
	}
	if rep.Summary.BackupBytes != 10 {
		t.Errorf("BackupBytes=%d", rep.Summary.BackupBytes)
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
