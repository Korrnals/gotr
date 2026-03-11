package cases

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newDeleteCmd создаёт команду 'cases delete'
// Эндпоинт: POST /delete_case/{case_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <case_id>",
		Short: "Удалить тест-кейс",
		Long:  `Удаляет тест-кейс по его ID.`,
		Example: `  # Удалить тест-кейс
  gotr cases delete 12345

  # Проверить перед удалением
  gotr cases delete 12345 --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			caseID, err := flags.ValidateRequiredID(args, 0, "case_id")
			if err != nil {
				return err
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("cases delete")
				dr.PrintSimple("Delete Case", fmt.Sprintf("Case ID: %d", caseID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeleteCase(ctx, caseID); err != nil {
				return fmt.Errorf("failed to delete case: %w", err)
			}

			fmt.Printf("✅ Case %d deleted\n", caseID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
