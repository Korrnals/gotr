package plans

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'plans get' command.
// Endpoint: GET /get_plan/{plan_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [plan_id]",
		Short: "Получить тест-план по ID",
		Long:  `Получает детальную информацию о тест-плане, включая записи (entries).`,
		Example: `  # Получить информацию о плане
  gotr plans get 12345

  # Сохранить в файл
  gotr plans get 12345 --save`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var planID int64
			if len(args) > 0 {
				var err error
				planID, err = flags.ValidateRequiredID(args, 0, "plan_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("plan_id is required in non-interactive mode: gotr plans get [plan_id]")
				}
				var err error
				planID, err = resolvePlanIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetPlan(ctx, planID)
			if err != nil {
				return fmt.Errorf("failed to get plan: %w", err)
			}

			return output.OutputResult(cmd, resp, "plans")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
