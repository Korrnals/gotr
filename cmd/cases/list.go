package cases

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'cases list' command.
// Endpoint: GET /get_cases/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "List test cases",
		Long:  `Lists test cases for a project with optional filtering.`,
		Example: `  # List all cases in a project
  gotr cases list 1

  # Filter by suite and section
  gotr cases list 1 --suite-id=100 --section-id=50`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var projectID int64
			var err error
			if len(args) > 0 {
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id required: gotr cases list [project_id]")
				}

				projectID, err = interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}
			}

			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			sectionID, _ := cmd.Flags().GetInt64("section-id")

			quiet, _ := cmd.Flags().GetBool("quiet")
			resp, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Loading cases",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (data.GetCasesResponse, error) {
				return cli.GetCases(ctx, projectID, suiteID, sectionID)
			})
			if err != nil {
				return fmt.Errorf("failed to list cases: %w", err)
			}

			return output.OutputResult(cmd, resp, "cases")
		},
	}

	cmd.Flags().Int64("suite-id", 0, "Filter by suite ID")
	cmd.Flags().Int64("section-id", 0, "Filter by section ID")

	return cmd
}
