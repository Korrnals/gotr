package snap

import (
	"fmt"
	"os"
	"strings"

	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newPinCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pin [snapshot_id]",
		Short: "Protect a snapshot from retention cleanup",
		Long:  "Adds a protection prefix to snapshot label (for example: pinned_*) so retention GC can skip it.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snaplib.NewStore()
			if err != nil {
				return fmt.Errorf("newPinCmd.func: %w", err)
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("newPinCmd.func: %w", err)
			}

			snapID, err := resolveSnapshotID(cmd.Context(), args, manifest, "Select snapshot to pin:")
			if err != nil {
				return fmt.Errorf("newPinCmd.func: %w", err)
			}

			entry := manifest.Find(snapID)
			if entry == nil {
				return fmt.Errorf("snapshot %q not found in manifest", snapID)
			}

			prefix, _ := cmd.Flags().GetString("prefix")
			if prefix == "" {
				prefix = defaultProtectionPrefix()
			}

			if strings.HasPrefix(entry.Label, prefix) {
				ui.Successf(os.Stdout, "Snapshot %s already pinned (%s)", snapID, entry.Label)
				return nil
			}

			baseLabel := entry.Label
			if baseLabel == "" {
				baseLabel = "snapshot"
			}
			newLabel := prefix + baseLabel
			if err := snaplib.ValidateLabel(newLabel); err != nil {
				return fmt.Errorf("pin snapshot: invalid resulting label %q: %w", newLabel, err)
			}

			meta, err := store.LoadMeta(snapID)
			if err != nil {
				return fmt.Errorf("pin snapshot: load meta: %w", err)
			}
			meta.Label = newLabel
			if err := store.SaveMeta(meta); err != nil {
				return fmt.Errorf("pin snapshot: save meta: %w", err)
			}
			if err := manifest.UpdateLabel(snapID, newLabel); err != nil {
				return fmt.Errorf("pin snapshot: update manifest: %w", err)
			}

			ui.Successf(os.Stdout, "Snapshot %s pinned with label: %s", snapID, newLabel)
			return nil
		},
	}

	cmd.Flags().String("prefix", "", "Protection label prefix (default from config: snap.retention.protected_prefixes[0] or pinned_)")
	return cmd
}

func newUnpinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unpin [snapshot_id]",
		Short: "Remove protection prefix from snapshot label",
		Long:  "Removes configured protection prefixes from snapshot label, making it eligible for retention cleanup.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snaplib.NewStore()
			if err != nil {
				return fmt.Errorf("newUnpinCmd.func: %w", err)
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("newUnpinCmd.func: %w", err)
			}

			snapID, err := resolveSnapshotID(cmd.Context(), args, manifest, "Select snapshot to unpin:")
			if err != nil {
				return fmt.Errorf("newUnpinCmd.func: %w", err)
			}

			entry := manifest.Find(snapID)
			if entry == nil {
				return fmt.Errorf("snapshot %q not found in manifest", snapID)
			}

			newLabel, changed := stripProtectionPrefixes(entry.Label, protectionPrefixes())
			if !changed {
				ui.Successf(os.Stdout, "Snapshot %s is not pinned", snapID)
				return nil
			}

			if err := snaplib.ValidateLabel(newLabel); err != nil {
				return fmt.Errorf("unpin snapshot: invalid resulting label %q: %w", newLabel, err)
			}

			meta, err := store.LoadMeta(snapID)
			if err != nil {
				return fmt.Errorf("unpin snapshot: load meta: %w", err)
			}
			meta.Label = newLabel
			if err := store.SaveMeta(meta); err != nil {
				return fmt.Errorf("unpin snapshot: save meta: %w", err)
			}
			if err := manifest.UpdateLabel(snapID, newLabel); err != nil {
				return fmt.Errorf("unpin snapshot: update manifest: %w", err)
			}

			ui.Successf(os.Stdout, "Snapshot %s unpinned, label: %s", snapID, newLabel)
			return nil
		},
	}
}

func protectionPrefixes() []string {
	prefixes := viper.GetStringSlice("snap.retention.protected_prefixes")
	if len(prefixes) == 0 {
		return []string{"pinned_"}
	}
	return prefixes
}

func defaultProtectionPrefix() string {
	prefixes := protectionPrefixes()
	if len(prefixes) == 0 {
		return "pinned_"
	}
	return prefixes[0]
}

func stripProtectionPrefixes(label string, prefixes []string) (string, bool) {
	if label == "" {
		return label, false
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(label, prefix) {
			trimmed := strings.TrimPrefix(label, prefix)
			if trimmed == "" {
				trimmed = "snapshot"
			}
			return trimmed, true
		}
	}
	return label, false
}
