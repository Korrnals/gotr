package configurations

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newDeleteConfigCmd создаёт команду 'configurations delete-config'
// Эндпоинт: POST /delete_config/{config_id}
func newDeleteConfigCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-config <config_id>",
		Short: "Удалить конфигурацию",
		Long: `Удаляет конфигурацию из группы.

⚠️ Внимание: удаление нельзя отменить! Убедитесь, что конфигурация
не используется в активных тест-планах.`,
		Example: `  # Удалить конфигурацию
  gotr configurations delete-config 10

  # Проверить перед удалением
  gotr configurations delete-config 10 --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || configID <= 0 {
				return fmt.Errorf("некорректный config_id: %s", args[0])
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations delete-config")
				dr.PrintSimple("Удалить конфигурацию", fmt.Sprintf("Config ID: %d", configID))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeleteConfig(configID); err != nil {
				return fmt.Errorf("не удалось удалить конфигурацию: %w", err)
			}

			fmt.Printf("✅ Конфигурация %d удалена\n", configID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
