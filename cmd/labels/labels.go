// Package labels implements CLI commands for TestRail Labels API
package labels

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all labels commands with the root command
func Register(root *cobra.Command, getClient GetClientFunc) {
	labelsCmd := &cobra.Command{
		Use:   "labels",
		Short: "Manage test labels",
		Long:  `Update labels for tests and test runs.`,
	}

	// Create 'update' parent command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update labels for tests",
		Long:  `Update labels for a single test or multiple tests in a run.`,
	}

	// Persistent flags for all update subcommands
	updateCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without making changes")

	// Add subcommands to 'update'
	updateCmd.AddCommand(newUpdateTestCmd(getClient))
	updateCmd.AddCommand(newUpdateTestsCmd(getClient))

	// Add 'update' to labels
	labelsCmd.AddCommand(updateCmd)

	root.AddCommand(labelsCmd)
}
