package bundlecmd

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Korrnals/gotr/internal/paths"
	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/spf13/cobra"
)

// completeExportReportArg completes the positional argument for
// `gotr export report <filename|all>`. It offers the token "all" plus
// relative paths/basenames of reports under ~/.gotr/reports/.
func completeExportReportArg(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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

	add("all")
	for _, e := range entries {
		add(e.Rel)
		add(e.Basename)
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

// completeExportSnapArg completes the positional argument for
// `gotr export snap <snap_id|label>` using the snapshot manifest.
func completeExportSnapArg(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	store, err := snap.NewStore()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	manifest, err := snap.LoadManifest(store)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	seen := make(map[string]struct{}, len(manifest.Entries)*2)
	out := make([]string, 0, len(manifest.Entries)*2)
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

	for _, e := range manifest.Entries {
		add(e.ID)
		add(e.Label)
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

// completeImportSnapArg suggests bundle files under ~/.gotr/exports/snaps/
// (fallbacks to default filesystem completion filtered to .tar.gz/.tgz).
func completeImportSnapArg(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	dir, err := paths.ExportsSnapsDirPath()
	if err != nil {
		return []string{"tar.gz", "tgz"}, cobra.ShellCompDirectiveFilterFileExt
	}
	out := collectFilesByExt(dir, toComplete, ".tar.gz", ".tgz")
	if len(out) == 0 {
		return []string{"tar.gz", "tgz"}, cobra.ShellCompDirectiveFilterFileExt
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

// completeImportReportArg suggests bundle/report files under
// ~/.gotr/exports/reports/ (zip/pdf/md/json).
func completeImportReportArg(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	dir, err := paths.ExportsReportsDirPath()
	if err != nil {
		return []string{"zip", "pdf", "md", "json"}, cobra.ShellCompDirectiveFilterFileExt
	}
	out := collectFilesByExt(dir, toComplete, ".zip", ".pdf", ".md", ".json")
	if len(out) == 0 {
		return []string{"zip", "pdf", "md", "json"}, cobra.ShellCompDirectiveFilterFileExt
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

// collectFilesByExt walks baseDir recursively and returns regular files whose
// extension (case-insensitive) matches any of the supplied allowed extensions.
// The suggestions are returned as absolute paths so that shell completion
// works from any working directory. If baseDir is missing, an empty slice is
// returned.
func collectFilesByExt(baseDir, toComplete string, allowed ...string) []string {
	allowSet := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		allowSet[strings.ToLower(a)] = struct{}{}
	}
	var out []string
	_ = filepath.WalkDir(baseDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // tolerate missing/unreadable paths during completion
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		// Handle multi-dot ".tar.gz" first.
		ext := ""
		if strings.HasSuffix(name, ".tar.gz") {
			ext = ".tar.gz"
		} else {
			ext = strings.ToLower(filepath.Ext(name))
		}
		if _, ok := allowSet[ext]; !ok {
			return nil
		}
		if toComplete != "" && !strings.HasPrefix(p, toComplete) {
			return nil
		}
		out = append(out, p)
		return nil
	})
	return out
}
