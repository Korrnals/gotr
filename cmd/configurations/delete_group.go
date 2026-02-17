package configurations

import (
	"fmt"
	"strconv"

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
			groupID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || groupID <= 0 {
				return fmt.Errorf("некорректный group_id: %s", args[0])
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations delete-group")
				dr.PrintSimple("Удалить группу", fmt.Sprintf("Group ID: %d", groupID))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeleteConfigGroup(groupID); err != nil {
				return fmt.Errorf("не удалось удалить группу: %w", err)
			}

			fmt.Printf("✅ Группа %d удалена\n", groupID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
