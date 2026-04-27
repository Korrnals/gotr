// Package snap implements CLI commands for snapshot management and rollback.
package snap

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type used to obtain an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers snapshot management commands on the root command.
func Register(root *cobra.Command, getClient GetClientFunc) {
	snapCmd := &cobra.Command{
		Use:   "snap",
		Short: "Manage snapshots and rollback",
		Long: `Manage pre-mutation snapshots: list, inspect, rollback, and clean up.

Snapshots are automatically created before mutating operations (update, delete, etc.)
when snap.enabled is true in config or --snapshot flag is set.

Available operations:
  • list     — list all snapshots
  • info     — show snapshot details
	• pin      — protect snapshot from retention cleanup
	• unpin    — remove protection label prefix
  • rollback — reverse a mutation using saved data
    • rollback list — browse rolled-back snapshots
    • rollback undo — undo a previous rollback
  • export   — export snapshot to a portable JSON file
  • delete   — remove a snapshot
  • gc       — clean up orphaned snapshots
  • manifest repair — reconcile manifest.json with on-disk snapshot dirs`,
	}

	snapCmd.AddCommand(newListCmd())
	snapCmd.AddCommand(newInfoCmd())
	snapCmd.AddCommand(newRollbackCmd(getClient))
	snapCmd.AddCommand(newExportCmd())
	snapCmd.AddCommand(newDeleteCmd())
	snapCmd.AddCommand(newPinCmd())
	snapCmd.AddCommand(newUnpinCmd())
	snapCmd.AddCommand(newGCCmd())
	snapCmd.AddCommand(newManifestCmd())

	root.AddCommand(snapCmd)
}
