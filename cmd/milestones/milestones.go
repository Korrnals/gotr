// Package milestones implements CLI commands for managing TestRail milestones.
package milestones

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all milestone management commands on the given root.
func Register(root *cobra.Command, getClient GetClientFunc) {
	milestonesCmd := &cobra.Command{
		Use:   "milestones",
		Short: "Manage project milestones",
		Long: `Manage project milestones.

Milestones are used to group test runs by development stages
(e.g.: "Release 1.0", "Sprint 5", "Beta").

Available operations:
  • add    — create a new milestone
  • get    — get milestone information
  • list   — list all project milestones  
  • update — update a milestone
  • delete — delete a milestone`,
	}

	// Add subcommands
	milestonesCmd.AddCommand(newAddCmd(getClient))
	milestonesCmd.AddCommand(newGetCmd(getClient))
	milestonesCmd.AddCommand(newListCmd(getClient))
	milestonesCmd.AddCommand(newUpdateCmd(getClient))
	milestonesCmd.AddCommand(newDeleteCmd(getClient))

	root.AddCommand(milestonesCmd)
}
