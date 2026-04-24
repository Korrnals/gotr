package retention

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	intreport "github.com/Korrnals/gotr/internal/report"
)

// Decision annotates a report entry with the retention verdict.
type Decision struct {
	Rel      string
	AbsPath  string
	Category string
	ModTime  time.Time
	Keep     bool
	Reason   string // why it was kept or removed ("age>Nd", "over_count", "keep_category", "under_limits")
}

// Result summarizes a CleanupReports invocation.
type Result struct {
	Decisions []Decision
	Removed   int
	Kept      int
	DryRun    bool
	BytesFreed int64
}

// Summary returns a human-friendly multi-line report.
func (r *Result) Summary() string {
	if r == nil {
		return ""
	}
	var sb strings.Builder
	verb := "Removed"
	if r.DryRun {
		verb = "Would remove"
	}
	fmt.Fprintf(&sb, "%s %d report(s), kept %d (freed %d bytes).\n", verb, r.Removed, r.Kept, r.BytesFreed)
	for _, d := range r.Decisions {
		if d.Keep {
			continue
		}
		fmt.Fprintf(&sb, "  - %s  [%s]\n", d.Rel, d.Reason)
	}
	return sb.String()
}

// CleanupReports evaluates the policy against baseDir and deletes reports
// that fall outside the retention window (unless policy.DryRun is set). When
// policy.Enabled is false the function returns an empty Result immediately.
//
// Reports whose top-level category matches policy.KeepCategories are
// preserved regardless of age/count limits. INDEX.md is never considered.
func CleanupReports(baseDir string, policy Policy, now time.Time) (*Result, error) {
	res := &Result{DryRun: policy.DryRun}
	if !policy.Enabled {
		return res, nil
	}
	entries, err := intreport.RecursiveListReports(baseDir)
	if err != nil {
		return nil, fmt.Errorf("retention: list reports: %w", err)
	}
	keep := policy.KeepCategorySet()
	cutoff := policy.Cutoff(now)

	// Build decisions (newest first; RecursiveListReports sorts by ModTime desc).
	decisions := make([]Decision, 0, len(entries))
	for _, e := range entries {
		cls := intreport.ClassifyReport(e.Basename)
		cat := string(cls.Category)
		d := Decision{
			Rel:      e.Rel,
			AbsPath:  e.AbsPath,
			Category: cat,
			ModTime:  e.ModTime,
			Keep:     true,
			Reason:   "under_limits",
		}
		if _, ok := keep[cat]; ok {
			d.Reason = "keep_category"
			decisions = append(decisions, d)
			continue
		}
		if !cutoff.IsZero() && e.ModTime.Before(cutoff) {
			d.Keep = false
			d.Reason = fmt.Sprintf("age>%dd", policy.MaxAgeDays)
		}
		decisions = append(decisions, d)
	}

	if policy.MaxCount > 0 {
		// Rank surviving (non-keep_category) entries by mtime desc; drop the tail.
		applyMaxCount(decisions, policy.MaxCount)
	}

	for _, d := range decisions {
		if d.Keep {
			res.Kept++
			res.Decisions = append(res.Decisions, d)
			continue
		}
		if err := applyDecision(res, d, policy.DryRun); err != nil {
			return res, err
		}
	}

	if res.Removed > 0 && !policy.DryRun {
		if err := intreport.Reindex(baseDir); err != nil {
			return res, fmt.Errorf("retention: reindex: %w", err)
		}
	}
	return res, nil
}

// applyDecision executes (or simulates) a single deletion decision, updating
// res counters. Extracted to keep CleanupReports under gocyclo threshold.
func applyDecision(res *Result, d Decision, dryRun bool) error {
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

// applyMaxCount keeps the N most recent reports (among entries not already
// flagged for deletion and not marked keep_category) and flags the remainder
// for removal with reason "over_count".
func applyMaxCount(decisions []Decision, maxCount int) {
	// Collect indices of entries currently-kept and not in keep_category.
	kept := make([]int, 0, len(decisions))
	for i, d := range decisions {
		if !d.Keep {
			continue
		}
		if d.Reason == "keep_category" {
			continue
		}
		kept = append(kept, i)
	}
	if len(kept) <= maxCount {
		return
	}
	// Sort indices by mtime desc.
	sort.SliceStable(kept, func(a, b int) bool {
		return decisions[kept[a]].ModTime.After(decisions[kept[b]].ModTime)
	})
	for _, idx := range kept[maxCount:] {
		decisions[idx].Keep = false
		decisions[idx].Reason = fmt.Sprintf("over_count>%d", maxCount)
	}
}
