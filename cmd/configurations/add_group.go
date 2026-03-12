package configurations

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
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
			projectID, err := flags.ValidateRequiredID(args, 0, "project_id")
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations add-group")
				dr.PrintSimple("Создать группу", fmt.Sprintf("Project ID: %d, Name: %s", projectID, name))
				return nil
			}

			req := data.AddConfigGroupRequest{Name: name}
			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.AddConfigGroup(ctx, projectID, &req)
			if err != nil {
				return fmt.Errorf("failed to create group: %w", err)
			}

			fmt.Printf("✅ Group created (ID: %d)\n", resp.ID)
			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без создания")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Название группы (обязательно)")

	return cmd
}
