package variables

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates the 'variables delete' command.
// Endpoint: POST /delete_variable/{variable_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [variable_id]",
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
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var variableID int64
			if len(args) > 0 {
				var err error
				variableID, err = flags.ValidateRequiredID(args, 0, "variable_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("variable_id is required in non-interactive mode: gotr variables delete [variable_id]")
				}
				var err error
				variableID, err = resolveVariableIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
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

			ui.Successf(os.Stdout, "Variable %d deleted", variableID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
