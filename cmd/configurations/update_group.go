package configurations

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newUpdateGroupCmd создаёт команду 'configurations update-group'
// Эндпоинт: POST /update_config_group/{group_id}
func newUpdateGroupCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group <group_id>",
		Short: "Обновить группу конфигураций",
		Long:  `Обновляет название существующей группы конфигураций.`,
		Example: `  # Изменить название группы
  gotr configurations update-group 5 --name="Новое название"

  # Проверить перед обновлением
  gotr configurations update-group 5 --name="Новое название" --dry-run`,
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
				dr := output.NewDryRunPrinter("configurations update-group")
				dr.PrintSimple("Обновить группу", fmt.Sprintf("Group ID: %d, New Name: %s", groupID, name))
				return nil
			}

			req := data.UpdateConfigGroupRequest{Name: name}
			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.UpdateConfigGroup(ctx, groupID, &req)
			if err != nil {
				return fmt.Errorf("failed to update group: %w", err)
			}

			ui.Successf(os.Stdout, "Group %d updated", groupID)
			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Новое название группы (обязательно)")

	return cmd
}
