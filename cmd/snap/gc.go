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

			cutoff := time.Now().UTC().Add(-time.Duration(ttlDays) * 24 * time.Hour)
			entries := manifest.All()

			candidates := make([]snaplib.ManifestEntry, 0)
			for _, e := range entries {
				if _, frozen := frozenSet[e.ID]; frozen {
					continue
				}
				if hasAnyPrefix(e.Label, protectedPrefixes) {
					continue
				}
				if e.Timestamp.Before(cutoff) {
					candidates = append(candidates, e)
				}
			}

			if len(candidates) == 0 {
				fmt.Fprintf(os.Stdout, "No snapshots eligible for retention cleanup (ttl=%d days).\n", ttlDays)
				return nil
			}

			confirm, _ := cmd.Flags().GetBool("confirm")
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
