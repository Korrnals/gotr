package retention

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ExportResult mirrors Result but for the exports directory. Entries are
// paths relative to the exports root.
type ExportResult struct {
	Decisions  []ExportDecision
	Removed    int
	Kept       int
	BytesFreed int64
	DryRun     bool
}

// ExportDecision describes the verdict for a single export artifact.
type ExportDecision struct {
	Rel     string
	AbsPath string
	ModTime time.Time
	Keep    bool
	Reason  string
}

// Summary renders a human-friendly multi-line report.
func (r *ExportResult) Summary() string {
	if r == nil {
		return ""
	}
	var sb strings.Builder
	verb := "Removed"
	if r.DryRun {
		verb = "Would remove"
	}
	fmt.Fprintf(&sb, "%s %d export(s), kept %d (freed %d bytes).\n", verb, r.Removed, r.Kept, r.BytesFreed)
	for _, d := range r.Decisions {
		if d.Keep {
			continue
		}
		fmt.Fprintf(&sb, "  - %s  [%s]\n", d.Rel, d.Reason)
	}
	return sb.String()
}

// CleanupExports prunes files under baseDir (recursively) according to
// policy. Only regular files with bundle-like extensions are considered:
// .zip, .tar.gz, .pdf, .md, .json. Empty directories are not removed.
func CleanupExports(baseDir string, policy Policy, now time.Time) (*ExportResult, error) {
	res := &ExportResult{DryRun: policy.DryRun}
	if !policy.Enabled {
		return res, nil
	}

	items, err := listExportFiles(baseDir)
	if err != nil {
		return nil, err
	}

	cutoff := policy.Cutoff(now)
	decisions := make([]ExportDecision, 0, len(items))
	for _, it := range items {
		d := ExportDecision{Rel: it.rel, AbsPath: it.abs, ModTime: it.modTime, Keep: true, Reason: "under_limits"}
		if !cutoff.IsZero() && it.modTime.Before(cutoff) {
			d.Keep = false
			d.Reason = fmt.Sprintf("age>%dd", policy.MaxAgeDays)
		}
		decisions = append(decisions, d)
	}
	if policy.MaxCount > 0 {
		applyExportMaxCount(decisions, policy.MaxCount)
	}

	for _, d := range decisions {
		if d.Keep {
			res.Kept++
			res.Decisions = append(res.Decisions, d)
			continue
		}
		if err := applyExportDecision(res, d, policy.DryRun); err != nil {
			return res, err
		}
	}
	return res, nil
}

func applyExportDecision(res *ExportResult, d ExportDecision, dryRun bool) error {
	if info, ierr := os.Stat(d.AbsPath); ierr == nil {
		res.BytesFreed += info.Size()
	}
	if !dryRun {
		if err := os.Remove(d.AbsPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("retention: remove %s: %w", d.Rel, err)
		}
	}
	res.Removed++
	res.Decisions = append(res.Decisions, d)
	return nil
}

func exportLikeExt(name string) bool {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return true
	}
	switch filepath.Ext(lower) {
	case ".zip", ".pdf", ".md", ".json":
		return true
	}
	return false
}

// exportFile is a flat record returned by listExportFiles.
type exportFile struct {
	rel     string
	abs     string
	size    int64
	modTime time.Time
}

// listExportFiles walks baseDir and returns every regular file whose
// extension looks like an exported bundle or artifact. Missing baseDir is
// treated as an empty directory (no error).
func listExportFiles(baseDir string) ([]exportFile, error) {
	var items []exportFile
	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			if os.IsNotExist(werr) {
				return nil
			}
			return werr
		}
		if d.IsDir() || !exportLikeExt(d.Name()) {
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil {
			return nil //nolint:nilerr // skip unreadable file
		}
		rel, rerr := filepath.Rel(baseDir, path)
		if rerr != nil {
			rel = path
		}
		items = append(items, exportFile{
			rel:     filepath.ToSlash(rel),
			abs:     path,
			size:    info.Size(),
			modTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("retention: walk exports: %w", err)
	}
	return items, nil
}

func applyExportMaxCount(decisions []ExportDecision, maxCount int) {
	kept := make([]int, 0, len(decisions))
	for i, d := range decisions {
		if !d.Keep {
			continue
		}
		kept = append(kept, i)
	}
	if len(kept) <= maxCount {
		return
	}
	sort.SliceStable(kept, func(a, b int) bool {
		return decisions[kept[a]].ModTime.After(decisions[kept[b]].ModTime)
	})
	for _, idx := range kept[maxCount:] {
		decisions[idx].Keep = false
		decisions[idx].Reason = fmt.Sprintf("over_count>%d", maxCount)
	}
}
