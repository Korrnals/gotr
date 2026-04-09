// Package templates implements CLI commands for managing TestRail case templates.
package templates

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register adds all template-related subcommands to the root command.
func Register(root *cobra.Command, getClient GetClientFunc) {
	templatesCmd := &cobra.Command{
		Use:   "templates",
		Short: "Manage test case templates",
		Long: `Manage templates — display formats for test cases.

Templates define the structure and fields available when creating
and editing test cases in a project.

Available operations:
  • list   — list project templates`,
	}

	// Register subcommands
	templatesCmd.AddCommand(newListCmd(getClient))

	root.AddCommand(templatesCmd)
}
