package configurations

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newUpdateConfigCmd создаёт команду 'configurations update-config'
// Эндпоинт: POST /update_config/{config_id}
func newUpdateConfigCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-config [config_id]",
		Short: "Обновить конфигурацию",
		Long:  `Обновляет название существующей конфигурации.`,
		Example: `  # Изменить название конфигурации
  gotr configurations update-config 10 --name="Chrome 120"

  # Проверить перед обновлением
  gotr configurations update-config 10 --name="Chrome 120" --dry-run`,
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
					return fmt.Errorf("config_id is required in non-interactive mode: gotr configurations update-config [config_id]")
				}
				if _, ok := interactive.PrompterFromContext(ctx).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("config_id is required in non-interactive mode: gotr configurations update-config [config_id]")
				}

				configID, err = resolveConfigIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations update-config")
				dr.PrintSimple("Обновить конфигурацию", fmt.Sprintf("Config ID: %d, New Name: %s", configID, name))
				return nil
			}

			req := data.UpdateConfigRequest{Name: name}
			resp, err := cli.UpdateConfig(ctx, configID, &req)
			if err != nil {
				return fmt.Errorf("failed to update configuration: %w", err)
			}

			ui.Successf(os.Stdout, "Configuration %d updated", configID)
			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Новое название конфигурации (обязательно)")

	return cmd
}
