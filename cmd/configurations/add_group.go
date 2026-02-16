package configurations

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newAddGroupCmd создаёт команду 'configurations add-group'
// Эндпоинт: POST /add_config_group/{project_id}
func newAddGroupCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-group <project_id>",
		Short: "Создать группу конфигураций",
		Long: `Создаёт новую группу конфигураций в указанном проекте.

Группа — это контейнер для связанных конфигураций (например: "Browsers",
"Operating Systems", "Devices"). После создания группы можно добавлять
в неё отдельные конфигурации.`,
		Example: `  # Создать группу "Browsers"
  gotr configurations add-group 1 --name="Browsers"

  # Проверить перед созданием
  gotr configurations add-group 1 --name="OS" --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("некорректный project_id: %s", args[0])
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name обязателен")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("configurations add-group")
				dr.PrintSimple("Создать группу", fmt.Sprintf("Project ID: %d, Name: %s", projectID, name))
				return nil
			}

			req := data.AddConfigGroupRequest{Name: name}
			cli := getClient(cmd)
			resp, err := cli.AddConfigGroup(projectID, &req)
			if err != nil {
				return fmt.Errorf("не удалось создать группу: %w", err)
			}

			fmt.Printf("✅ Группа создана (ID: %d)\n", resp.ID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без создания")
	save.AddFlag(cmd)
	cmd.Flags().String("name", "", "Название группы (обязательно)")

	return cmd
}
