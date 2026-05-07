package snapbundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/Korrnals/gotr/internal/snap"
)

// TestE2E_OrganizeThenExportImport_WithReports walks the complete v3.3.0
// user journey: a user lands on a fresh machine with a flat reports
// directory, runs `report organize`, then exports a snapshot with
// --with-reports (default) and imports it back into an empty store. The
// categorized report must be discovered by ExportOne's snapID matcher and
// must survive the export/import round-trip.
func TestE2E_OrganizeThenExportImport_WithReports(t *testing.T) {
	// 1. Seed a snap store with a fake sync snapshot.
	store, snapID := newStoreWithSnap(t)
	snapBase := filepath.Base(snapID)

	// 2. Seed a FLAT reports layout with one report whose filename
	//    references the snapID (the matcher uses filepath.Base) and one
	//    unrelated report that must NOT be embedded.
	reportsDir := t.TempDir()
	flatMatch := filepath.Join(reportsDir,
		"migration-20260101T000000Z-"+snapBase+".md")
	flatOther := filepath.Join(reportsDir, "gotr_migration_foo_p1_to_p2.md")
	for _, p := range []string{flatMatch, flatOther} {
		if err := os.WriteFile(p, []byte("data:"+filepath.Base(p)), 0o644); err != nil {
			t.Fatalf("seed %s: %v", p, err)
		}
	}

	// 3. Flat-layout detection + organize.
	flat, n, err := intreport.IsFlatLayout(reportsDir)
	if err != nil {
		t.Fatalf("IsFlatLayout: %v", err)
	}
	if !flat || n != 2 {
		t.Fatalf("IsFlatLayout = (%v, %d), want (true, 2)", flat, n)
	}

	if _, err := intreport.MigrateFlatLayout(reportsDir, false); err != nil {
		t.Fatalf("MigrateFlatLayout: %v", err)
	}

	// Verify the reports were categorized.
	catMatch := filepath.Join(reportsDir, "migrations", "default", "2026-01-01",
		"migration-20260101T000000Z-"+snapBase+".md")
	if _, err := os.Stat(catMatch); err != nil {
		t.Fatalf("expected categorized match report at %s: %v", catMatch, err)
	}

	// 4. Export the snapshot with --with-reports.
	dest := filepath.Join(t.TempDir(), "snap.tar.gz")
	res, err := ExportOne(store, snapID, dest, ExportOptions{
		GotrVersion:    "e2e",
		IncludeReports: true,
		ReportsDir:     reportsDir,
	})
	if err != nil {
		t.Fatalf("ExportOne: %v", err)
	}
	if len(res.IncludedReports) != 1 {
		t.Fatalf("expected exactly 1 embedded report, got %d: %v",
			len(res.IncludedReports), res.IncludedReports)
	}
	embedded := res.IncludedReports[0]
	if !strings.HasPrefix(embedded, "reports/migrations/default/2026-01-01/") {
		t.Errorf("embedded archive path = %q, want reports/migrations/... prefix",
			embedded)
	}
	if !strings.Contains(embedded, snapBase) {
		t.Errorf("embedded report must contain snap basename %s: %s", snapBase, embedded)
	}

	// 5. Import into a fresh empty store, restoring reports into a
	//    scratch reports dir (simulates a different machine).
	freshBase := t.TempDir()
	fresh, err := snap.NewStoreAt(freshBase)
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	freshReports := t.TempDir()
	imp, err := Import(fresh, dest, ImportOptions{ReportsDir: freshReports})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if imp.SnapID != snapID {
		t.Errorf("imported snapID = %q, want %q", imp.SnapID, snapID)
	}
	if !fresh.Exists(snapID) {
		t.Errorf("fresh store missing imported snapshot %s", snapID)
	}
	// The embedded report must land in the scratch reports dir under its
	// categorized path (migrations/default/2026-01-01/<basename>.md).
	if len(imp.IncludedReports) != 1 {
		t.Fatalf("expected 1 restored report, got %d: %v",
			len(imp.IncludedReports), imp.IncludedReports)
	}
	wantRel := "migrations/default/2026-01-01/migration-20260101T000000Z-" + snapBase + ".md"
	if imp.IncludedReports[0] != wantRel {
		t.Errorf("restored report rel = %q, want %q", imp.IncludedReports[0], wantRel)
	}
	restoredAbs := filepath.Join(freshReports, filepath.FromSlash(wantRel))
	if _, err := os.Stat(restoredAbs); err != nil {
		t.Errorf("expected restored report at %s: %v", restoredAbs, err)
	}

	// 6. Re-import must skip (collision-safe; not overwrite).
	imp2, err := Import(fresh, dest, ImportOptions{
		Overwrite: true, ReportsDir: freshReports,
	})
	if err != nil {
		t.Fatalf("second Import: %v", err)
	}
	if len(imp2.SkippedReports) != 1 || imp2.SkippedReports[0] != wantRel {
		t.Errorf("expected 1 skipped report on re-import, got %v", imp2.SkippedReports)
	}
	if len(imp2.IncludedReports) != 0 {
		t.Errorf("expected 0 restored on re-import, got %v", imp2.IncludedReports)
	}
}

// TestImport_SkipReportsOptOut verifies that ImportOptions.SkipReports
// suppresses restoration while still importing the snapshot itself.
func TestImport_SkipReportsOptOut(t *testing.T) {
	store, snapID := newStoreWithSnap(t)
	snapBase := filepath.Base(snapID)

	reportsDir := t.TempDir()
	match := filepath.Join(reportsDir, "migrations", "default", "2026-01-01",
		"migration-20260101T000000Z-"+snapBase+".md")
	if err := os.MkdirAll(filepath.Dir(match), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(match, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	dest := filepath.Join(t.TempDir(), "snap.tar.gz")
	if _, err := ExportOne(store, snapID, dest, ExportOptions{
		IncludeReports: true,
		ReportsDir:     reportsDir,
	}); err != nil {
		t.Fatalf("ExportOne: %v", err)
	}

	fresh, err := snap.NewStoreAt(t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	freshReports := t.TempDir()
	imp, err := Import(fresh, dest, ImportOptions{
		SkipReports: true,
		ReportsDir:  freshReports,
	})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if !fresh.Exists(snapID) {
		t.Fatalf("snapshot must be imported even when SkipReports=true")
	}
	if len(imp.IncludedReports) != 0 {
		t.Errorf("SkipReports=true must not restore reports, got %v", imp.IncludedReports)
	}
	// Verify the reports dir is empty.
	entries, _ := os.ReadDir(freshReports)
	if len(entries) != 0 {
		t.Errorf("expected empty reports dir, got %d entries", len(entries))
	}
}
