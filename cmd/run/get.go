package run

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'run get' command.
func newGetCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [run-id]",
		Short: "Get test run information",
		Long: `Gets detailed information about a test run by its ID.

A test run is an instance of a test suite launched for test execution.
The response contains: name, description, execution statistics,
creation/update dates, assignedto_id, and other fields.

Examples:
	# Get run information
	gotr run get 12345

	# Save the result to a file
	gotr run get 12345 -o run_info.json

	# Dry-run mode
	gotr run get 12345 --dry-run
`,
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
				dr := output.NewDryRunPrinter("run get")
				dr.PrintOperation(
					fmt.Sprintf("Get Run %d", runID),
					"GET",
					fmt.Sprintf("/index.php?/api/v2/get_run/%d", runID),
					nil,
				)
				return nil
			}

			run, err := svc.Get(ctx, runID)
			if err != nil {
				return fmt.Errorf("failed to get test run: %w", err)
			}

			return output.OutputResultWithFlags(cmd, run)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be executed without making actual changes")

	return cmd
}

// getCmd is the exported command.
var getCmd = newGetCmd(getClientSafe)
