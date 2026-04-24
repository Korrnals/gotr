package retention

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFile(t *testing.T, path, content string, mtime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !mtime.IsZero() {
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatalf("chtimes: %v", err)
		}
	}
}

func TestCleanupReports_Disabled(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "migrations", "default", "2025-01", "migration-20250101T120000Z-sync_full_p48.md"), "old", time.Now().AddDate(-1, 0, 0))
	res, err := CleanupReports(dir, Policy{Enabled: false, MaxAgeDays: 1}, time.Now())
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if res.Removed != 0 {
		t.Errorf("Removed=%d, want 0 when disabled", res.Removed)
	}
}

func TestCleanupReports_AgePrune(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	old := filepath.Join(dir, "migrations", "default", "2025-01", "migration-20250101T120000Z-sync_full_p48.md")
	fresh := filepath.Join(dir, "migrations", "default", "2026-04", "migration-20260401T120000Z-sync_full_p48.md")
	writeFile(t, old, "old", now.AddDate(0, 0, -120))
	writeFile(t, fresh, "fresh", now.AddDate(0, 0, -5))

	policy := Policy{Enabled: true, MaxAgeDays: 30}
	res, err := CleanupReports(dir, policy, now)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if res.Removed != 1 {
		t.Errorf("Removed=%d, want 1", res.Removed)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("old report should have been deleted, stat err=%v", err)
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Errorf("fresh report must remain, err=%v", err)
	}
}

func TestCleanupReports_DryRun(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	old := filepath.Join(dir, "migrations", "default", "2025-01", "migration-20250101T120000Z-sync_full_p48.md")
	writeFile(t, old, "old", now.AddDate(0, 0, -120))

	res, err := CleanupReports(dir, Policy{Enabled: true, MaxAgeDays: 30, DryRun: true}, now)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if res.Removed != 1 {
		t.Errorf("Removed=%d planned, want 1", res.Removed)
	}
	if _, err := os.Stat(old); err != nil {
		t.Errorf("dry run must not delete: %v", err)
	}
}

func TestCleanupReports_KeepCategory(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	cov := filepath.Join(dir, "coverage", "default", "2025-01", "gotr_migration_20250101.md")
	mig := filepath.Join(dir, "migrations", "default", "2025-01", "migration-20250101T120000Z-sync_full_p48.md")
	writeFile(t, cov, "cov", now.AddDate(0, 0, -120))
	writeFile(t, mig, "m", now.AddDate(0, 0, -120))

	res, err := CleanupReports(dir, Policy{Enabled: true, MaxAgeDays: 30, KeepCategories: []string{"coverage"}}, now)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if res.Removed != 1 {
		t.Errorf("Removed=%d, want 1 (migration only)", res.Removed)
	}
	if _, err := os.Stat(cov); err != nil {
		t.Errorf("coverage must be kept: %v", err)
	}
}

func TestCleanupReports_MaxCount(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	for i := 0; i < 5; i++ {
		p := filepath.Join(dir, "migrations", "default", "2026-04", "migration-2026040"+string(rune('1'+i))+"T120000Z-sync_full_p48.md")
		writeFile(t, p, "m", now.AddDate(0, 0, -i))
	}
	res, err := CleanupReports(dir, Policy{Enabled: true, MaxCount: 2}, now)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if res.Removed != 3 {
		t.Errorf("Removed=%d, want 3 (keep newest 2 of 5)", res.Removed)
	}
	if res.Kept != 2 {
		t.Errorf("Kept=%d, want 2", res.Kept)
	}
}

func TestCleanupExports_AgePrune(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	old := filepath.Join(dir, "snaps", "old-bundle.tar.gz")
	fresh := filepath.Join(dir, "reports", "new-bundle.zip")
	writeFile(t, old, "o", now.AddDate(0, 0, -60))
	writeFile(t, fresh, "f", now.AddDate(0, 0, -1))

	res, err := CleanupExports(dir, Policy{Enabled: true, MaxAgeDays: 30}, now)
	if err != nil {
		t.Fatalf("cleanup exports: %v", err)
	}
	if res.Removed != 1 {
		t.Errorf("Removed=%d, want 1", res.Removed)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("old export must be deleted")
	}
}
