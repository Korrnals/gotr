// Package plans implements CLI commands for TestRail Plans API
package plans

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all plans commands with the root command
func Register(root *cobra.Command, getClient GetClientFunc) {
	plansCmd := &cobra.Command{
		Use:   "plans",
		Short: "Manage test plans",
		Long:  `Create, read, update, close, delete, and manage entries for test plans.`,
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
