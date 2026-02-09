package plans

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/spf13/cobra"
)

// newDeleteCmd создаёт команду 'plans delete'
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <plan_id>",
		Short: "Удалить тест-план",
		Long:  `Удаляет тест-план по его ID.`,
		Example: `  # Удалить план
  gotr plans delete 12345

  # Проверить перед удалением
  gotr plans delete 12345 --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans delete")
				dr.PrintSimple("Delete Plan", fmt.Sprintf("Plan ID: %d", planID))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeletePlan(planID); err != nil {
				return fmt.Errorf("failed to delete plan: %w", err)
			}

			fmt.Printf("✅ Plan %d deleted\n", planID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
