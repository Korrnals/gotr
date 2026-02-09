package plans

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// newGetCmd creates 'plans get' command
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "get <plan_id>",
		Short: "Get a test plan by ID",
		Long:  `Retrieve details of a specific test plan including entries.`,
		Example: `  gotr plans get 12345
  gotr plans get 12345 -o plan.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetPlan(planID)
			if err != nil {
				return fmt.Errorf("failed to get plan: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}
}
