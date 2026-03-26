package configurations

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newDeleteGroupCmd создаёт команду 'configurations delete-group'
// Эндпоинт: POST /delete_config_group/{group_id}
func newDeleteGroupCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-group [group_id]",
		Short: "Удалить группу конфигураций",
		Long: `Удаляет группу конфигураций и все её конфигурации.

⚠️ Внимание: удаление нельзя отменить! Все конфигурации в группе
будут также удалены. Убедитесь, что группа не используется
в активных тест-планах.`,
		Example: `  # Удалить группу
  gotr configurations delete-group 5

  # Проверить перед удалением
  gotr configurations delete-group 5 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var groupID int64
			var err error
			if len(args) > 0 {
				groupID, err = flags.ValidateRequiredID(args, 0, "group_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations delete-group [group_id]")
				}
				if _, ok := interactive.PrompterFromContext(ctx).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations delete-group [group_id]")
				}

				groupID, err = resolveGroupIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations delete-group")
				dr.PrintSimple("Удалить группу", fmt.Sprintf("Group ID: %d", groupID))
				return nil
			}

			if err := cli.DeleteConfigGroup(ctx, groupID); err != nil {
				return fmt.Errorf("failed to delete group: %w", err)
			}

			ui.Successf(os.Stdout, "Group %d deleted", groupID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
