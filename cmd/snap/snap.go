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
  • rollback — reverse a mutation using saved data
  • delete   — remove a snapshot
  • gc       — clean up orphaned snapshots`,
	}

	snapCmd.AddCommand(newListCmd())
	snapCmd.AddCommand(newInfoCmd())
	snapCmd.AddCommand(newRollbackCmd(getClient))
	snapCmd.AddCommand(newDeleteCmd())
	snapCmd.AddCommand(newGCCmd())

	root.AddCommand(snapCmd)
}
