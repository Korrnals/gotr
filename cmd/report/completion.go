package report

import (
	"strings"

	"github.com/Korrnals/gotr/internal/paths"
	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/spf13/cobra"
)

// completeReportArg provides shell completion for a single positional
// argument that names a report. It offers:
//   - the token "latest" (resolves to the most recent report)
//   - relative paths under ~/.gotr/reports/ (e.g. "migrations/default/2026-04/foo.md")
//   - basenames (for flat-layout reports)
//
// Only entries whose rel-path or basename matches the current prefix are
// returned. Errors collapse to "no completion".
func completeReportArg(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	reportsDir, err := paths.ReportsDirPath()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	entries, err := intreport.RecursiveListReports(reportsDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	seen := make(map[string]struct{}, len(entries)*2+1)
	out := make([]string, 0, len(entries)+1)
	add := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		if toComplete != "" && !strings.HasPrefix(s, toComplete) {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}

	add("latest")
	for _, e := range entries {
		add(e.Rel)
		add(e.Basename)
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}
