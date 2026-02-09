package plans

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// newListCmd creates 'plans list' command
func newListCmd(getClient GetClientFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "list <project_id>",
		Short: "List test plans",
		Long:  `List all test plans in a project.`,
		Example: `  gotr plans list 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("invalid project_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetPlans(projectID)
			if err != nil {
				return fmt.Errorf("failed to list plans: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}
}
