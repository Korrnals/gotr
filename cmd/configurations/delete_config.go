package configurations

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
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
			configID, err := flags.ValidateRequiredID(args, 0, "config_id")
			if err != nil {
				return err
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations delete-config")
				dr.PrintSimple("Удалить конфигурацию", fmt.Sprintf("Config ID: %d", configID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeleteConfig(ctx, configID); err != nil {
				return fmt.Errorf("failed to delete configuration: %w", err)
			}

			fmt.Printf("✅ Configuration %d deleted\n", configID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
