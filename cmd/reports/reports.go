// Package reports implements CLI commands for managing TestRail reports.
package reports

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register adds all report-related subcommands to the root command.
func Register(root *cobra.Command, getClient GetClientFunc) {
	reportsCmd := &cobra.Command{
		Use:   "reports",
		Short: "Manage project reports",
		Long: `Manage TestRail report templates and generate reports.

Report templates are used to create various types of testing reports
(summary reports, coverage reports, comparison reports).

Available operations:
  • list               — list project report templates
  • list-cross-project — list cross-project reports
  • run                — run report generation from a template
  • run-cross-project  — run a cross-project report`,
	}

	// Register subcommands
	reportsCmd.AddCommand(newListCmd(getClient))
	reportsCmd.AddCommand(newListCrossProjectCmd(getClient))
	reportsCmd.AddCommand(newRunCmd(getClient))
	reportsCmd.AddCommand(newRunCrossProjectCmd(getClient))

	root.AddCommand(reportsCmd)
}
