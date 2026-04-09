// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package reports

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// newListCrossProjectCmd creates the 'reports list-cross-project' command.
// Endpoint: GET /get_cross_project_reports
func newListCrossProjectCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-cross-project",
		Short: "List cross-project reports",
		Long:  `Lists all available cross-project report templates.`,
		Example: `  # List cross-project reports
  gotr reports list-cross-project

  # Output as JSON
  gotr reports list-cross-project -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getClient(cmd)
			ctx := cmd.Context()
			reports, err := client.GetCrossProjectReports(ctx)
			if err != nil {
				return fmt.Errorf("failed to list cross-project reports: %w", err)
			}

			saveFlag, _ := cmd.Flags().GetBool("save")
			if saveFlag {
				_, err := output.Output(cmd, reports, "reports", "json")
				return err
			}

			if len(reports) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No cross-project reports found")
				return nil
			}

			t := ui.NewTable(cmd)
			t.AppendHeader(table.Row{"ID", "NAME", "DESCRIPTION"})
			for _, r := range reports {
				t.AppendRow(table.Row{r.ID, r.Name, r.Description})
			}
			ui.Table(cmd, t)
			return nil
		},
	}
	output.AddFlag(cmd)
	return cmd
}
