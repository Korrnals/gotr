package milestones

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
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
			milestoneID, err := flags.ValidateRequiredID(args, 0, "milestone_id")
			if err != nil {
				return err
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("milestones delete")
				dr.PrintSimple("Delete Milestone", fmt.Sprintf("Milestone ID: %d", milestoneID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeleteMilestone(ctx, milestoneID); err != nil {
				return fmt.Errorf("failed to delete milestone: %w", err)
			}

			ui.Successf(os.Stdout, "Milestone %d deleted", milestoneID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
