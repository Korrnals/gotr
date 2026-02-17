package variables

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
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
			variableID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || variableID <= 0 {
				return fmt.Errorf("некорректный variable_id: %s", args[0])
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("variables delete")
				dr.PrintSimple("Удалить переменную", fmt.Sprintf("Variable ID: %d", variableID))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeleteVariable(variableID); err != nil {
				return fmt.Errorf("не удалось удалить переменную: %w", err)
			}

			fmt.Printf("✅ Переменная %d удалена\n", variableID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
