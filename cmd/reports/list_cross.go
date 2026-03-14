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

// newListCrossProjectCmd создаёт команду 'reports list-cross-project'
// Эндпоинт: GET /get_cross_project_reports
func newListCrossProjectCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-cross-project",
		Short: "Получить список кросс-проектных отчётов",
		Long:  `Получает список всех доступных кросс-проектных шаблонов отчётов.`,
		Example: `  # Список кросс-проектных отчётов
  gotr reports list-cross-project

  # Вывод в JSON
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
