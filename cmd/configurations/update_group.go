package configurations

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newUpdateGroupCmd создаёт команду 'configurations update-group'
// Эндпоинт: POST /update_config_group/{group_id}
func newUpdateGroupCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group <group_id>",
		Short: "Обновить группу конфигураций",
		Long: `Обновляет название существующей группы конфигураций.`,
		Example: `  # Изменить название группы
  gotr configurations update-group 5 --name="Новое название"

  # Проверить перед обновлением
  gotr configurations update-group 5 --name="Новое название" --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || groupID <= 0 {
				return fmt.Errorf("некорректный group_id: %s", args[0])
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name обязателен")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("configurations update-group")
				dr.PrintSimple("Обновить группу", fmt.Sprintf("Group ID: %d, New Name: %s", groupID, name))
				return nil
			}

			req := data.UpdateConfigGroupRequest{Name: name}
			cli := getClient(cmd)
			resp, err := cli.UpdateConfigGroup(groupID, &req)
			if err != nil {
				return fmt.Errorf("не удалось обновить группу: %w", err)
			}

			fmt.Printf("✅ Группа %d обновлена\n", groupID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")
	cmd.Flags().String("name", "", "Новое название группы (обязательно)")

	return cmd
}
