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

// newDeleteConfigCmd создаёт команду 'configurations delete-config'
// Эндпоинт: POST /delete_config/{config_id}
func newDeleteConfigCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-config [config_id]",
		Short: "Удалить конфигурацию",
		Long: `Удаляет конфигурацию из группы.

⚠️ Внимание: удаление нельзя отменить! Убедитесь, что конфигурация
не используется в активных тест-планах.`,
		Example: `  # Удалить конфигурацию
  gotr configurations delete-config 10

  # Проверить перед удалением
  gotr configurations delete-config 10 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var configID int64
			var err error
			if len(args) > 0 {
				configID, err = flags.ValidateRequiredID(args, 0, "config_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("config_id is required in non-interactive mode: gotr configurations delete-config [config_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("config_id is required in non-interactive mode: gotr configurations delete-config [config_id]")
				}

				configID, err = resolveConfigIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations delete-config")
				dr.PrintSimple("Удалить конфигурацию", fmt.Sprintf("Config ID: %d", configID))
				return nil
			}

			if err := cli.DeleteConfig(ctx, configID); err != nil {
				return fmt.Errorf("failed to delete configuration: %w", err)
			}

			ui.Successf(os.Stdout, "Configuration %d deleted", configID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
