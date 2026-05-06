package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClassifyReport(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		explicit  string
		wantCat   Category
		wantLabel string
		wantDate  string
		wantProj  int
	}{
		{
			name:      "migration sync_full with explicit label",
			filename:  "migration-20260424T020240Z-sync_full_suite_42.md",
			explicit:  "q2-run",
			wantCat:   CategoryMigrations,
			wantLabel: "q2-run",
			wantDate:  "2026-04-24",
		},
		{
			name:      "migration sync_full default label",
			filename:  "migration-20260424T020240Z-sync_full_suite_42.pdf",
			wantCat:   CategoryMigrations,
			wantLabel: DefaultLabel,
			wantDate:  "2026-04-24",
		},
		{
			name:      "no snapshot always default label",
			filename:  "migration-20260101T000000Z-no_snapshot.md",
			explicit:  "ignored",
			wantCat:   CategoryNoSnapshot,
			wantLabel: DefaultLabel,
			wantDate:  "2026-01-01",
		},
		{
			name:      "rollback",
			filename:  "rollback-20260215T101010Z-suite_1.md",
			wantCat:   CategoryRollbacks,
			wantLabel: DefaultLabel,
			wantDate:  "2026-02-15",
		},
		{
			name:      "coverage with tag",
			filename:  "gotr_migration_shared-steps_p48_to_p49.pdf",
			wantCat:   CategoryCoverage,
			wantLabel: "shared-steps",
			wantDate:  "",
		},
		{
			name:      "coverage with explicit label overrides tag",
			filename:  "gotr_migration_shared-steps_p48_to_p49.pdf",
			explicit:  "forced",
			wantCat:   CategoryCoverage,
			wantLabel: "forced",
		},
		{
			name:     "testrail raw dump",
			filename: "testrail_plans_p48.json",
			wantCat:  CategoryTestrail,
			wantProj: 48,
		},
		{
			name:     "unclassified",
			filename: "random_note.md",
			wantCat:  CategoryUnclassified,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyReportWithLabel(tt.filename, tt.explicit)
			if got.Category != tt.wantCat {
				t.Errorf("Category = %q, want %q", got.Category, tt.wantCat)
			}
			if tt.wantCat != CategoryTestrail && tt.wantCat != CategoryUnclassified && got.Label != tt.wantLabel {
				t.Errorf("Label = %q, want %q", got.Label, tt.wantLabel)
			}
			if tt.wantDate != "" && got.Date != tt.wantDate {
				t.Errorf("Date = %q, want %q", got.Date, tt.wantDate)
			}
			if tt.wantProj != 0 && got.Project != tt.wantProj {
				t.Errorf("Project = %d, want %d", got.Project, tt.wantProj)
			}
		})
	}
}

func TestClassificationRelDir(t *testing.T) {
	tests := []struct {
		name string
		in   Classification
		want string
	}{
		{
			name: "migrations with label and date",
			in:   Classification{Category: CategoryMigrations, Label: "q2", Date: "2026-04-15"},
			want: filepath.Join("migrations", "q2", "2026-04-15"),
		},
		{
			name: "migrations empty label -> default",
			in:   Classification{Category: CategoryMigrations, Date: "2026-04-15"},
			want: filepath.Join("migrations", DefaultLabel, "2026-04-15"),
		},
		{
			name: "testrail with project",
			in:   Classification{Category: CategoryTestrail, Project: 48, Date: "2026-04-15"},
			want: filepath.Join("testrail", "p48", "2026-04-15"),
		},
		{
			name: "testrail no project",
			in:   Classification{Category: CategoryTestrail, Date: "2026-04-15"},
			want: filepath.Join("testrail", "p0", "2026-04-15"),
		},
		{
			name: "unclassified",
			in:   Classification{Category: CategoryUnclassified},
			want: "_unclassified",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.RelDir(); got != tt.want {
				t.Errorf("RelDir = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRecursiveListReports(t *testing.T) {
	dir := t.TempDir()
	// Build a small hierarchy.
	mustWrite(t, filepath.Join(dir, "INDEX.md"), "# index")
	mustWrite(t, filepath.Join(dir, "migrations", "default", "2026-04", "a.md"), "a")
	mustWrite(t, filepath.Join(dir, "migrations", "q2", "2026-04", "b.pdf"), "b")
	mustWrite(t, filepath.Join(dir, "rollbacks", "default", "c.md"), "c")
	mustWrite(t, filepath.Join(dir, "ignore.txt.bak"), "nope")

	// Force distinct mtimes so the newest-first ordering is deterministic.
	mustChtime(t, filepath.Join(dir, "migrations", "default", "2026-04", "a.md"), time.Now().Add(-2*time.Hour))
	mustChtime(t, filepath.Join(dir, "migrations", "q2", "2026-04", "b.pdf"), time.Now().Add(-1*time.Hour))
	mustChtime(t, filepath.Join(dir, "rollbacks", "default", "c.md"), time.Now())

	entries, err := RecursiveListReports(dir)
	if err != nil {
		t.Fatalf("RecursiveListReports: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3: %+v", len(entries), entries)
	}
	if entries[0].Basename != "c.md" {
		t.Errorf("expected newest c.md first, got %q", entries[0].Basename)
	}
	for _, e := range entries {
		if e.Basename == "INDEX.md" {
			t.Errorf("INDEX.md must be excluded")
		}
	}
}

func TestResolveReportPath(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "migrations", "default", "2026-04", "alpha.md")
	p2 := filepath.Join(dir, "rollbacks", "default", "beta.md")
	mustWrite(t, p1, "a")
	mustWrite(t, p2, "b")
	mustChtime(t, p1, time.Now().Add(-time.Hour))
	mustChtime(t, p2, time.Now())

	// Exact basename.
	got, err := ResolveReportPath(dir, "alpha.md")
	if err != nil || got != p1 {
		t.Errorf("basename: got=%q err=%v, want %q", got, err, p1)
	}
	// Basename without ext.
	got, err = ResolveReportPath(dir, "alpha")
	if err != nil || got != p1 {
		t.Errorf("stem: got=%q err=%v, want %q", got, err, p1)
	}
	// "latest".
	got, err = ResolveReportPath(dir, "latest")
	if err != nil || got != p2 {
		t.Errorf("latest: got=%q err=%v, want %q", got, err, p2)
	}
	// Relative path.
	rel := filepath.Join("rollbacks", "default", "beta.md")
	got, err = ResolveReportPath(dir, rel)
	if err != nil || got != p2 {
		t.Errorf("rel: got=%q err=%v, want %q", got, err, p2)
	}
	// Missing.
	if _, err := ResolveReportPath(dir, "nope.md"); err == nil {
		t.Error("expected error for missing report")
	}
}

func TestIsFlatLayout(t *testing.T) {
	t.Run("missing dir", func(t *testing.T) {
		flat, n, err := IsFlatLayout(filepath.Join(t.TempDir(), "does-not-exist"))
		if err != nil || flat || n != 0 {
			t.Fatalf("missing: flat=%v n=%d err=%v", flat, n, err)
		}
	})
	t.Run("flat root", func(t *testing.T) {
		dir := t.TempDir()
		mustWrite(t, filepath.Join(dir, "migration-20260101T000000Z-sync_full_p48.md"), "x")
		mustWrite(t, filepath.Join(dir, "INDEX.md"), "# idx")
		flat, n, err := IsFlatLayout(dir)
		if err != nil || !flat || n != 1 {
			t.Fatalf("flat=%v n=%d err=%v", flat, n, err)
		}
	})
	t.Run("categorized root", func(t *testing.T) {
		dir := t.TempDir()
		mustWrite(t, filepath.Join(dir, "migrations", "default", "x.md"), "x")
		flat, n, err := IsFlatLayout(dir)
		if err != nil || flat || n != 0 {
			t.Fatalf("flat=%v n=%d err=%v", flat, n, err)
		}
	})
}

func TestReindex(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "migrations", "default", "2026-04", "one.md"), "1")
	mustWrite(t, filepath.Join(dir, "rollbacks", "default", "two.md"), "2")

	if err := Reindex(dir); err != nil {
		t.Fatalf("Reindex: %v", err)
	}
	idx, err := os.ReadFile(filepath.Join(dir, "INDEX.md"))
	if err != nil {
		t.Fatalf("read INDEX.md: %v", err)
	}
	body := string(idx)
	for _, want := range []string{"Migration Reports Index", "one.md", "two.md", "migrations", "rollbacks"} {
		if !strings.Contains(body, want) {
			t.Errorf("INDEX.md missing %q\n---\n%s", want, body)
		}
	}
}

func TestReindexEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := Reindex(dir); err != nil {
		t.Fatalf("Reindex: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "INDEX.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(b), "No reports yet") {
		t.Errorf("expected placeholder, got: %s", b)
	}
}

// --- helpers -----------------------------------------------------------------

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustChtime(t *testing.T, path string, when time.Time) {
	t.Helper()
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatal(err)
	}
}
