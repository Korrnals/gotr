package cases

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// newListCmd creates 'cases list' command
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
		Short: "List test cases",
		Long:  `List test cases in a project with optional filtering.`,
		Example: `  gotr cases list 1
  gotr cases list 1 --suite-id=100 --section-id=50`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("invalid project_id: %s", args[0])
			}

			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			sectionID, _ := cmd.Flags().GetInt64("section-id")

			cli := getClient(cmd)
			resp, err := cli.GetCases(projectID, suiteID, sectionID)
			if err != nil {
				return fmt.Errorf("failed to list cases: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Int64("suite-id", 0, "Suite ID filter")
	cmd.Flags().Int64("section-id", 0, "Section ID filter")

	return cmd
}
