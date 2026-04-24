package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Reindex regenerates baseDir/INDEX.md by walking baseDir recursively.
// Entries are grouped by category and labeled with their relative path.
// The index lists the newest 50 reports, matching the prior behavior.
//
// Reindex is safe to call on a missing or empty baseDir: the function is a
// no-op in that case and returns nil.
func Reindex(baseDir string) error {
	entries, err := RecursiveListReports(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("walk reports: %w", err)
	}
	if len(entries) == 0 {
		// Write a minimal index anyway so "gotr report list" has a stable anchor.
		if err := os.MkdirAll(baseDir, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(baseDir, "INDEX.md"),
			[]byte("# Migration Reports Index\n\n_No reports yet._\n"), 0o644)
	}

	const maxIndex = 50
	if len(entries) > maxIndex {
		entries = entries[:maxIndex]
	}

	var sb strings.Builder
	sb.WriteString("# Migration Reports Index\n\n")
	sb.WriteString("Recent reports (newest first):\n\n")
	sb.WriteString("| Date | Category | Report |\n")
	sb.WriteString("|------|----------|--------|\n")
	for _, e := range entries {
		cls := ClassifyReport(e.Basename)
		when := e.ModTime.UTC().Format("2006-01-02 15:04")
		fmt.Fprintf(&sb, "| %s | `%s` | [%s](%s) |\n",
			when, cls.Category, e.Basename, e.Rel)
	}

	return os.WriteFile(filepath.Join(baseDir, "INDEX.md"), []byte(sb.String()), 0o644)
}
