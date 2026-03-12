package configurations

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newDeleteGroupCmd создаёт команду 'configurations delete-group'
// Эндпоинт: POST /delete_config_group/{group_id}
func newDeleteGroupCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-group <group_id>",
		Short: "Удалить группу конфигураций",
		Long: `Удаляет группу конфигураций и все её конфигурации.

⚠️ Внимание: удаление нельзя отменить! Все конфигурации в группе
будут также удалены. Убедитесь, что группа не используется
в активных тест-планах.`,
		Example: `  # Удалить группу
  gotr configurations delete-group 5

  # Проверить перед удалением
  gotr configurations delete-group 5 --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID, err := flags.ValidateRequiredID(args, 0, "group_id")
			if err != nil {
				return err
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations delete-group")
				dr.PrintSimple("Удалить группу", fmt.Sprintf("Group ID: %d", groupID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeleteConfigGroup(ctx, groupID); err != nil {
				return fmt.Errorf("failed to delete group: %w", err)
			}

			fmt.Printf("✅ Group %d deleted\n", groupID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
