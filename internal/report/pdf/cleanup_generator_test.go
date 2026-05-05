// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package pdf

import (
	"bytes"
	"testing"
	"time"

	cleanupreport "github.com/Korrnals/gotr/internal/report/cleanup"
)

func TestCleanupGenerator_RenderSmoke(t *testing.T) {
	r := &cleanupreport.Report{
		ID:         "cleanup-smoke",
		SnapshotID: "snap_test",
		Timestamp:  time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
		Server:     "https://tr.example.com",
		GotrVer:    "3.5.1",
		Label:      "smoke",
		User:       "alice",
		CLIArgs:    []string{"attachments", "cleanup", "--all-projects=true"},
		Filters: cleanupreport.Filters{
			AllProjects:  true,
			OlderThan:    "3M",
			EntityTypes:  []string{"result"},
			ScanStrategy: "auto",
		},
		Summary: cleanupreport.Summary{
			TotalSelected: 2,
			BackedUp:      2,
			BackupBytes:   2048,
			Deleted:       2,
		},
		Projects: []cleanupreport.ProjectGroup{{
			ProjectID:   30,
			ProjectName: "Acme",
			Count:       2,
			TotalBytes:  2048,
			Items: []cleanupreport.Record{
				{AttachmentID: 1, Name: "a.png", Size: 1024, ParentKind: "result", ParentID: "555"},
				{AttachmentID: 2, Name: "b.log", Size: 1024, ParentKind: "result", ParentID: "556"},
			},
		}},
	}
	g := NewCleanupGenerator()
	data, err := g.Render(r)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty PDF")
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Errorf("output is not a PDF (no %%PDF- header): %q", data[:min(16, len(data))])
	}
}

func TestCleanupGenerator_RejectsNil(t *testing.T) {
	if _, err := NewCleanupGenerator().Render(nil); err == nil {
		t.Error("expected error for nil report")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
