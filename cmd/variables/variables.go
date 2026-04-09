// Package variables implements CLI commands for managing TestRail variables.
package variables

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is a function type that returns a client instance.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all variable-related subcommands on the given root command.
func Register(root *cobra.Command, getClient GetClientFunc) {
	variablesCmd := &cobra.Command{
		Use:   "variables",
		Short: "Manage test case variables",
		Long: `Manage variables — configuration values for test cases.

Variables allow creating flexible test cases that can
adapt to different conditions without creating duplicates.
Variable values can be modified at the dataset level.

Available operations:
  • list   — list dataset variables
  • add    — create a variable
  • update — update a variable
  • delete — delete a variable`,
	}

	// Register subcommands
	variablesCmd.AddCommand(newListCmd(getClient))
	variablesCmd.AddCommand(newAddCmd(getClient))
	variablesCmd.AddCommand(newUpdateCmd(getClient))
	variablesCmd.AddCommand(newDeleteCmd(getClient))

	root.AddCommand(variablesCmd)
}
