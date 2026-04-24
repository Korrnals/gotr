package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// setHome redirects HOME and returns the reports dir path inside that
// temporary home. Callers are responsible for populating files under the
// returned dir.
func setHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	reportsDir := filepath.Join(tmp, ".gotr", "reports")
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		t.Fatalf("mkdir reports: %v", err)
	}
	return reportsDir
}

func TestCompleteReportArg_ReturnsLatestAndNames(t *testing.T) {
	reportsDir := setHome(t)
	// flat + categorized entries
	files := []string{
		"migrations/default/2026-04/migration-20260424T120000Z-sync_foo.md",
		"coverage/2026-04/gotr_migration_foo.pdf",
		"legacy_flat_report.md",
	}
	for _, rel := range files {
		full := filepath.Join(reportsDir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	out, directive := completeReportArg(&cobra.Command{}, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want NoFileComp", directive)
	}
	want := map[string]bool{
		"latest":                    false,
		"legacy_flat_report.md":     false,
		"migration-20260424T120000Z-sync_foo.md": false,
		"gotr_migration_foo.pdf":    false,
	}
	for _, s := range out {
		if _, ok := want[s]; ok {
			want[s] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("expected suggestion %q not found in %v", k, out)
		}
	}
}

func TestCompleteReportArg_FiltersByPrefix(t *testing.T) {
	reportsDir := setHome(t)
	for _, rel := range []string{"a.md", "ab.md", "x.md"} {
		if err := os.WriteFile(filepath.Join(reportsDir, rel), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	out, _ := completeReportArg(&cobra.Command{}, nil, "a")
	for _, s := range out {
		if s == "x.md" {
			t.Errorf("prefix filter leaked %q", s)
		}
	}
}

func TestCompleteReportArg_IgnoresExtraArgs(t *testing.T) {
	setHome(t)
	out, directive := completeReportArg(&cobra.Command{}, []string{"already"}, "")
	if len(out) != 0 {
		t.Errorf("expected empty, got %v", out)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v", directive)
	}
}
