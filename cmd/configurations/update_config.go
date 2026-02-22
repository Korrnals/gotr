package configurations

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newUpdateConfigCmd создаёт команду 'configurations update-config'
// Эндпоинт: POST /update_config/{config_id}
func newUpdateConfigCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-config <config_id>",
		Short: "Обновить конфигурацию",
		Long: `Обновляет название существующей конфигурации.`,
		Example: `  # Изменить название конфигурации
  gotr configurations update-config 10 --name="Chrome 120"

  # Проверить перед обновлением
  gotr configurations update-config 10 --name="Chrome 120" --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || configID <= 0 {
				return fmt.Errorf("некорректный config_id: %s", args[0])
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name обязателен")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations update-config")
				dr.PrintSimple("Обновить конфигурацию", fmt.Sprintf("Config ID: %d, New Name: %s", configID, name))
				return nil
			}

			req := data.UpdateConfigRequest{Name: name}
			cli := getClient(cmd)
			resp, err := cli.UpdateConfig(configID, &req)
			if err != nil {
				return fmt.Errorf("не удалось обновить конфигурацию: %w", err)
			}

			fmt.Printf("✅ Конфигурация %d обновлена\n", configID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Новое название конфигурации (обязательно)")

	return cmd
}
