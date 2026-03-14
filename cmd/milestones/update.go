package milestones

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду 'milestones update'
// Эндпоинт: POST /update_milestone/{milestone_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <milestone_id>",
		Short: "Обновить существующий майлстон",
		Long: `Обновляет данные существующего майлстона.

Можно изменить название, описание, дедлайн и статус завершения.
Все флаги опциональны — изменятся только указанные поля.`,
		Example: `  # Изменить название майлстона
  gotr milestones update 12345 --name="Релиз 1.1"

  # Изменить дедлайн
  gotr milestones update 12345 --due-on="2026-04-01"

  # Отметить как завершённый
  gotr milestones update 12345 --is-completed=true

  # Изменить несколько полей
  gotr milestones update 12345 --name="Новое название" --description="Новое описание"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			milestoneID, err := flags.ValidateRequiredID(args, 0, "milestone_id")
			if err != nil {
				return err
			}

			req := data.UpdateMilestoneRequest{}

			if v, _ := cmd.Flags().GetString("name"); v != "" {
				req.Name = v
			}
			if v, _ := cmd.Flags().GetString("description"); v != "" {
				req.Description = v
			}
			if v, _ := cmd.Flags().GetString("due-on"); v != "" {
				req.DueOn = v
			}
			if v, _ := cmd.Flags().GetBool("is-completed"); v {
				req.IsCompleted = true
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("milestones update")
				dr.PrintSimple("Update Milestone", fmt.Sprintf("Milestone ID: %d", milestoneID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.UpdateMilestone(ctx, milestoneID, &req)
			if err != nil {
				return fmt.Errorf("failed to update milestone: %w", err)
			}

			ui.Successf(os.Stdout, "Milestone %d updated", milestoneID)
			return output.OutputResult(cmd, resp, "milestones")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без реального выполнения")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Новое название майлстона")
	cmd.Flags().String("description", "", "Новое описание")
	cmd.Flags().String("due-on", "", "Новый дедлайн (YYYY-MM-DD)")
	cmd.Flags().Bool("is-completed", false, "Отметить как завершённый")

	return cmd
}
