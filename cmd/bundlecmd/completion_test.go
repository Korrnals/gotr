package bundlecmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/snap"
	"github.com/spf13/cobra"
)

func setHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("GOTR_HOME", filepath.Join(tmp, ".gotr"))
	return tmp
}

func TestCompleteExportReportArg_IncludesAllAndEntries(t *testing.T) {
	home := setHome(t)
	reportsDir := filepath.Join(home, ".gotr", "reports")
	if err := os.MkdirAll(filepath.Join(reportsDir, "migrations", "default", "2026-04"), 0o755); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(reportsDir, "migrations", "default", "2026-04", "r.md")
	if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, directive := completeExportReportArg(&cobra.Command{}, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v", directive)
	}
	sawAll := false
	sawRel := false
	sawBase := false
	for _, s := range out {
		switch s {
		case "all":
			sawAll = true
		case "migrations/default/2026-04/r.md":
			sawRel = true
		case "r.md":
			sawBase = true
		}
	}
	if !sawAll || !sawRel || !sawBase {
		t.Errorf("missing suggestion: all=%v rel=%v base=%v (got %v)", sawAll, sawRel, sawBase, out)
	}
}

func TestCompleteExportSnapArg_UsesManifest(t *testing.T) {
	setHome(t)
	store, err := snap.NewStore()
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	manifest, err := snap.LoadManifest(store)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	meta := &snap.Meta{
		ID:         "sync/20260424T120000_full_p1_to_p2",
		Label:      "full_p1_to_p2",
		Category:   snap.CatSync,
		Operation:  snap.OpSyncFull,
		EntityType: "cases",
		Status:     snap.StatusAvailable,
		Timestamp:  time.Now(),
	}
	if err := manifest.Add(meta); err != nil {
		t.Fatalf("manifest.Add: %v", err)
	}

	out, directive := completeExportSnapArg(&cobra.Command{}, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v", directive)
	}
	sawID, sawLabel := false, false
	for _, s := range out {
		if s == meta.ID {
			sawID = true
		}
		if s == meta.Label {
			sawLabel = true
		}
	}
	if !sawID || !sawLabel {
		t.Errorf("want id=%v and label=%v in %v", sawID, sawLabel, out)
	}
}

func TestCompleteImportSnapArg_ScansExportsSnaps(t *testing.T) {
	home := setHome(t)
	dir := filepath.Join(home, ".gotr", "exports", "snaps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	good := filepath.Join(dir, "snap_abc.tar.gz")
	bad := filepath.Join(dir, "notes.txt")
	for _, p := range []string{good, bad} {
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	out, _ := completeImportSnapArg(&cobra.Command{}, nil, "")
	sawGood, sawBad := false, false
	for _, s := range out {
		if s == good {
			sawGood = true
		}
		if s == bad {
			sawBad = true
		}
	}
	if !sawGood {
		t.Errorf("expected %q in %v", good, out)
	}
	if sawBad {
		t.Errorf("unexpected %q in %v", bad, out)
	}
}

func TestCompleteImportReportArg_ScansExportsReports(t *testing.T) {
	home := setHome(t)
	dir := filepath.Join(home, ".gotr", "exports", "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	ok := []string{"bundle.zip", "report.pdf", "note.md", "data.json"}
	skip := []string{"ignore.txt", "archive.tar"}
	for _, n := range append(ok, skip...) {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	out, _ := completeImportReportArg(&cobra.Command{}, nil, "")
	seen := map[string]bool{}
	for _, s := range out {
		seen[filepath.Base(s)] = true
	}
	for _, n := range ok {
		if !seen[n] {
			t.Errorf("missing %q in %v", n, out)
		}
	}
	for _, n := range skip {
		if seen[n] {
			t.Errorf("unexpected %q in %v", n, out)
		}
	}
}
