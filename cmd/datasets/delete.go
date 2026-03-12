package datasets

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newDeleteCmd создаёт команду 'datasets delete'
// Эндпоинт: POST /delete_dataset/{dataset_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <dataset_id>",
		Short: "Удалить датасет",
		Long: `Удаляет датасет по его ID.

⚠️ Внимание: удаление нельзя отменить! Все данные из датасета
будут безвозвратно удалены. Убедитесь, что датасет не используется
в активных тест-планах перед удалением.

Используйте --dry-run для проверки перед удалением.`,
		Example: `  # Удалить датасет
  gotr datasets delete 123

  # Проверить перед удалением
  gotr datasets delete 123 --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			datasetID, err := flags.ValidateRequiredID(args, 0, "dataset_id")
			if err != nil {
				return err
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("datasets delete")
				dr.PrintSimple("Удалить датасет", fmt.Sprintf("Dataset ID: %d", datasetID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeleteDataset(ctx, datasetID); err != nil {
				return fmt.Errorf("failed to delete dataset: %w", err)
			}

			fmt.Printf("✅ Dataset %d deleted\n", datasetID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
