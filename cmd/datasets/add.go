package datasets

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду 'datasets add'
// Эндпоинт: POST /add_dataset/{project_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <project_id>",
		Short: "Создать новый датасет",
		Long: `Создаёт новый датасет (набор тестовых данных) в указанном проекте.

Датасет создаётся с указанным названием. После создания можно
добавлять колонки (параметры) и строки (значения) через веб-интерфейс
или другие API методы.`,
		Example: `  # Создать датасет с названием
  gotr datasets add 1 --name="Login Test Data"

  # Проверить перед созданием
  gotr datasets add 1 --name="Test Data" --dry-run`,
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

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("datasets add")
				dr.PrintSimple("Создать датасет", fmt.Sprintf("Project ID: %d, Name: %s", projectID, name))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.AddDataset(projectID, name)
			if err != nil {
				return fmt.Errorf("не удалось создать датасет: %w", err)
			}

			fmt.Printf("✅ Датасет создан (ID: %d)\n", resp.ID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без создания")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")
	cmd.Flags().String("name", "", "Название датасета (обязательно)")

	return cmd
}
