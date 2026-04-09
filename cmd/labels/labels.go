// Package labels implements CLI commands for managing TestRail test labels.
package labels

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all label management commands on the given root.
func Register(root *cobra.Command, getClient GetClientFunc) {
	labelsCmd := &cobra.Command{
		Use:   "labels",
		Short: "Manage test labels",
		Long: `Manage labels for tests and test runs.

Labels allow you to categorize and group tests for convenient analysis.
You can update labels for a single test or for all tests in a run.`,
	}

	// Add get and management subcommands
	labelsCmd.AddCommand(newGetCmd(getClient))
	labelsCmd.AddCommand(newListCmd(getClient))
	labelsCmd.AddCommand(newUpdateLabelCmd(getClient))

	// Create the parent 'update' command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update labels for tests",
		Long: `Updates labels for a single test or for all tests in a run.

Available subcommands:
  • test  — update labels for a single test by ID
  • tests — update labels for all tests in a run`,
	}

	// Shared flags for all update subcommands
	updateCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without making changes")

	// Add subcommands to 'update'
	updateCmd.AddCommand(newUpdateTestCmd(getClient))
	updateCmd.AddCommand(newUpdateTestsCmd(getClient))

	// Attach 'update' to the labels command
	labelsCmd.AddCommand(updateCmd)

	root.AddCommand(labelsCmd)
}
