package tests

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'tests list' command for listing tests in a run.
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [run_id]",
		Short: "List tests in a run",
		Long: `Get a list of all tests for the specified test run.

Use the --status-id flag to filter by status.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getClient(cmd)
			ctx := cmd.Context()

			var runID int64
			var err error
			if len(args) > 0 {
				runID, err = flags.ValidateRequiredID(args, 0, "run_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("run_id is required in non-interactive mode: gotr tests list [run_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("run_id is required in non-interactive mode: gotr tests list [run_id]")
				}
				runID, err = resolveRunIDInteractive(ctx, client)
				if err != nil {
					return err
				}
			}

			statusID, _ := cmd.Flags().GetInt64("status-id")

			filters := make(map[string]string)
			if statusID > 0 {
				filters["status_id"] = fmt.Sprintf("%d", statusID)
			}

			// Check dry-run mode
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("tests list")
				details := fmt.Sprintf("Run ID: %d", runID)
				if statusID > 0 {
					details += fmt.Sprintf(", Status ID: %d", statusID)
				}
				dr.PrintSimple("Get test list", details)
				return nil
			}

			tests, err := client.GetTests(ctx, runID, filters)
			if err != nil {
				return err
			}

			return output.OutputResult(cmd, tests, "tests")
		},
	}

	cmd.Flags().Int64("status-id", 0, "Status ID for filtering")
	cmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	output.AddFlag(cmd)

	return cmd
}
