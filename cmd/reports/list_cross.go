// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package reports

import (
	"fmt"
	"text/tabwriter"

	"github.com/Korrnals/gotr/internal/output"
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
			reports, err := client.GetCrossProjectReports()
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

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION")
			for _, r := range reports {
				fmt.Fprintf(w, "%d\t%s\t%s\n", r.ID, r.Name, r.Description)
			}
			return w.Flush()
		},
	}
	output.AddFlag(cmd)
	return cmd
}
