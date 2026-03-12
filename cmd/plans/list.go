package plans

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'plans list'
// Эндпоинт: GET /get_plans/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
		Short: "Список тест-планов",
		Long:  `Выводит список всех тест-планов проекта.`,
		Example: `  # Список планов проекта
  gotr plans list 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := flags.ValidateRequiredID(args, 0, "project_id")
			if err != nil {
				return err
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetPlans(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to list plans: %w", err)
			}

			_, err = output.Output(cmd, resp, "plans", "json")
			return err
		},
	}

	output.AddFlag(cmd)

	return cmd
}
