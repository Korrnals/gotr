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
