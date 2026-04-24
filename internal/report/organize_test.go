package report

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateFlatLayout_DryRun(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "migration-20260101T120000Z-sync_full_p48.md"), "m")
	mustWrite(t, filepath.Join(dir, "rollback-20260101T120500Z-suite_1.md"), "r")
	mustWrite(t, filepath.Join(dir, "INDEX.md"), "# idx")

	res, err := MigrateFlatLayout(dir, true)
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if !res.DryRun {
		t.Error("DryRun flag must be true")
	}
	if len(res.Plans) != 2 {
		t.Fatalf("plans=%d, want 2: %+v", len(res.Plans), res.Plans)
	}
	// Files must NOT have been moved.
	if _, err := os.Stat(filepath.Join(dir, "migration-20260101T120000Z-sync_full_p48.md")); err != nil {
		t.Error("dry run must not move files")
	}
}

func TestMigrateFlatLayout_Apply(t *testing.T) {
	dir := t.TempDir()
	mig := filepath.Join(dir, "migration-20260101T120000Z-sync_full_p48.md")
	rb := filepath.Join(dir, "rollback-20260101T120500Z-suite_1.md")
	mustWrite(t, mig, "m")
	mustWrite(t, rb, "r")

	res, err := MigrateFlatLayout(dir, false)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if res.Moved != 2 {
		t.Errorf("moved=%d, want 2", res.Moved)
	}
	// Files must be under categorized hierarchy now.
	for _, want := range []string{
		filepath.Join(dir, "migrations", "default", "2026-01", "migration-20260101T120000Z-sync_full_p48.md"),
		filepath.Join(dir, "rollbacks", "default", "2026-01", "rollback-20260101T120500Z-suite_1.md"),
	} {
		if _, err := os.Stat(want); err != nil {
			t.Errorf("expected %s, err=%v", want, err)
		}
	}
	// Originals gone.
	if _, err := os.Stat(mig); err == nil {
		t.Error("expected original migration file to be moved")
	}
	// INDEX.md regenerated.
	if _, err := os.Stat(filepath.Join(dir, "INDEX.md")); err != nil {
		t.Errorf("INDEX.md should be regenerated: %v", err)
	}
}

func TestMigrateFlatLayout_Idempotent(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "migrations", "default", "2026-01", "migration-20260101T120000Z-sync_full_p48.md"), "m")

	res, err := MigrateFlatLayout(dir, false)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if res.Moved != 0 {
		t.Errorf("moved=%d, want 0 (already organized)", res.Moved)
	}
}
