// Package attachments implements CLI commands for managing TestRail attachments.
package attachments

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining a client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all attachment-related commands.
func Register(root *cobra.Command, getClient GetClientFunc) {
	attachmentsCmd := &cobra.Command{
		Use:   "attachments",
		Short: "Manage file attachments",
		Long: `Manage file attachments for test cases, plans, results, and runs.

Supported resource types for attaching files:
  • case       — attachment to a test case
  • plan       — attachment to a test plan
  • plan-entry — attachment to a plan entry
  • result     — attachment to a test result
  • run        — attachment to a test run`,
	}

	// Create the parent 'add' command
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add attachment to a resource",
		Long: `Uploads a file and attaches it to the specified resource.

Supported resource types: test case, plan, plan entry,
test result, or test run.`,
	}

	// Shared flags for all 'add' subcommands
	addCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without uploading the file")
	output.AddFlag(addCmd)

	// Register subcommands under 'add'
	addCmd.AddCommand(newAddCaseCmd(getClient))
	addCmd.AddCommand(newAddPlanCmd(getClient))
	addCmd.AddCommand(newAddPlanEntryCmd(getClient))
	addCmd.AddCommand(newAddResultCmd(getClient))
	addCmd.AddCommand(newAddRunCmd(getClient))

	// Add 'add' to the attachments command
	attachmentsCmd.AddCommand(addCmd)

	root.AddCommand(attachmentsCmd)
}
