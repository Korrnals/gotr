package cases

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates 'cases delete' command
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <case_id>",
		Short: "Delete a test case",
		Long:  `Delete a test case by ID.`,
		Example: `  gotr cases delete 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			caseID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || caseID <= 0 {
				return fmt.Errorf("invalid case_id: %s", args[0])
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases delete")
				dr.PrintSimple("Delete Case", fmt.Sprintf("Case ID: %d", caseID))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeleteCase(caseID); err != nil {
				return fmt.Errorf("failed to delete case: %w", err)
			}

			fmt.Printf("✅ Case %d deleted\n", caseID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")

	return cmd
}
