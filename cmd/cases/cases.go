// Package cases implements CLI commands for TestRail Cases bulk operations
package cases

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) *client.HTTPClient

// Register registers all cases commands with the root command
func Register(root *cobra.Command, getClient GetClientFunc) {
	casesCmd := &cobra.Command{
		Use:   "cases",
		Short: "Manage test cases (bulk operations)",
		Long:  `Bulk operations for test cases: update, delete, copy, move.`,
	}

	// Persistent flags for all subcommands
	casesCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without making changes")
	casesCmd.PersistentFlags().StringP("output", "o", "", "Save response to file")

	// Add subcommands
	casesCmd.AddCommand(newUpdateCmd(getClient))
	casesCmd.AddCommand(newDeleteCmd(getClient))
	casesCmd.AddCommand(newCopyCmd(getClient))
	casesCmd.AddCommand(newMoveCmd(getClient))

	root.AddCommand(casesCmd)
}
