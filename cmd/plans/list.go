package plans

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'plans list' command.
// Endpoint: GET /get_plans/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "Список тест-планов",
		Long:  `Выводит список всех тест-планов проекта.`,
		Example: `  # Список планов проекта
  gotr plans list 1`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectID int64
			if len(args) > 0 {
				var err error
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr plans list [project_id]")
				}
				var err error
				projectID, err = resolveProjectIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
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
