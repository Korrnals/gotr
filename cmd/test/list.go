package test

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

// newListCmd creates the command for listing tests.
func newListCmd(getClient func(cmd *cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [run-id]",
		Short: "List tests in a run",
		Long: `Retrieves a list of all tests for the specified test run.

Filters can be applied:
	--status-id      Filter by status (1=passed, 5=failed, etc.)
	--assigned-to    Filter by assigned user

Examples:
	# Get all tests in a run
	gotr test list 100

	# Get only failed tests
	gotr test list 100 --status-id 5

	# Get tests assigned to a user
	gotr test list 100 --assigned-to 10

	# Save to file
	gotr test list 100 -o tests.json
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			httpClient := getClient(cmd)
			ctx := cmd.Context()
			if httpClient == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := service.NewTestService(httpClient)

			var runID int64
			var err error

			if len(args) > 0 {
				runID, err = svc.ParseID(ctx, args, 0)
				if err != nil {
					return fmt.Errorf("invalid run ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("run_id is required in non-interactive mode: gotr test list [run_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("run_id is required in non-interactive mode: gotr test list [run_id]")
				}
				runID, err = resolveRunIDInteractive(ctx, httpClient)
				if err != nil {
					return err
				}
			}

			// Collect filters
			filters := make(map[string]string)

			if cmd.Flags().Changed("status-id") {
				statusID, _ := cmd.Flags().GetInt64("status-id")
				filters["status_id"] = strconv.FormatInt(statusID, 10)
			}

			if cmd.Flags().Changed("assigned-to") {
				assignedTo, _ := cmd.Flags().GetInt64("assigned-to")
				filters["assignedto_id"] = strconv.FormatInt(assignedTo, 10)
			}

			tests, err := svc.GetForRun(ctx, runID, filters)
			if err != nil {
				return fmt.Errorf("failed to get test list: %w", err)
			}

			// Check if output should be saved to file
			saveFlag, _ := cmd.Flags().GetBool("save")
			if saveFlag {
				filepath, err := output.Output(cmd, tests, "test", "json")
				if err != nil {
					return fmt.Errorf("save error: %w", err)
				}
				if filepath != "" {
					output.PrintSuccess(cmd, "Test list (%d) saved to %s", len(tests), filepath)
				}
				return nil
			}

			return output.OutputResultWithFlags(cmd, tests)
		},
	}

	output.AddFlag(cmd)
	cmd.Flags().Int64("status-id", 0, "Filter by status ID")
	cmd.Flags().Int64("assigned-to", 0, "Filter by assigned user ID")

	return cmd
}
