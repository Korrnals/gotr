// Package plans implements CLI commands for managing TestRail test plans.
package plans

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all test plan management commands on the given root.
func Register(root *cobra.Command, getClient GetClientFunc) {
	plansCmd := &cobra.Command{
		Use:   "plans",
		Short: "Manage test plans",
		Long: `Manage test plans: create, update, close, delete, and manage entries.

A test plan is a collection of test runs (entries) grouped by a common goal.

Main operations:
  • add    — create a test plan
  • get    — get plan information
  • list   — list project plans
  • update — update a plan
  • close  — close a plan (complete)
  • delete — delete a plan
  • entry  — manage plan entries (add/update/delete)`,
	}

	// Add subcommands
	plansCmd.AddCommand(newAddCmd(getClient))
	plansCmd.AddCommand(newGetCmd(getClient))
	plansCmd.AddCommand(newListCmd(getClient))
	plansCmd.AddCommand(newUpdateCmd(getClient))
	plansCmd.AddCommand(newCloseCmd(getClient))
	plansCmd.AddCommand(newDeleteCmd(getClient))
	plansCmd.AddCommand(newEntryCmd(getClient))

	root.AddCommand(plansCmd)
}
