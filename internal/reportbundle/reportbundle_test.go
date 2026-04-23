package reportbundle

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func seedReports(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("content of "+n), 0o644); err != nil {
			t.Fatalf("seed %s: %v", n, err)
		}
	}
	return dir
}

func TestExportSingle_PlainCopy(t *testing.T) {
	src := seedReports(t, "migration-2026-01-01.md")
	dest := filepath.Join(t.TempDir(), "copy.md")
	res, err := ExportSingle(src, "migration-2026-01-01.md", dest)
	if err != nil {
		t.Fatalf("ExportSingle: %v", err)
	}
	if res.ArchivePath != dest {
		t.Errorf("archive = %s, want %s", res.ArchivePath, dest)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if !strings.Contains(string(data), "migration-2026-01-01.md") {
		t.Errorf("unexpected copied content: %s", data)
	}
}

func TestExportAll_And_ImportZip_Roundtrip(t *testing.T) {
	src := seedReports(t, "migration-2026-01-01.md", "migration-2026-01-01.pdf", "other.json")
	dest := filepath.Join(t.TempDir(), "bundle.zip")

	res, err := ExportAll(src, dest, ExportOptions{GotrVersion: "test"})
	if err != nil {
		t.Fatalf("ExportAll: %v", err)
	}
	if len(res.Files) != 3 {
		t.Errorf("exported files = %d, want 3", len(res.Files))
	}

	// Import into a fresh reports dir.
	freshReports := t.TempDir()
	imp, err := Import(freshReports, dest, ImportOptions{})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if len(imp.Copied) != 3 {
		t.Errorf("imported = %d, want 3", len(imp.Copied))
	}
	sort.Strings(imp.Copied)
	for _, want := range []string{"migration-2026-01-01.md", "migration-2026-01-01.pdf", "other.json"} {
		p := filepath.Join(freshReports, want)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing imported file %s", want)
		}
	}
}

func TestExportAll_Filter_NarrowsSelection(t *testing.T) {
	src := seedReports(t, "migration-a.md", "migration-b.md", "ignore.log")
	dest := filepath.Join(t.TempDir(), "bundle.zip")
	res, err := ExportAll(src, dest, ExportOptions{GotrVersion: "test", Filter: "migration-*.md"})
	if err != nil {
		t.Fatalf("ExportAll: %v", err)
	}
	if len(res.Files) != 2 {
		t.Errorf("with filter expected 2 files, got %d", len(res.Files))
	}
}

func TestImport_RefusesExistingWithoutOverwrite(t *testing.T) {
	src := seedReports(t, "migration.md")
	dest := filepath.Join(t.TempDir(), "bundle.zip")
	if _, err := ExportAll(src, dest, ExportOptions{GotrVersion: "test"}); err != nil {
		t.Fatalf("ExportAll: %v", err)
	}

	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "migration.md"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("seed existing: %v", err)
	}
	if _, err := Import(target, dest, ImportOptions{}); err == nil {
		t.Fatalf("expected conflict error, got nil")
	}
	if _, err := Import(target, dest, ImportOptions{Overwrite: true}); err != nil {
		t.Fatalf("overwrite should succeed: %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(target, "migration.md"))
	if !strings.Contains(string(got), "content of migration.md") {
		t.Errorf("overwrite did not replace contents: %s", got)
	}
}

func TestImportSingle_DryRun_NoWrite(t *testing.T) {
	src := seedReports(t, "migration.md")
	target := t.TempDir()
	res, err := Import(target, filepath.Join(src, "migration.md"), ImportOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Import dry-run: %v", err)
	}
	if len(res.Copied) != 1 {
		t.Errorf("expected 1 planned copy, got %d", len(res.Copied))
	}
	if _, err := os.Stat(filepath.Join(target, "migration.md")); err == nil {
		t.Errorf("dry-run must not write file")
	}
}
