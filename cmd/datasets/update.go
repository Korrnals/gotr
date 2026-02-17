package datasets

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду 'datasets update'
// Эндпоинт: POST /update_dataset/{dataset_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <dataset_id>",
		Short: "Обновить датасет",
		Long: `Обновляет название существующего датасета.

⚠️ Обратите внимание: через API можно обновить только название датасета.
Для изменения структуры таблицы (добавления/изменения колонок и строк)
используйте веб-интерфейс TestRail.`,
		Example: `  # Изменить название датасета
  gotr datasets update 123 --name="Новое название"

  # Проверить перед обновлением
  gotr datasets update 123 --name="Новое название" --dry-run`,
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

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("datasets update")
				dr.PrintSimple("Обновить датасет", fmt.Sprintf("Dataset ID: %d, New Name: %s", datasetID, name))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdateDataset(datasetID, name)
			if err != nil {
				return fmt.Errorf("не удалось обновить датасет: %w", err)
			}

			fmt.Printf("✅ Датасет %d обновлён\n", datasetID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Новое название датасета (обязательно)")

	return cmd
}
