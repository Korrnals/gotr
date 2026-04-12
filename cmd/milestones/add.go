package milestones

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

// newAddCmd creates the 'milestones add' command.
// Endpoint: POST /add_milestone/{project_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [project_id]",
		Short: "Create a new milestone",
		Long: `Creates a new milestone in the specified project.

A milestone is a development stage to which test runs are linked.
You can specify a deadline, description, and parent milestone for hierarchy.

Usage examples:
  # Create a simple milestone
  gotr milestones add 1 --name="Release 1.0"

  # Milestone with deadline and description
  gotr milestones add 1 --name="Sprint 5" --due-on="2026-03-15" --description="Sprint goal"

  # Nested milestone (sub-stage)
  gotr milestones add 1 --name="Iteration 1.1" --parent-id=123`,
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
				ctx := cmd.Context()
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr milestones add [project_id]")
				}
				cli := getClient(cmd)
				var err error
				projectID, err = resolveProjectIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			req := data.AddMilestoneRequest{
				Name: name,
			}

			if v, _ := cmd.Flags().GetString("description"); v != "" {
				req.Description = v
			}
			if v, _ := cmd.Flags().GetString("due-on"); v != "" {
				req.DueOn = v
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("milestones add")
				dr.PrintSimple("Create Milestone", fmt.Sprintf("Project ID: %d, Name: %s", projectID, req.Name))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			quiet, _ := cmd.Flags().GetBool("quiet")
			resp, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Creating milestone",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*data.Milestone, error) {
				return cli.AddMilestone(ctx, projectID, &req)
			})
			if err != nil {
				return fmt.Errorf("failed to create milestone: %w", err)
			}

			ui.Successf(os.Stdout, "Milestone created (ID: %d)", resp.ID)
			return output.OutputResult(cmd, resp, "milestones")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without actually executing")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Milestone name (required)")
	cmd.Flags().String("description", "", "Milestone description")
	cmd.Flags().String("due-on", "", "Deadline in YYYY-MM-DD format")

	return cmd
}
