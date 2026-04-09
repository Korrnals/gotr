// Package cases implements CLI commands for managing TestRail test cases.
package cases

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type used to obtain an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all test case management commands on the root command.
func Register(root *cobra.Command, getClient GetClientFunc) {
	casesCmd := &cobra.Command{
		Use:   "cases",
		Short: "Manage test cases",
		Long: `Manage test cases: create, read, update, delete and bulk operations.

Available operations:
  • add    — create a new test case
  • get    — retrieve a test case by ID
  • list   — list test cases with filters
  • update — update a test case
  • delete — delete a test case
  • bulk   — bulk operations (update/delete/copy/move)`,
	}

	// Register subcommands
	casesCmd.AddCommand(newAddCmd(getClient))
	casesCmd.AddCommand(newGetCmd(getClient))
	casesCmd.AddCommand(newListCmd(getClient))
	casesCmd.AddCommand(newUpdateCmd(getClient))
	casesCmd.AddCommand(newDeleteCmd(getClient))
	casesCmd.AddCommand(newBulkCmd(getClient))

	root.AddCommand(casesCmd)
}
