// Package configurations implements CLI commands for managing TestRail configurations.
package configurations

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is a function type for obtaining the API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all configuration management subcommands.
func Register(root *cobra.Command, getClient GetClientFunc) {
	configsCmd := &cobra.Command{
		Use:   "configurations",
		Short: "Manage test environment configurations",
		Long: `Manage configurations — test environments for test runs.

Configurations represent different testing environments:
  • Browsers (Chrome, Firefox, Safari)
  • Operating systems (Windows, macOS, Linux)
  • Mobile devices (iOS, Android)
  • Software versions and other parameters

Configurations are organized into groups (e.g., "Browsers", "OS").
Each group contains individual configurations (e.g., "Chrome", "Firefox").

Available operations:
  • list          — list project configurations
  • add-group     — create a configuration group
  • add-config    — add a configuration to a group
  • update-group  — update a group
  • update-config — update a configuration
  • delete-group  — delete a group
  • delete-config — delete a configuration`,
	}

	// Register subcommands
	configsCmd.AddCommand(newListCmd(getClient))
	configsCmd.AddCommand(newAddGroupCmd(getClient))
	configsCmd.AddCommand(newAddConfigCmd(getClient))
	configsCmd.AddCommand(newUpdateGroupCmd(getClient))
	configsCmd.AddCommand(newUpdateConfigCmd(getClient))
	configsCmd.AddCommand(newDeleteGroupCmd(getClient))
	configsCmd.AddCommand(newDeleteConfigCmd(getClient))

	root.AddCommand(configsCmd)
}
