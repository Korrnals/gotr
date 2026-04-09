// Package datasets implements CLI commands for managing TestRail datasets.
package datasets

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is a function type for obtaining the API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all dataset management subcommands.
func Register(root *cobra.Command, getClient GetClientFunc) {
	datasetsCmd := &cobra.Command{
		Use:   "datasets",
		Short: "Manage datasets (test data)",
		Long: `Manage datasets — test data tables for parameterized testing.

Datasets allow running the same test case with different input data sets
without creating duplicate cases. Each dataset contains a name and a table
with columns (parameters) and rows (values).

Used when creating test plans with parameterized test runs.

Available operations:
  • list   — list project datasets
  • get    — get a dataset by ID
  • add    — create a new dataset
  • update — update a dataset
  • delete — delete a dataset`,
	}

	// Register subcommands
	datasetsCmd.AddCommand(newListCmd(getClient))
	datasetsCmd.AddCommand(newGetCmd(getClient))
	datasetsCmd.AddCommand(newAddCmd(getClient))
	datasetsCmd.AddCommand(newUpdateCmd(getClient))
	datasetsCmd.AddCommand(newDeleteCmd(getClient))

	root.AddCommand(datasetsCmd)
}
