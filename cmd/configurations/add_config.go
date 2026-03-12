package configurations

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newAddConfigCmd создаёт команду 'configurations add-config'
// Эндпоинт: POST /add_config/{group_id}
func newAddConfigCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-config <group_id>",
		Short: "Добавить конфигурацию в группу",
		Long: `Добавляет новую конфигурацию в существующую группу.

Конфигурация — это конкретное значение (например: "Chrome", "Windows 10",
"iPhone 12") в рамках группы. Конфигурации используются при создании
тест-планов с множественными конфигурациями.`,
		Example: `  # Добавить "Chrome" в группу 5
  gotr configurations add-config 5 --name="Chrome"

  # Проверить перед добавлением
  gotr configurations add-config 5 --name="Firefox" --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID, err := flags.ValidateRequiredID(args, 0, "group_id")
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations add-config")
				dr.PrintSimple("Добавить конфигурацию", fmt.Sprintf("Group ID: %d, Name: %s", groupID, name))
				return nil
			}

			req := data.AddConfigRequest{Name: name}
			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.AddConfig(ctx, groupID, &req)
			if err != nil {
				return fmt.Errorf("failed to add configuration: %w", err)
			}

			fmt.Printf("✅ Configuration added (ID: %d)\n", resp.ID)
			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без добавления")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Название конфигурации (обязательно)")

	return cmd
}
