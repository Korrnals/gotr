// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newUpdateLabelCmd создаёт команду 'labels update-label'
// Эндпоинт: POST /update_label/{label_id}
func newUpdateLabelCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-label [label_id]",
		Short: "Обновить метку",
		Long: `Обновляет существующую метку.

Требуются права на редактирование меток in project.
Максимальная длина названия метки — 20 символов.`,
		Example: `  # Обновить название метки
  gotr labels update-label 123 --project 1 --title "New Label Name"

  # Вывод в JSON
  gotr labels update-label 123 --project 1 --title "Bug" -o json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var labelID int64
			var err error
			if len(args) > 0 {
				labelID, err = flags.ValidateRequiredID(args, 0, "label_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("label_id is required in non-interactive mode: gotr labels update-label [label_id]")
				}
				if _, ok := interactive.PrompterFromContext(cmd.Context()).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("label_id is required in non-interactive mode: gotr labels update-label [label_id]")
				}
				if labelID, err = resolveLabelIDInteractive(cmd.Context(), getClient(cmd)); err != nil {
					return err
				}
			}

			projectID, _ := cmd.Flags().GetInt64("project")
			title, _ := cmd.Flags().GetString("title")

			req := data.UpdateLabelRequest{
				ProjectID: projectID,
				Title:     title,
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("labels update-label")
				dr.PrintOperation(
					fmt.Sprintf("Update label %d", labelID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/update_label/%d", labelID),
					req,
				)
				return nil
			}

			client := getClient(cmd)
			ctx := cmd.Context()

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
	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без обновления")
	output.AddFlag(cmd)

	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("title")

	return cmd
}
