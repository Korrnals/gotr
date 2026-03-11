// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

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
			ctx := cmd.Context()
			labels, err := client.GetLabels(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to list labels: %w", err)
			}

			saveFlag, _ := cmd.Flags().GetBool("save")
			if saveFlag {
				_, err := output.Output(cmd, labels, "labels", "json")
				return err
			}

			if len(labels) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No labels found")
				return nil
			}

			t := ui.NewTable(cmd)
			t.AppendHeader(table.Row{"ID", "NAME"})
			for _, l := range labels {
				t.AppendRow(table.Row{l.ID, l.Name})
			}
			ui.Table(cmd, t)
			return nil
		},
	}
	output.AddFlag(cmd)
	return cmd
}

// Verify interface compliance
var _ interface {
	GetLabels(ctx context.Context, projectID int64) (data.GetLabelsResponse, error)
} = (*client.MockClient)(nil)
