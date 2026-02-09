package milestones

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду 'milestones update'
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
			milestoneID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || milestoneID <= 0 {
				return fmt.Errorf("invalid milestone_id: %s", args[0])
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
				dr := dryrun.New("milestones update")
				dr.PrintSimple("Update Milestone", fmt.Sprintf("Milestone ID: %d", milestoneID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdateMilestone(milestoneID, &req)
			if err != nil {
				return fmt.Errorf("failed to update milestone: %w", err)
			}

			fmt.Printf("✅ Milestone %d updated\n", milestoneID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без реального выполнения")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")
	cmd.Flags().String("name", "", "Новое название майлстона")
	cmd.Flags().String("description", "", "Новое описание")
	cmd.Flags().String("due-on", "", "Новый дедлайн (YYYY-MM-DD)")
	cmd.Flags().Bool("is-completed", false, "Отметить как завершённый")

	return cmd
}
