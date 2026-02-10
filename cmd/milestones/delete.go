package milestones

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/spf13/cobra"
)

// newDeleteCmd создаёт команду 'milestones delete'
// Эндпоинт: POST /delete_milestone/{milestone_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <milestone_id>",
		Short: "Удалить майлстон",
		Long: `Удаляет майлстон по его идентификатору.

⚠️ Внимание: удаление нельзя отменить!
Удалённый майлстон нельзя восстановить, придётся создавать заново.
Используйте --dry-run для проверки перед удалением.`,
		Example: `  # Удалить майлстон (с подтверждением опасности)
  gotr milestones delete 12345

  # Проверить что будет удалено (без реального удаления)
  gotr milestones delete 12345 --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			milestoneID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || milestoneID <= 0 {
				return fmt.Errorf("invalid milestone_id: %s", args[0])
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("milestones delete")
				dr.PrintSimple("Delete Milestone", fmt.Sprintf("Milestone ID: %d", milestoneID))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeleteMilestone(milestoneID); err != nil {
				return fmt.Errorf("failed to delete milestone: %w", err)
			}

			fmt.Printf("✅ Milestone %d deleted\n", milestoneID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
