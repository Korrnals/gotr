package report

import (
	"github.com/spf13/cobra"
)

// Register adds migration report commands to root.
func Register(root *cobra.Command) {
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Inspect local migration reports",
		Long:  "View migration reports saved in ~/.gotr/reports.",
	}

	reportCmd.AddCommand(newListCmd())
	reportCmd.AddCommand(newViewCmd())
	reportCmd.AddCommand(newShowCmd())
	root.AddCommand(reportCmd)
}
