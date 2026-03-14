// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newUpdateLabelCmd создаёт команду 'labels update-label'
// Эндпоинт: POST /update_label/{label_id}
func newUpdateLabelCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-label <label_id>",
		Short: "Обновить метку",
		Long: `Обновляет существующую метку.

Требуются права на редактирование меток in project.
Максимальная длина названия метки — 20 символов.`,
		Example: `  # Обновить название метки
  gotr labels update-label 123 --project 1 --title "New Label Name"

  # Вывод в JSON
  gotr labels update-label 123 --project 1 --title "Bug" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			labelID, err := flags.ValidateRequiredID(args, 0, "label_id")
			if err != nil {
				return err
			}

			client := getClient(cmd)
			ctx := cmd.Context()

			projectID, _ := cmd.Flags().GetInt64("project")
			title, _ := cmd.Flags().GetString("title")

			req := data.UpdateLabelRequest{
				ProjectID: projectID,
				Title:     title,
			}

			resp, err := client.UpdateLabel(ctx, labelID, req)
			if err != nil {
				return fmt.Errorf("failed to update label: %w", err)
			}

			_, err = output.Output(cmd, resp, "labels", "json")
			return err
		},
	}

	cmd.Flags().Int64P("project", "p", 0, "ID проекта (обязательно)")
	cmd.Flags().StringP("title", "t", "", "Новое название метки (обязательно, max 20 символов)")
	output.AddFlag(cmd)

	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("title")

	return cmd
}
