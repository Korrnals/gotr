package snap

import (
	"fmt"
	"os"
	"strings"
	"time"

	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//nolint:gocyclo // GC command flow keeps retention rules and preview/apply paths together.
func newGCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gc",
		Short: "Run snapshot retention cleanup (preview by default)",
		Long: `Finds snapshots older than configured TTL and cleans them up.

By default, this command runs in preview mode and shows what would be deleted.
Use --confirm to actually delete snapshots.

Retention settings are read from config:
  • snap.retention.default_ttl_days
  • snap.retention.protected_prefixes
  • snap.retention.frozen_snapshots

Protected/frozen snapshots are never deleted by this command.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snaplib.NewStore()
			if err != nil {
				return fmt.Errorf("newGCCmd.func: %w", err)
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("newGCCmd.func: %w", err)
			}

			cfg := snaplib.ReadConfig()
			ttlDays, _ := cmd.Flags().GetInt("ttl-days")
			if ttlDays <= 0 {
				ttlDays = cfg.RetentionDays
			}
			if ttlDays <= 0 {
				return fmt.Errorf("gc failed: ttl-days must be positive")
			}

			protectedPrefixes := protectionPrefixes()
			frozenSet := make(map[string]struct{})
			for _, id := range viper.GetStringSlice("snap.retention.frozen_snapshots") {
				frozenSet[id] = struct{}{}
			}

			// Per-category TTL overrides. Reads `snap.retention.category_ttl_days`
			// as a map[string]int and applies built-in defaults for categories
			// that have a recommended retention shorter than the global TTL.
			categoryTTL := loadCategoryTTLs()

			now := time.Now().UTC()
			confirmFlag, _ := cmd.Flags().GetBool("confirm")
			ttlDaysOverride := cmd.Flags().Changed("ttl-days")
			entries := manifest.All()

			candidates := make([]snaplib.ManifestEntry, 0)
			for _, e := range entries {
				if _, frozen := frozenSet[e.ID]; frozen {
					continue
				}
				if hasAnyPrefix(e.Label, protectedPrefixes) {
					continue
				}
				effTTL := ttlDays
				if !ttlDaysOverride {
					if cat, ok := categoryTTL[string(e.Category)]; ok && cat > 0 {
						effTTL = cat
					}
				}
				cutoff := now.Add(-time.Duration(effTTL) * 24 * time.Hour)
				if e.Timestamp.Before(cutoff) {
					candidates = append(candidates, e)
				}
			}

			if len(candidates) == 0 {
				fmt.Fprintf(os.Stdout, "No snapshots eligible for retention cleanup (default ttl=%d days).\n", ttlDays)
				return nil
			}

			confirm := confirmFlag
			verb := "Would delete"
			if confirm {
				verb = "Deleting"
			}
			fmt.Fprintf(os.Stdout, "%s %d snapshot(s) older than %d days:\n", verb, len(candidates), ttlDays)
			for _, e := range candidates {
				fmt.Fprintf(os.Stdout, "  • %s (%s)\n", e.ID, e.Timestamp.Format(time.RFC3339))
			}

			if !confirm {
				fmt.Fprintln(os.Stdout, "Run with --confirm to apply deletion.")
				return nil
			}

			deleted := 0
			for _, e := range candidates {
				if err := store.Delete(e.ID); err != nil {
					return fmt.Errorf("gc delete snapshot %s: %w", e.ID, err)
				}
				if err := manifest.Remove(e.ID); err != nil {
					return fmt.Errorf("gc remove manifest %s: %w", e.ID, err)
				}
				deleted++
			}

			ui.Successf(os.Stdout, "Deleted %d snapshot(s)", deleted)
			return nil
		},
	}

	cmd.Flags().Bool("confirm", false, "Actually delete snapshots (default is preview only)")
	cmd.Flags().Int("ttl-days", 0, "Override retention TTL in days (default: config value)")
	return cmd
}

func hasAnyPrefix(label string, prefixes []string) bool {
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(label, p) {
			return true
		}
	}
	return false
}

// builtinCategoryTTLs encodes recommended per-category snapshot retentions.
// Categories not listed inherit the global default TTL.
//
// `cleanup-attachments` snapshots are produced by `gotr attachments cleanup`
// and are conceptually short-lived rollback safety nets — keep them around
// for a week by default to bound disk usage.
var builtinCategoryTTLs = map[string]int{
	"cleanup-attachments": 7,
}

// loadCategoryTTLs merges the built-in defaults with any user override
// from `snap.retention.category_ttl_days` (map of category -> int days).
// User-supplied entries take precedence; unknown categories are ignored.
func loadCategoryTTLs() map[string]int {
	out := make(map[string]int, len(builtinCategoryTTLs))
	for k, v := range builtinCategoryTTLs {
		out[k] = v
	}
	if !viper.IsSet("snap.retention.category_ttl_days") {
		return out
	}
	if raw := viper.GetStringMap("snap.retention.category_ttl_days"); raw != nil {
		for k, v := range raw {
			switch n := v.(type) {
			case int:
				if n > 0 {
					out[k] = n
				}
			case int64:
				if n > 0 {
					out[k] = int(n)
				}
			case float64:
				if n > 0 {
					out[k] = int(n)
				}
			}
		}
	}
	return out
}
