// Package retention implements age- and count-based cleanup for gotr
// artifacts (reports, snapshots, exports) driven by the `retention:` section
// of the user config file.
//
// Policies are intentionally opt-in: a zero-value Policy with Enabled=false
// skips all work. The package exposes three cleanup entry points — one per
// artifact family — plus a viper-backed loader that honors the documented
// configuration schema (see guides/configuration.md).
package retention

import (
	"time"

	"github.com/spf13/viper"
)

// Policy captures the retention rules shared by reports, snaps and exports.
// Unset fields have no effect, so a Policy with only MaxAgeDays>0 will prune
// by age alone.
type Policy struct {
	Enabled        bool
	MaxAgeDays     int
	MaxCount       int
	KeepCategories []string // only consulted by report cleanup
	DryRun         bool
}

// Cutoff returns the "anything older than this timestamp is eligible for
// deletion" boundary, or the zero value when age-based pruning is disabled.
func (p Policy) Cutoff(now time.Time) time.Time {
	if p.MaxAgeDays <= 0 {
		return time.Time{}
	}
	return now.Add(-time.Duration(p.MaxAgeDays) * 24 * time.Hour)
}

// KeepCategorySet returns a lookup set for KeepCategories (lower-cased).
func (p Policy) KeepCategorySet() map[string]struct{} {
	if len(p.KeepCategories) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(p.KeepCategories))
	for _, c := range p.KeepCategories {
		out[c] = struct{}{}
	}
	return out
}

// LoadReports reads `retention.reports.*` from viper.
func LoadReports() Policy {
	return Policy{
		Enabled:        viper.GetBool("retention.reports.enabled"),
		MaxAgeDays:     viper.GetInt("retention.reports.max_age_days"),
		MaxCount:       viper.GetInt("retention.reports.max_count"),
		KeepCategories: viper.GetStringSlice("retention.reports.keep_categories"),
		DryRun:         viper.GetBool("retention.reports.dry_run"),
	}
}

// LoadExports reads `retention.exports.*` from viper.
func LoadExports() Policy {
	return Policy{
		Enabled:    viper.GetBool("retention.exports.enabled"),
		MaxAgeDays: viper.GetInt("retention.exports.max_age_days"),
		MaxCount:   viper.GetInt("retention.exports.max_count"),
		DryRun:     viper.GetBool("retention.exports.dry_run"),
	}
}

// LoadSnaps reads `retention.snaps.*` from viper. Snapshot cleanup is
// currently handled by `gotr snap gc`, so this loader exists for the unified
// `gotr cleanup` command to report on configuration without duplicating the
// store logic.
func LoadSnaps() Policy {
	return Policy{
		Enabled:    viper.GetBool("retention.snaps.enabled"),
		MaxAgeDays: viper.GetInt("retention.snaps.max_age_days"),
		MaxCount:   viper.GetInt("retention.snaps.max_count"),
		DryRun:     viper.GetBool("retention.snaps.dry_run"),
	}
}
