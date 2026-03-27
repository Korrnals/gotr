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

// newUpdateGroupCmd создаёт команду 'configurations update-group'
// Эндпоинт: POST /update_config_group/{group_id}
func newUpdateGroupCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group [group_id]",
		Short: "Обновить группу конфигураций",
		Long:  `Обновляет название существующей группы конфигураций.`,
		Example: `  # Изменить название группы
  gotr configurations update-group 5 --name="Новое название"

  # Проверить перед обновлением
  gotr configurations update-group 5 --name="Новое название" --dry-run`,
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
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations update-group [group_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations update-group [group_id]")
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
				dr := output.NewDryRunPrinter("configurations update-group")
				dr.PrintSimple("Обновить группу", fmt.Sprintf("Group ID: %d, New Name: %s", groupID, name))
				return nil
			}

			req := data.UpdateConfigGroupRequest{Name: name}
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
