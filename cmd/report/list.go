package report

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Korrnals/gotr/internal/paths"
	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/Korrnals/gotr/internal/state"
	"github.com/Korrnals/gotr/internal/warnings"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local migration reports",
		RunE: func(cmd *cobra.Command, _ []string) error {
			reportsDir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("report list: resolve reports dir: %w", err)
			}

			entries, err := intreport.RecursiveListReports(reportsDir)
			if err != nil {
				return fmt.Errorf("report list: walk reports dir: %w", err)
			}

			filter, _ := cmd.Flags().GetString("filter")
			filtered := entries[:0:0]
			for _, e := range entries {
				if filter != "" {
					match, err := filepath.Match(filter, e.Basename)
					if err != nil {
						return fmt.Errorf("report list: bad filter %q: %w", filter, err)
					}
					if !match && !strings.Contains(e.Rel, filter) {
						continue
					}
				}
				filtered = append(filtered, e)
			}
			if len(filtered) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No reports found")
				maybeFlatLayoutHint(cmd, reportsDir)
				return nil
			}

			limit, _ := cmd.Flags().GetInt("limit")
			if limit <= 0 || limit > len(filtered) {
				limit = len(filtered)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Migration reports (%d):\n", len(filtered))
			for i := 0; i < limit; i++ {
				e := filtered[i]
				fmt.Fprintf(cmd.OutOrStdout(), "  %d. %s\n", i+1, e.Rel)
			}
			maybeFlatLayoutHint(cmd, reportsDir)
			return nil
		},
	}

	cmd.Flags().Int("limit", 20, "Max number of reports to show")
	cmd.Flags().String("filter", "", "Glob applied to report basenames (or substring of relative path)")
	return cmd
}

// maybeFlatLayoutHint prints a one-time hint on stderr when the reports
// directory still contains legacy flat files. It is best-effort and never
// returns an error.
//
// The hint is routed through internal/warnings (key "flat_layout") so it
// honors ui.suppress_warnings / --show-warnings, and through
// internal/state so the hint is shown at most once per installation
// (state.FlatLayoutWarned persists in ~/.gotr/state.json).
func maybeFlatLayoutHint(cmd *cobra.Command, reportsDir string) {
	flat, n, err := intreport.IsFlatLayout(reportsDir)
	if err != nil || !flat {
		return
	}
	if warnings.Suppressed(warnings.KeyFlatLayout) {
		return
	}
	// Persistent one-time guard: skip if the user has already seen it.
	if st, err := state.Load(); err == nil && st.FlatLayoutWarned {
		return
	}
	warnings.Emitf(cmd.ErrOrStderr(), warnings.KeyFlatLayout,
		"~/.gotr/reports/ contains %d legacy flat file(s). Run 'gotr report organize --dry-run' "+
			"to preview the new categorized hierarchy, then re-run without --dry-run to migrate.", n)
	// Best-effort persistence; ignore write errors so transient FS issues
	// never block the user's primary command.
	if st, err := state.Load(); err == nil {
		st.FlatLayoutWarned = true
		_ = state.Save(st)
	}
}
