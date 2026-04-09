package run

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newCloseCmd creates the 'run close' command.
func newCloseCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close [run-id]",
		Short: "Close a test run",
		Long: `Closes a test run (marks it as completed).

A closed test run:
- Cannot be modified (update will return an error)
- Cannot have new test results added
- Is preserved in the system for history and reporting
- The is_completed field becomes true

This action is reversible — the run can be reopened via the TestRail web interface.

Examples:
	# Close a run after testing is complete
	gotr run close 12345

	# Close and save the closed run information
	gotr run close 12345 -o closed_run.json

	# Dry-run mode
	gotr run close 12345 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			runID, err := resolveRunID(ctx, cli, args)
			if err != nil {
				return fmt.Errorf("invalid test run ID: %w", err)
			}

			svc := newRunServiceFromInterface(cli)

			// Check dry-run mode
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("run close")
				dr.PrintOperation(
					fmt.Sprintf("Close Run %d", runID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/close_run/%d", runID),
					nil,
				)
				return nil
			}

			run, err := svc.Close(ctx, runID)
			if err != nil {
				return fmt.Errorf("failed to close test run: %w", err)
			}

			output.PrintSuccess(cmd, "Test run closed successfully:")
			return output.OutputResultWithFlags(cmd, run)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be executed without making actual changes")

	return cmd
}

// closeCmd is the exported command.
var closeCmd = newCloseCmd(getClientSafe)
