package variables

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду 'variables update'
// Эндпоинт: POST /update_variable/{variable_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <variable_id>",
		Short: "Обновить переменную",
		Long: `Обновляет название существующей переменной.

⚠️ Обратите внимание: через API можно обновить только название переменной.
Для изменения значений используйте веб-интерфейс TestRail.`,
		Example: `  # Изменить название переменной
  gotr variables update 789 --name="new_name"

  # Проверить перед обновлением
  gotr variables update 789 --name="new_name" --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			variableID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || variableID <= 0 {
				return fmt.Errorf("некорректный variable_id: %s", args[0])
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name обязателен")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("variables update")
				dr.PrintSimple("Обновить переменную", fmt.Sprintf("Variable ID: %d, New Name: %s", variableID, name))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdateVariable(variableID, name)
			if err != nil {
				return fmt.Errorf("не удалось обновить переменную: %w", err)
			}

			fmt.Printf("✅ Переменная %d обновлена\n", variableID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Новое название переменной (обязательно)")

	return cmd
}
