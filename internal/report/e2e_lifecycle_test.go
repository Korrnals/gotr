package report

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2E_FlatToHierarchy_RoundTrip exercises the full v3.3.0 report
// lifecycle end-to-end: seed a flat ~/.gotr/reports/ layout, detect it,
// migrate it through MigrateFlatLayout, verify the categorized hierarchy,
// reindex, and confirm lookup helpers find the new paths.
//
// This is the single integration test that ties together ClassifyReport,
// IsFlatLayout, MigrateFlatLayout, RecursiveListReports, ResolveReportPath,
// and Reindex.
func TestE2E_FlatToHierarchy_RoundTrip(t *testing.T) {
	base := t.TempDir()

	// 1. Seed a flat layout with one report per category bucket plus
	//    an un-classifiable file and a random sidecar (.log) that must
	//    be ignored.
	files := map[string]string{
		// migration with snapshot  → migrations/default/2026-04
		"migration-20260401T120000Z-sync_20260401T120000_full_p1_to_p2.md": "migration body",
		// coverage tag "foo"       → coverage/foo/2026-04 (tag drives label)
		"gotr_migration_foo_p1_to_p2.md": "coverage body",
		// rollback                 → rollbacks/default/2026-03
		"rollback-20260315T030000Z-sync_abc.md": "rollback body",
		// no-snapshot              → no-snapshot/default/2026-04
		"migration-20260410T080000Z-no_snapshot_p1.md": "no-snapshot body",
		// testrail project dump    → testrail/p7 (YYYYMMDD without T → no month bucket)
		"testrail_cases_p7_20260420.json": "[]",
		// unknown                  → _unclassified
		"weird-report.md": "mystery",
	}
	for name, body := range files {
		p := filepath.Join(base, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	// Non-report sidecar that must NOT be migrated.
	if err := os.WriteFile(filepath.Join(base, "audit.log"), []byte("noise"), 0o644); err != nil {
		t.Fatalf("seed audit.log: %v", err)
	}

	// 2. IsFlatLayout should detect the flat files.
	flat, n, err := IsFlatLayout(base)
	if err != nil {
		t.Fatalf("IsFlatLayout: %v", err)
	}
	if !flat || n != len(files) {
		t.Fatalf("IsFlatLayout = (%v, %d), want (true, %d)", flat, n, len(files))
	}

	// 3. Dry run must produce a plan without touching disk.
	dry, err := MigrateFlatLayout(base, true /*dryRun*/)
	if err != nil {
		t.Fatalf("MigrateFlatLayout dry: %v", err)
	}
	if !dry.DryRun || dry.Moved != 0 {
		t.Fatalf("dry result: DryRun=%v Moved=%d", dry.DryRun, dry.Moved)
	}
	if len(dry.Plans) != len(files) {
		t.Fatalf("dry plans = %d, want %d", len(dry.Plans), len(files))
	}
	// Files must still be at the root after dry run.
	for name := range files {
		if _, err := os.Stat(filepath.Join(base, name)); err != nil {
			t.Errorf("dry-run removed %s: %v", name, err)
		}
	}

	// 4. Apply for real.
	res, err := MigrateFlatLayout(base, false)
	if err != nil {
		t.Fatalf("MigrateFlatLayout apply: %v", err)
	}
	if res.Moved != len(files) {
		t.Fatalf("moved = %d, want %d", res.Moved, len(files))
	}

	// 5. Flat files must be gone from the root; audit.log must survive.
	for name := range files {
		if _, err := os.Stat(filepath.Join(base, name)); !os.IsNotExist(err) {
			t.Errorf("flat file %s still present after migrate: err=%v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(base, "audit.log")); err != nil {
		t.Errorf("audit.log must not be touched: %v", err)
	}

	// 6. Expected categorized destinations exist.
	expectedRel := []string{
		"migrations/default/2026-04-01/migration-20260401T120000Z-sync_20260401T120000_full_p1_to_p2.md",
		"coverage/foo/gotr_migration_foo_p1_to_p2.md",
		"rollbacks/default/2026-03-15/rollback-20260315T030000Z-sync_abc.md",
		"no-snapshot/default/2026-04-10/migration-20260410T080000Z-no_snapshot_p1.md",
		"testrail/p7/testrail_cases_p7_20260420.json",
		"_unclassified/weird-report.md",
	}
	for _, rel := range expectedRel {
		if _, err := os.Stat(filepath.Join(base, filepath.FromSlash(rel))); err != nil {
			t.Errorf("expected categorized file %s: %v", rel, err)
		}
	}

	// 7. IsFlatLayout must now report false.
	flat, n, err = IsFlatLayout(base)
	if err != nil {
		t.Fatalf("IsFlatLayout post: %v", err)
	}
	if flat || n != 0 {
		t.Errorf("IsFlatLayout post-migrate = (%v, %d), want (false, 0)", flat, n)
	}

	// 8. RecursiveListReports must return all migrated files.
	entries, err := RecursiveListReports(base)
	if err != nil {
		t.Fatalf("RecursiveListReports: %v", err)
	}
	if len(entries) != len(files) {
		t.Errorf("RecursiveListReports count = %d, want %d", len(entries), len(files))
	}
	seen := make(map[string]bool, len(entries))
	for _, e := range entries {
		seen[e.Rel] = true
	}
	for _, rel := range expectedRel {
		if !seen[rel] {
			t.Errorf("RecursiveListReports missing %s", rel)
		}
	}

	// 9. Reindex produces INDEX.md at the root referencing moved files.
	if err := Reindex(base); err != nil {
		t.Fatalf("Reindex: %v", err)
	}
	idx, err := os.ReadFile(filepath.Join(base, "INDEX.md"))
	if err != nil {
		t.Fatalf("read INDEX.md: %v", err)
	}
	for _, rel := range expectedRel {
		if !strings.Contains(string(idx), filepath.Base(rel)) {
			t.Errorf("INDEX.md missing reference to %s", rel)
		}
	}

	// 10. ResolveReportPath: by basename, by "latest", by relative path.
	byBase, err := ResolveReportPath(base, "rollback-20260315T030000Z-sync_abc.md")
	if err != nil {
		t.Fatalf("resolve by basename: %v", err)
	}
	if filepath.Base(byBase) != "rollback-20260315T030000Z-sync_abc.md" {
		t.Errorf("basename resolve → %s", byBase)
	}

	latest, err := ResolveReportPath(base, "latest")
	if err != nil {
		t.Fatalf("resolve latest: %v", err)
	}
	if _, err := os.Stat(latest); err != nil {
		t.Errorf("latest path invalid: %v", err)
	}

	byRel, err := ResolveReportPath(base, "coverage/foo/gotr_migration_foo_p1_to_p2.md")
	if err != nil {
		t.Fatalf("resolve by rel: %v", err)
	}
	if filepath.Base(byRel) != "gotr_migration_foo_p1_to_p2.md" {
		t.Errorf("rel resolve → %s", byRel)
	}

	// 11. Re-running MigrateFlatLayout is idempotent.
	again, err := MigrateFlatLayout(base, false)
	if err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if again.Moved != 0 {
		t.Errorf("second migrate moved = %d, want 0", again.Moved)
	}
}

// TestE2E_EmptyReportsDir_IsNotFlat guards against false positives when the
// reports root exists but is empty or contains only subdirectories.
func TestE2E_EmptyReportsDir_IsNotFlat(t *testing.T) {
	base := t.TempDir()
	// Only a subdirectory, no flat files.
	if err := os.MkdirAll(filepath.Join(base, "migrations", "default", "2026-04"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	flat, n, err := IsFlatLayout(base)
	if err != nil {
		t.Fatalf("IsFlatLayout: %v", err)
	}
	if flat || n != 0 {
		t.Errorf("IsFlatLayout on subdir-only = (%v, %d), want (false, 0)", flat, n)
	}

	// RecursiveListReports must tolerate an empty tree.
	entries, err := RecursiveListReports(base)
	if err != nil {
		t.Fatalf("RecursiveListReports: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(entries))
	}

	// ResolveReportPath("latest") on empty tree is ErrNotExist.
	if _, err := ResolveReportPath(base, "latest"); err == nil {
		t.Errorf("expected error on empty tree")
	} else if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected fs.ErrNotExist, got %v", err)
	}
}
