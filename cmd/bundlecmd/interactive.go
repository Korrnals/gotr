package bundlecmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/paths"
	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ErrNoInteractiveTarget signals that interactive selection was requested but
// no candidates exist for the user to pick from.
var ErrNoInteractiveTarget = errors.New("no candidates for interactive selection")

// shouldPrompt reports whether the command should offer an interactive survey
// prompt when positional arguments are missing. A terminal prompter + real
// TTY on stdin are required; `--non-interactive` always wins.
func shouldPrompt(cmd *cobra.Command) bool {
	ctx := cmd.Context()
	if interactive.IsNonInteractive(ctx) {
		return false
	}
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// promptExportReport asks the user to pick a report (or "all") under
// reportsDir for `gotr export report`.
func promptExportReport(cmd *cobra.Command, reportsDir string) (string, error) {
	entries, err := intreport.RecursiveListReports(reportsDir)
	if err != nil {
		return "", fmt.Errorf("list reports: %w", err)
	}
	options := make([]string, 0, len(entries)+1)
	options = append(options, "all")
	for _, e := range entries {
		options = append(options, e.Rel)
	}
	if len(options) == 1 {
		// Only "all" available — caller can still pass it through, but we
		// signal that there is nothing meaningful to pick.
		return "", ErrNoInteractiveTarget
	}
	p := interactive.PrompterFromContext(cmd.Context())
	_, value, err := p.Select("Select report to export", options)
	if err != nil {
		return "", err
	}
	return value, nil
}

// promptExportSnap asks the user to pick a snapshot id from the manifest.
func promptExportSnap(cmd *cobra.Command) (string, error) {
	store, err := snap.NewStore()
	if err != nil {
		return "", fmt.Errorf("snap store: %w", err)
	}
	manifest, err := snap.LoadManifest(store)
	if err != nil {
		return "", fmt.Errorf("snap manifest: %w", err)
	}
	if len(manifest.Entries) == 0 {
		return "", ErrNoInteractiveTarget
	}
	options := make([]string, 0, len(manifest.Entries))
	for _, e := range manifest.Entries {
		label := e.ID
		if e.Label != "" {
			label = e.ID + "  (label: " + e.Label + ")"
		}
		options = append(options, label)
	}
	p := interactive.PrompterFromContext(cmd.Context())
	idx, _, err := p.Select("Select snapshot to export", options)
	if err != nil {
		return "", err
	}
	return manifest.Entries[idx].ID, nil
}

// promptImportBundle asks the user to pick a bundle file under dir with any
// of the given extensions (two-dot `.tar.gz` is handled).
func promptImportBundle(cmd *cobra.Command, dir, label string, exts ...string) (string, error) {
	candidates := listBundleFiles(dir, exts...)
	if len(candidates) == 0 {
		return "", ErrNoInteractiveTarget
	}
	p := interactive.PrompterFromContext(cmd.Context())
	_, value, err := p.Select(label, candidates)
	if err != nil {
		return "", err
	}
	return value, nil
}

// listBundleFiles walks dir (recursively) and returns absolute paths of files
// whose extension matches one of exts. Missing directories yield an empty
// slice rather than an error.
func listBundleFiles(dir string, exts ...string) []string {
	if dir == "" {
		return nil
	}
	allow := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		allow[strings.ToLower(e)] = struct{}{}
	}
	var out []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return nil //nolint:nilerr // tolerate transient walk errors during interactive listing
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		ext := ""
		if strings.HasSuffix(name, ".tar.gz") {
			ext = ".tar.gz"
		} else {
			ext = strings.ToLower(filepath.Ext(name))
		}
		if _, ok := allow[ext]; !ok {
			return nil
		}
		out = append(out, p)
		return nil
	})
	return out
}

// resolveExportReportTarget returns args[0] or runs an interactive picker.
func resolveExportReportTarget(cmd *cobra.Command, args []string, reportsDir string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if !shouldPrompt(cmd) {
		return "", fmt.Errorf("export report: a filename or 'all' is required (pass as argument or run interactively)")
	}
	return promptExportReport(cmd, reportsDir)
}

// resolveExportSnapTarget returns args[0] or runs an interactive snap picker.
func resolveExportSnapTarget(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if !shouldPrompt(cmd) {
		return "", fmt.Errorf("export snap: a snap_id or label is required (pass as argument or run interactively)")
	}
	return promptExportSnap(cmd)
}

// resolveImportSnapTarget returns args[0] or runs an interactive bundle picker.
func resolveImportSnapTarget(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if !shouldPrompt(cmd) {
		return "", fmt.Errorf("import snap: a bundle path is required (pass as argument or run interactively)")
	}
	dir, err := paths.ExportsSnapsDirPath()
	if err != nil {
		return "", err
	}
	return promptImportBundle(cmd, dir, "Select snapshot bundle to import", ".tar.gz", ".tgz")
}

// resolveImportReportTarget returns args[0] or runs an interactive bundle picker.
func resolveImportReportTarget(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if !shouldPrompt(cmd) {
		return "", fmt.Errorf("import report: a bundle or report path is required (pass as argument or run interactively)")
	}
	dir, err := paths.ExportsReportsDirPath()
	if err != nil {
		return "", err
	}
	return promptImportBundle(cmd, dir, "Select report bundle or file to import", ".zip", ".pdf", ".md", ".json")
}
