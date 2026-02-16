// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"fmt"
	"strconv"
	"text/tabwriter"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'labels list'
// Эндпоинт: GET /get_labels/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
		Short: "Получить список меток проекта",
		Long:  `Получает список всех меток для указанного проекта.`,
		Example: `  # Список меток проекта
  gotr labels list 123

  # Вывод в JSON
  gotr labels list 123 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("invalid project_id: %s", args[0])
			}

			client := getClient(cmd)
			labels, err := client.GetLabels(projectID)
			if err != nil {
				return fmt.Errorf("failed to list labels: %w", err)
			}

			outputFlag, _ := cmd.Flags().GetString("save")
			if outputFlag != "" {
				_, err := save.Output(cmd, labels, "labels", "json")
				return err
			}

			if len(labels) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No labels found")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME")
			for _, l := range labels {
				fmt.Fprintf(w, "%d\t%s\n", l.ID, l.Name)
			}
			return w.Flush()
		},
	}
	save.AddFlag(cmd)
	return cmd
}

// Verify interface compliance
var _ interface {
	GetLabels(projectID int64) (data.GetLabelsResponse, error)
} = (*client.MockClient)(nil)
