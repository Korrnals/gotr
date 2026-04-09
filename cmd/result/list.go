package result

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'result list' command.
// Endpoint: GET /get_results_for_run/{run_id}
func newListCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "list [run-id]",
		Short: "Get results for a test run",
		Long: `Gets the list of results for the specified test run.

If run-id is not specified, an interactive selection will be offered:
1. Select a project from the list
2. Select a test run from the project

Examples:
	# Get results with interactive run selection
	gotr result list

	# Get results for a specific run
	gotr result list 12345

	# Save to file
	gotr result list 12345 -o results.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := newResultServiceFromInterface(cli)

			var runID int64
			var err error

			if len(args) > 0 {
				// Explicit run-id provided
				runID, err = flags.ValidateRequiredID(args, 0, "run")
				if err != nil {
					return err
				}
			} else {
				// Interactive selection: project -> run
				p := interactive.PrompterFromContext(ctx)
				projectID, err := interactive.SelectProject(ctx, p, cli, "")
				if err != nil {
					return err
				}

				// Fetch project runs
				runs, err := svc.GetRunsForProject(ctx, projectID)
				if err != nil {
					return fmt.Errorf("failed to get runs list: %w", err)
				}

				if len(runs) == 0 {
					return fmt.Errorf("no test runs found in project %d", projectID)
				}

				// Select run interactively
				runID, err = interactive.SelectRun(ctx, p, runs, "")
				if err != nil {
					return err
				}
			}

			results, err := svc.GetForRun(ctx, runID)
			if err != nil {
				return fmt.Errorf("failed to get results: %w", err)
			}

			return output.OutputResultWithFlags(cmd, results)
		},
	}
}

// Backward compatibility: exported var for registration in result.go
var listCmd = newListCmd(getClientSafe)
