package exportsorg

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name  string
		dir   bool
		want  Category
	}{
		{"snap_abc_20260101.tar.gz", false, CategorySnaps},
		{"reports_20260101.zip", false, CategoryReports},
		{"report_foo.pdf", false, CategoryReports},
		{"report_foo.md", false, CategoryReports},
		{"runs_foo.json", false, CategoryReports},
		{"cases", true, CategoryAPI},
		{"snaps", true, CategoryUnknown},
		{"reports", true, CategoryUnknown},
		{"api", true, CategoryUnknown},
		{"README.txt", false, CategoryUnknown},
	}
	for _, tc := range cases {
		got := Classify(tc.name, tc.dir)
		if got != tc.want {
			t.Errorf("Classify(%q, dir=%v) = %q, want %q", tc.name, tc.dir, got, tc.want)
		}
	}
}

func TestMigrateExportsLayout_DryRun(t *testing.T) {
	base := t.TempDir()
	writeFile(t, filepath.Join(base, "snap_x_20260101.tar.gz"))
	writeFile(t, filepath.Join(base, "reports_20260101.zip"))
	if err := os.MkdirAll(filepath.Join(base, "cases"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := MigrateExportsLayout(base, true)
	if err != nil {
		t.Fatalf("MigrateExportsLayout: %v", err)
	}
	if !res.DryRun || res.Moved != 0 {
		t.Fatalf("dry-run should not move; got %+v", res)
	}
	if len(res.Plans) != 3 {
		t.Fatalf("expected 3 plans, got %d: %+v", len(res.Plans), res.Plans)
	}
	// Original files still in place
	if _, err := os.Stat(filepath.Join(base, "snap_x_20260101.tar.gz")); err != nil {
		t.Fatalf("snap should remain: %v", err)
	}
}

func TestMigrateExportsLayout_Apply(t *testing.T) {
	base := t.TempDir()
	writeFile(t, filepath.Join(base, "snap_x_20260101.tar.gz"))
	writeFile(t, filepath.Join(base, "reports_20260101.zip"))
	writeFile(t, filepath.Join(base, "report_one.pdf"))
	if err := os.MkdirAll(filepath.Join(base, "cases"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(base, "cases", "inner.json"))

	res, err := MigrateExportsLayout(base, false)
	if err != nil {
		t.Fatalf("MigrateExportsLayout: %v", err)
	}
	if res.Moved != 4 {
		t.Fatalf("expected 4 moves, got %d: %+v", res.Moved, res)
	}
	// Verify destinations.
	checks := []string{
		filepath.Join(base, "snaps", "snap_x_20260101.tar.gz"),
		filepath.Join(base, "reports", "reports_20260101.zip"),
		filepath.Join(base, "reports", "report_one.pdf"),
		filepath.Join(base, "api", "cases", "inner.json"),
	}
	for _, p := range checks {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing moved file %s: %v", p, err)
		}
	}
	// Original locations gone.
	if _, err := os.Stat(filepath.Join(base, "snap_x_20260101.tar.gz")); !os.IsNotExist(err) {
		t.Errorf("original snap should be removed, got err=%v", err)
	}
}

func TestMigrateExportsLayout_SkipsExistingTarget(t *testing.T) {
	base := t.TempDir()
	writeFile(t, filepath.Join(base, "snap_x_20260101.tar.gz"))
	// Pre-existing destination — must be preserved, source must remain.
	writeFile(t, filepath.Join(base, "snaps", "snap_x_20260101.tar.gz"))

	res, err := MigrateExportsLayout(base, false)
	if err != nil {
		t.Fatalf("MigrateExportsLayout: %v", err)
	}
	if res.Moved != 0 || res.Skipped != 1 {
		t.Fatalf("expected 0 moved / 1 skipped, got %+v", res)
	}
	if _, err := os.Stat(filepath.Join(base, "snap_x_20260101.tar.gz")); err != nil {
		t.Errorf("source must remain when target exists: %v", err)
	}
}

func TestMigrateExportsLayout_Idempotent(t *testing.T) {
	base := t.TempDir()
	writeFile(t, filepath.Join(base, "snap_x_20260101.tar.gz"))

	if _, err := MigrateExportsLayout(base, false); err != nil {
		t.Fatal(err)
	}
	res, err := MigrateExportsLayout(base, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Moved != 0 {
		t.Fatalf("second run must be no-op, got %+v", res)
	}
}

func TestMigrateExportsLayout_MissingBaseDir(t *testing.T) {
	base := filepath.Join(t.TempDir(), "does-not-exist")
	res, err := MigrateExportsLayout(base, false)
	if err != nil {
		t.Fatalf("missing dir should not error, got %v", err)
	}
	if res.Moved != 0 || len(res.Plans) != 0 {
		t.Fatalf("missing dir should yield empty result, got %+v", res)
	}
}

// When a legacy resource directory (e.g. `templates/`) conflicts with an
// already-categorized destination (`api/templates/`), MigrateExportsLayout
// must merge file-by-file without overwriting conflicts, then remove the
// orphan legacy directory if empty.
func TestMigrateExportsLayout_MergesDirIntoExisting(t *testing.T) {
	base := t.TempDir()

	// Legacy content at the root.
	writeFile(t, filepath.Join(base, "templates", "old_a.json"))
	writeFile(t, filepath.Join(base, "templates", "shared.json"))
	// Pre-existing destination from new-layout writes.
	writeFile(t, filepath.Join(base, "api", "templates", "new_b.json"))
	writeFile(t, filepath.Join(base, "api", "templates", "shared.json")) // conflict

	res, err := MigrateExportsLayout(base, false)
	if err != nil {
		t.Fatalf("MigrateExportsLayout: %v", err)
	}
	if res.Merged != 1 || res.Moved != 0 {
		t.Fatalf("expected merged=1 moved=0, got %+v", res)
	}
	p := res.Plans[0]
	if p.Action != ActionPartial {
		t.Fatalf("expected ActionPartial (shared.json conflict), got %q", p.Action)
	}
	if p.MergedFiles != 1 || p.SkippedFiles != 1 {
		t.Fatalf("expected merged=1 skipped=1 per-plan, got merged=%d skipped=%d", p.MergedFiles, p.SkippedFiles)
	}
	// Non-conflicting legacy file moved to destination.
	if _, err := os.Stat(filepath.Join(base, "api", "templates", "old_a.json")); err != nil {
		t.Errorf("old_a.json should be merged into api/templates: %v", err)
	}
	// Pre-existing file remains untouched.
	if _, err := os.Stat(filepath.Join(base, "api", "templates", "new_b.json")); err != nil {
		t.Errorf("new_b.json must be preserved: %v", err)
	}
	// Conflict kept in legacy dir; orphan directory still exists.
	if _, err := os.Stat(filepath.Join(base, "templates", "shared.json")); err != nil {
		t.Errorf("conflicting shared.json should remain in legacy templates/: %v", err)
	}
}

// When the legacy directory has no conflicts, it must be fully drained and
// removed after merge, with action reported as ActionMerged.
func TestMigrateExportsLayout_MergesDirFully(t *testing.T) {
	base := t.TempDir()

	writeFile(t, filepath.Join(base, "templates", "only_legacy.json"))
	writeFile(t, filepath.Join(base, "api", "templates", "only_new.json"))

	res, err := MigrateExportsLayout(base, false)
	if err != nil {
		t.Fatalf("MigrateExportsLayout: %v", err)
	}
	if res.Merged != 1 {
		t.Fatalf("expected merged=1, got %+v", res)
	}
	if res.Plans[0].Action != ActionMerged {
		t.Fatalf("expected ActionMerged, got %q", res.Plans[0].Action)
	}
	// Legacy dir must be gone.
	if _, err := os.Stat(filepath.Join(base, "templates")); !os.IsNotExist(err) {
		t.Errorf("legacy templates/ should be removed after full merge, stat err=%v", err)
	}
	// Both files live under api/templates.
	for _, f := range []string{"only_legacy.json", "only_new.json"} {
		if _, err := os.Stat(filepath.Join(base, "api", "templates", f)); err != nil {
			t.Errorf("expected %s under api/templates: %v", f, err)
		}
	}
}
