package snap

import (
	"context"
	"fmt"
	"os"

	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

func newRollbackCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback <snapshot_id>",
		Short: "Rollback a mutation using a snapshot",
		Long: `Reverses a mutation by applying the saved pre-mutation state.

For update operations: restores original field values (Tier 1 — full rollback).
For delete operations: re-creates the entity with a new ID (Tier 2).
For add operations: deletes the created entity.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapID := args[0]
			cli := getClient(cmd)
			ctx := cmd.Context()

			store, err := snaplib.NewStore()
			if err != nil {
				return err
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return err
			}

			entry := manifest.Find(snapID)
			if entry == nil {
				return fmt.Errorf("snapshot %q not found in manifest", snapID)
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			result, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Rolling back",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*snaplib.RollbackResult, error) {
				return snaplib.Rollback(ctx, cli, store, manifest, snapID)
			})
			if err != nil {
				return fmt.Errorf("rollback failed: %w", err)
			}

			ui.Successf(os.Stdout, "Rollback complete: %s", result.Message)
			return nil
		},
	}
	return cmd
}
