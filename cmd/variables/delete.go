package variables

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newDeleteCmd создаёт команду 'variables delete'
// Эндпоинт: POST /delete_variable/{variable_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <variable_id>",
		Short: "Удалить переменную",
		Long: `Удаляет переменную из датасета.

⚠️ Внимание: удаление нельзя отменить! Все значения этой переменной
будут безвозвратно удалены. Убедитесь, что переменная не используется
в активных тест-кейсах.

Используйте --dry-run для проверки перед удалением.`,
		Example: `  # Удалить переменную
  gotr variables delete 789

  # Проверить перед удалением
  gotr variables delete 789 --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			variableID, err := flags.ValidateRequiredID(args, 0, "variable_id")
			if err != nil {
				return err
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("variables delete")
				dr.PrintSimple("Удалить переменную", fmt.Sprintf("Variable ID: %d", variableID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeleteVariable(ctx, variableID); err != nil {
				return fmt.Errorf("failed to delete variable: %w", err)
			}

			fmt.Printf("✅ Variable %d deleted\n", variableID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
