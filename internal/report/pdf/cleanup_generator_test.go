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

// TestCleanupGenerator_HandlesUnicodeAndInvalidUTF8 exercises the
// rune-aware cell wrapper. Previously fpdf's byte-based SplitLines
// could split Cyrillic strings inside a multi-byte UTF-8 sequence and
// then panic in utf8toutf16 ("index out of range"). The names below
// mix long Cyrillic text with an explicit invalid UTF-8 tail to
// guard against regression.
func TestCleanupGenerator_HandlesUnicodeAndInvalidUTF8(t *testing.T) {
	longCyrillic := "Очень длинное название вложения с кириллицей которое заведомо не помещается в одну строку колонки"
	invalidTail := "имя_файла_" + string([]byte{0xD0}) // dangling 2-byte UTF-8 starter
	r := &cleanupreport.Report{
		ID:         "cleanup-unicode",
		SnapshotID: "snap_unicode",
		Timestamp:  time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
		Server:     "https://tr.example.com",
		GotrVer:    "3.6.0",
		Filters: cleanupreport.Filters{
			AllProjects: true,
			OlderThan:   "3M",
			EntityTypes: []string{"result"},
		},
		Summary: cleanupreport.Summary{TotalSelected: 2, BackedUp: 2, Deleted: 2},
		Projects: []cleanupreport.ProjectGroup{{
			ProjectID:   34,
			ProjectName: "Проект Логистика — большое имя для проверки переноса",
			Count:       2,
			TotalBytes:  4096,
			Items: []cleanupreport.Record{
				{AttachmentID: 1, Name: longCyrillic, Size: 2048, ParentKind: "result", ParentID: "555"},
				{AttachmentID: 2, Name: invalidTail, Size: 2048, ParentKind: "result", ParentID: "556"},
			},
		}},
	}
	g := NewCleanupGenerator()
	data, err := g.Render(r)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("output is not a PDF")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
