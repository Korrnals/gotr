// Package attachments implements CLI commands for TestRail Attachments API
package attachments

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) *client.HTTPClient

// Register registers all attachment commands with the root command
func Register(root *cobra.Command, getClient GetClientFunc) {
	attachmentsCmd := &cobra.Command{
		Use:   "attachments",
		Short: "Manage file attachments",
		Long:  `Add file attachments to cases, plans, plan entries, results, and runs.`,
	}

	// Create 'add' parent command
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add attachment to a resource",
		Long:  `Upload a file attachment to a specific resource (case, plan, plan-entry, result, or run).`,
	}

	// Persistent flags for all add subcommands
	addCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without making changes")
	addCmd.PersistentFlags().StringP("output", "o", "", "Save response to file")

	// Add subcommands to 'add'
	addCmd.AddCommand(newAddCaseCmd(getClient))
	addCmd.AddCommand(newAddPlanCmd(getClient))
	addCmd.AddCommand(newAddPlanEntryCmd(getClient))
	addCmd.AddCommand(newAddResultCmd(getClient))
	addCmd.AddCommand(newAddRunCmd(getClient))

	// Add 'add' to attachments
	attachmentsCmd.AddCommand(addCmd)

	root.AddCommand(attachmentsCmd)
}
