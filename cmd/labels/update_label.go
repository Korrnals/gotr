// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newUpdateLabelCmd создаёт команду 'labels update-label'
// Эндпоинт: POST /update_label/{label_id}
func newUpdateLabelCmd(getClient GetClientFunc) *cobra.Command {
	var (
		projectID int64
		title     string
	)

	cmd := &cobra.Command{
		Use:   "update-label <label_id>",
		Short: "Обновить метку",
		Long: `Обновляет существующую метку.

Требуются права на редактирование меток в проекте.
Максимальная длина названия метки — 20 символов.`,
		Example: `  # Обновить название метки
  gotr labels update-label 123 --project 1 --title "New Label Name"

  # Вывод в JSON
  gotr labels update-label 123 --project 1 --title "Bug" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			labelID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || labelID <= 0 {
				return fmt.Errorf("invalid label_id: %s", args[0])
			}

			client := getClient(cmd)
			req := data.UpdateLabelRequest{
				ProjectID: projectID,
				Title:     title,
			}

			resp, err := client.UpdateLabel(labelID, req)
			if err != nil {
				return fmt.Errorf("failed to update label: %w", err)
			}

			_, err = save.Output(cmd, resp, "labels", "json")
			return err
		},
	}

	cmd.Flags().Int64VarP(&projectID, "project", "p", 0, "ID проекта (обязательно)")
	cmd.Flags().StringVarP(&title, "title", "t", "", "Новое название метки (обязательно, max 20 символов)")
	save.AddFlag(cmd)

	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("title")

	return cmd
}
