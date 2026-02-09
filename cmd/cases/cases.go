// Package cases implements CLI commands for TestRail Cases API
package cases

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all cases commands with the root command
func Register(root *cobra.Command, getClient GetClientFunc) {
	casesCmd := &cobra.Command{
		Use:   "cases",
		Short: "Manage test cases",
		Long:  `Create, read, update, delete, and bulk operations for test cases.`,
	}

	// Add subcommands
	casesCmd.AddCommand(newAddCmd(getClient))
	casesCmd.AddCommand(newGetCmd(getClient))
	casesCmd.AddCommand(newListCmd(getClient))
	casesCmd.AddCommand(newUpdateCmd(getClient))
	casesCmd.AddCommand(newDeleteCmd(getClient))
	casesCmd.AddCommand(newBulkCmd(getClient))

	root.AddCommand(casesCmd)
}
