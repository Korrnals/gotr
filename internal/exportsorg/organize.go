// Package exportsorg migrates the pre-v3.3.0 flat ~/.gotr/exports/ layout
// into the new hierarchy:
//
//	~/.gotr/exports/snaps/    — portable snapshot bundles (*.tar.gz, *.tgz)
//	~/.gotr/exports/reports/  — exported report bundles (*.zip) and single-file reports
//	~/.gotr/exports/api/      — raw API responses (one subdirectory per resource)
//
// The flat layout still works for reads (both snap and report importers accept
// any path), but new writes land under the categorized subdirectories. This
// package provides a pure-Go planner + mover used by `gotr export organize`.
package exportsorg

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Korrnals/gotr/internal/paths"
)

// Category identifies the destination bucket for an entry in ~/.gotr/exports/.
type Category string

const (
	CategorySnaps   Category = "snaps"
	CategoryReports Category = "reports"
	CategoryAPI     Category = "api"
	CategoryUnknown Category = "unknown"
)

// Plan describes a single move the organizer proposes to make.
type Plan struct {
	AbsSrc   string
	AbsDest  string
	RelSrc   string
	RelDest  string
	Category Category
	IsDir    bool
}

// Result summarizes the outcome of a MigrateExportsLayout call.
type Result struct {
	Plans   []Plan
	Moved   int
	Skipped int
	DryRun  bool
}

// Classify inspects a direct child of ~/.gotr/exports/ and decides which
// subdirectory it belongs to. The already-categorized subdirectories
// (`snaps`, `reports`, `api`) return CategoryUnknown to signal "leave alone".
func Classify(name string, isDir bool) Category {
	switch name {
	case paths.ExportsSnapsSubdir, paths.ExportsReportsSubdir, paths.ExportsAPISubdir:
		return CategoryUnknown
	}
	if isDir {
		// Historically each `gotr export <resource> --save` created a
		// top-level resource directory (cases/, suites/, runs/, ...).
		// Those all move wholesale under api/.
		return CategoryAPI
	}
	low := strings.ToLower(name)
	switch {
	case strings.HasSuffix(low, ".tar.gz"), strings.HasSuffix(low, ".tgz"):
		return CategorySnaps
	case strings.HasSuffix(low, ".zip"):
		return CategoryReports
	case strings.HasSuffix(low, ".pdf"), strings.HasSuffix(low, ".md"):
		// Single-file report copies land under reports/.
		return CategoryReports
	case strings.HasSuffix(low, ".json"):
		// Could be either a report JSON or an API save; send to reports/
		// since single-file API saves always lived inside a resource dir
		// and thus wouldn't reach this branch.
		return CategoryReports
	}
	return CategoryUnknown
}

// MigrateExportsLayout scans baseDir (~/.gotr/exports/) and moves each
// uncategorized child into the appropriate subdirectory. Dry-run returns
// plans without touching the filesystem.
func MigrateExportsLayout(baseDir string, dryRun bool) (*Result, error) {
	if baseDir == "" {
		return nil, errors.New("exportsorg: empty baseDir")
	}
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &Result{DryRun: dryRun}, nil
		}
		return nil, fmt.Errorf("exportsorg: read %s: %w", baseDir, err)
	}

	plans := buildPlans(baseDir, entries)
	res := &Result{Plans: plans, DryRun: dryRun}

	if dryRun {
		return res, nil
	}

	for i := range plans {
		if err := applyPlan(&plans[i], res); err != nil {
			return res, err
		}
	}
	return res, nil
}

func buildPlans(baseDir string, entries []os.DirEntry) []Plan {
	var plans []Plan
	for _, e := range entries {
		name := e.Name()
		cat := Classify(name, e.IsDir())
		if cat == CategoryUnknown {
			continue
		}
		src := filepath.Join(baseDir, name)
		dest := filepath.Join(baseDir, string(cat), name)
		plans = append(plans, Plan{
			AbsSrc:   src,
			AbsDest:  dest,
			RelSrc:   name,
			RelDest:  filepath.Join(string(cat), name),
			Category: cat,
			IsDir:    e.IsDir(),
		})
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].RelSrc < plans[j].RelSrc })
	return plans
}

func applyPlan(p *Plan, res *Result) error {
	if _, err := os.Stat(p.AbsDest); err == nil {
		// Never overwrite existing categorized files.
		res.Skipped++
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("exportsorg: stat %s: %w", p.AbsDest, err)
	}
	if err := os.MkdirAll(filepath.Dir(p.AbsDest), 0o755); err != nil {
		return fmt.Errorf("exportsorg: mkdir %s: %w", filepath.Dir(p.AbsDest), err)
	}
	if err := os.Rename(p.AbsSrc, p.AbsDest); err != nil {
		// Fall back to copy+remove for cross-device moves.
		if p.IsDir {
			return fmt.Errorf("exportsorg: rename %s -> %s: %w", p.AbsSrc, p.AbsDest, err)
		}
		if cerr := copyFile(p.AbsSrc, p.AbsDest); cerr != nil {
			return fmt.Errorf("exportsorg: copy %s -> %s: %w", p.AbsSrc, p.AbsDest, cerr)
		}
		if rerr := os.Remove(p.AbsSrc); rerr != nil {
			return fmt.Errorf("exportsorg: remove %s: %w", p.AbsSrc, rerr)
		}
	}
	res.Moved++
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // paths controlled internally
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst) //nolint:gosec // destination under ~/.gotr/exports
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
