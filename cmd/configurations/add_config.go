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

// newAddConfigCmd создаёт команду 'configurations add-config'
// Эндпоинт: POST /add_config/{group_id}
func newAddConfigCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-config [group_id]",
		Short: "Добавить конфигурацию в группу",
		Long: `Добавляет новую конфигурацию в существующую группу.

Конфигурация — это конкретное значение (например: "Chrome", "Windows 10",
"iPhone 12") в рамках группы. Конфигурации используются при создании
тест-планов с множественными конфигурациями.`,
		Example: `  # Добавить "Chrome" в группу 5
  gotr configurations add-config 5 --name="Chrome"

  # Проверить перед добавлением
  gotr configurations add-config 5 --name="Firefox" --dry-run`,
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
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations add-config [group_id]")
				}
				if _, ok := interactive.PrompterFromContext(ctx).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations add-config [group_id]")
				}

				groupID, err = resolveGroupIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
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
			resp, err := cli.AddConfig(ctx, groupID, &req)
			if err != nil {
				return fmt.Errorf("failed to add configuration: %w", err)
			}

			ui.Successf(os.Stdout, "Configuration added (ID: %d)", resp.ID)
			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без добавления")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Название конфигурации (обязательно)")

	return cmd
}
