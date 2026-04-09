package plans

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddCmd creates the 'plans add' command.
// Endpoint: POST /add_plan/{project_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [project_id]",
		Short: "Create a new test plan",
		Long:  `Creates a new test plan in the specified project.`,
		Example: `  # Create a sprint plan
  gotr plans add 1 --name="Sprint 1 Plan"

  # Create a regression plan with description
  gotr plans add 1 --name="Regression" --description="Full regression test suite"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectID int64
			if len(args) > 0 {
				var err error
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr plans add [project_id]")
				}
				var err error
				projectID, err = resolveProjectIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			req := data.AddPlanRequest{
				Name: name,
			}

			if v, _ := cmd.Flags().GetString("description"); v != "" {
				req.Description = v
			}
			if v, _ := cmd.Flags().GetInt64("milestone-id"); v > 0 {
				req.MilestoneID = v
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("plans add")
				dr.PrintSimple("Create Plan", fmt.Sprintf("Project ID: %d, Name: %s", projectID, req.Name))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.AddPlan(ctx, projectID, &req)
			if err != nil {
				return fmt.Errorf("failed to create plan: %w", err)
			}

			ui.Successf(os.Stdout, "Plan created (ID: %d)", resp.ID)
			return output.OutputResult(cmd, resp, "plans")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without creating")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Plan name (required)")
	cmd.Flags().String("description", "", "Plan description")
	cmd.Flags().Int64("milestone-id", 0, "Milestone ID")

	return cmd
}
