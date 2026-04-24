package report

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// OrganizePlan describes a single move that MigrateFlatLayout intends to
// perform. AbsSrc is the current absolute path, AbsDest is the destination
// under the categorized hierarchy, and Rel{Src,Dest} are the paths relative
// to the reports root (used for human-readable output).
type OrganizePlan struct {
	AbsSrc   string
	AbsDest  string
	RelSrc   string
	RelDest  string
	Category Category
}

// OrganizeResult summarizes a MigrateFlatLayout invocation.
type OrganizeResult struct {
	Plans   []OrganizePlan
	Moved   int
	Skipped int // files that were already in the new hierarchy or a conflict was detected
	DryRun  bool
}

// MigrateFlatLayout relocates every top-level file under baseDir into the
// categorized hierarchy (<category>/<label|default>/<YYYY-MM>/). INDEX.md is
// never moved: it is re-rendered at the end. When dryRun is true nothing is
// written to disk and the returned plan describes what would happen.
//
// Files that already live in a subdirectory are left alone; only the flat
// top-level layout is migrated. The operation is idempotent.
func MigrateFlatLayout(baseDir string, dryRun bool) (*OrganizeResult, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("organize: read %s: %w", baseDir, err)
	}

	result := &OrganizeResult{DryRun: dryRun}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "INDEX.md" {
			continue
		}
		if !reportLikeExt(name) {
			continue
		}

		cls := ClassifyReport(name)
		relDest := filepath.Join(cls.RelDir(), name)
		absSrc := filepath.Join(baseDir, name)
		absDest := filepath.Join(baseDir, relDest)

		// Same path → nothing to do.
		if absSrc == absDest {
			result.Skipped++
			continue
		}

		result.Plans = append(result.Plans, OrganizePlan{
			AbsSrc:   absSrc,
			AbsDest:  absDest,
			RelSrc:   name,
			RelDest:  filepath.ToSlash(relDest),
			Category: cls.Category,
		})
	}

	// Deterministic ordering helps dry-run output and tests.
	sort.Slice(result.Plans, func(i, j int) bool {
		return result.Plans[i].RelSrc < result.Plans[j].RelSrc
	})

	if dryRun {
		return result, nil
	}

	for _, p := range result.Plans {
		if err := os.MkdirAll(filepath.Dir(p.AbsDest), 0o755); err != nil {
			return result, fmt.Errorf("organize: mkdir %s: %w", filepath.Dir(p.AbsDest), err)
		}
		// Refuse to overwrite an existing file at the destination: the caller
		// can re-run dry-run first to inspect conflicts.
		if _, err := os.Stat(p.AbsDest); err == nil {
			result.Skipped++
			continue
		}
		if err := os.Rename(p.AbsSrc, p.AbsDest); err != nil {
			// Cross-device? fall back to copy+remove.
			if cerr := copyFile(p.AbsSrc, p.AbsDest); cerr != nil {
				return result, fmt.Errorf("organize: move %s: rename err=%v, copy err=%w", p.RelSrc, err, cerr)
			}
			_ = os.Remove(p.AbsSrc)
		}
		result.Moved++
	}

	if result.Moved > 0 {
		if err := Reindex(baseDir); err != nil {
			return result, fmt.Errorf("organize: reindex: %w", err)
		}
	}
	return result, nil
}

// Summary renders a compact multi-line string describing the planned or
// executed moves. It is safe to call on a nil receiver (returns empty).
func (r *OrganizeResult) Summary() string {
	if r == nil {
		return ""
	}
	var sb strings.Builder
	if r.DryRun {
		fmt.Fprintf(&sb, "Dry run: would move %d file(s), skip %d.\n", len(r.Plans), r.Skipped)
	} else {
		fmt.Fprintf(&sb, "Organize: moved %d file(s), skipped %d.\n", r.Moved, r.Skipped)
	}
	for _, p := range r.Plans {
		fmt.Fprintf(&sb, "  %s → %s\n", p.RelSrc, p.RelDest)
	}
	return sb.String()
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src) //nolint:gosec // src is under baseDir controlled by caller
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}
