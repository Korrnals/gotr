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

// newUpdateCmd creates the 'variables update' command.
// Endpoint: POST /update_variable/{variable_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [variable_id]",
		Short: "Обновить переменную",
		Long: `Обновляет название существующей переменной.

⚠️ Обратите внимание: через API можно обновить только название переменной.
Для изменения значений используйте веб-интерфейс TestRail.`,
		Example: `  # Изменить название переменной
  gotr variables update 789 --name="new_name"

  # Проверить перед обновлением
  gotr variables update 789 --name="new_name" --dry-run`,
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
					return fmt.Errorf("variable_id is required in non-interactive mode: gotr variables update [variable_id]")
				}
				var err error
				variableID, err = resolveVariableIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("variables update")
				dr.PrintSimple("Обновить переменную", fmt.Sprintf("Variable ID: %d, New Name: %s", variableID, name))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.UpdateVariable(ctx, variableID, name)
			if err != nil {
				return fmt.Errorf("failed to update variable: %w", err)
			}

			ui.Successf(os.Stdout, "Variable %d updated", variableID)
			return output.OutputResult(cmd, resp, "variables")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Новое название переменной (обязательно)")

	return cmd
}
