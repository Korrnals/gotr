package variables

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду 'variables add'
// Эндпоинт: POST /add_variable/{dataset_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <dataset_id>",
		Short: "Создать переменную в датасете",
		Long: `Создаёт новую переменную (колонку) в указанном датасете.

Переменная представляет собой параметр, который будет использоваться
в тест-кейсах для параметризованного тестирования.

После создания переменной можно добавлять значения через
веб-интерфейс TestRail.`,
		Example: `  # Создать переменную "username"
  gotr variables add 123 --name="username"

  # Проверить перед созданием
  gotr variables add 123 --name="password" --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			datasetID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || datasetID <= 0 {
				return fmt.Errorf("некорректный dataset_id: %s", args[0])
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name обязателен")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("variables add")
				dr.PrintSimple("Создать переменную", fmt.Sprintf("Dataset ID: %d, Name: %s", datasetID, name))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.AddVariable(datasetID, name)
			if err != nil {
				return fmt.Errorf("не удалось создать переменную: %w", err)
			}

			fmt.Printf("✅ Переменная создана (ID: %d)\n", resp.ID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без создания")
	save.AddFlag(cmd)
	cmd.Flags().String("name", "", "Название переменной (обязательно)")

	return cmd
}
