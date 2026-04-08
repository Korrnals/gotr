package run

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'run list' command.
func newListCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project-id]",
		Short: "Get list of project test runs",
		Long: `Gets the list of all test runs for the specified project.

The list contains active and completed runs with basic information:
ID, name, description, test statistics (passed/failed/blocked).

If project-id is not specified, an interactive selection from the project list will be offered.

Examples:
	# Get project runs list (with interactive selection)
	gotr run list

	# Get project runs list (with explicit ID)
	gotr run list 30

	# Save to file for further processing
	gotr run list 30 -o runs.json

	# Dry-run mode
	gotr run list 30 --dry-run
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := newRunServiceFromInterface(cli)

			var projectID int64
			var err error

			if len(args) > 0 {
				// project-id explicitly provided
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				// Interactive project selection
				projectID, err = interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}
			}

			// Check dry-run mode
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("run list")
				dr.PrintOperation(
					fmt.Sprintf("List Runs for Project %d", projectID),
					"GET",
					fmt.Sprintf("/index.php?/api/v2/get_runs/%d", projectID),
					nil,
				)
				return nil
			}

			runs, err := svc.GetByProject(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to get test runs list: %w", err)
			}

			return output.OutputResultWithFlags(cmd, runs)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be executed without making actual changes")

	return cmd
}

// listCmd is the exported command.
var listCmd = newListCmd(getClientSafe)
