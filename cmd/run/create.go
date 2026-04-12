package run

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newCreateCmd creates the create command for test runs (also used in tests).
func newCreateCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [project-id]",
		Short: "Create a new test run",
		Long: `Creates a new test run in the specified project.

A test run is created based on a test suite. You can specify:
- name and description
- milestone to link to
- user to assign (assignedto_id)
- specific case_ids (if not all suite cases are needed)
- config_ids for configuration testing

Examples:
	# Create a run with minimal parameters
	gotr run create 30 --suite-id 20069 --name "Smoke Tests"

	# Create a run with description and assignment
	gotr run create 30 --suite-id 20069 --name "Regression" \\
		--description "Full regression suite" --assigned-to 5

	# Create a run with specific cases only
	gotr run create 30 --suite-id 20069 --name "Critical Path" \\
		--case-ids 123,456,789

	# Dry-run mode
	gotr run create 30 --suite-id 20069 --name "Test" --dry-run`,
		Args: cobra.MaximumNArgs(1),
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
				projectID, err = svc.ParseID(ctx, args, 0)
				if err != nil {
					return fmt.Errorf("invalid project ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr run create [project-id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr run create [project-id]")
				}
				projectID, err = interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}
			}

			// Collect parameters from flags
			name, _ := cmd.Flags().GetString("name")
			description, _ := cmd.Flags().GetString("description")
			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			if suiteID <= 0 {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("suite-id is required in non-interactive mode: gotr run create [project-id] --suite-id <suite_id>")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("suite-id is required in non-interactive mode: gotr run create [project-id] --suite-id <suite_id>")
				}
				suiteID, err = interactive.SelectSuiteForProject(ctx, interactive.PrompterFromContext(ctx), cli, projectID, "")
				if err != nil {
					return err
				}
			}
			milestoneID, _ := cmd.Flags().GetInt64("milestone-id")
			assignedTo, _ := cmd.Flags().GetInt64("assigned-to")
			caseIDs, _ := cmd.Flags().GetInt64Slice("case-ids")
			configIDs, _ := cmd.Flags().GetInt64Slice("config-ids")
			includeAll, _ := cmd.Flags().GetBool("include-all")

			req := &data.AddRunRequest{
				Name:        name,
				Description: description,
				SuiteID:     suiteID,
				MilestoneID: milestoneID,
				AssignedTo:  assignedTo,
				CaseIDs:     caseIDs,
				ConfigIDs:   configIDs,
				IncludeAll:  includeAll,
			}

			// Check dry-run mode
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("run create")
				dr.PrintOperation(
					fmt.Sprintf("Create Run in Project %d", projectID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/add_run/%d", projectID),
					req,
				)
				return nil
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			run, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Creating run",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*data.Run, error) {
				return svc.Create(ctx, projectID, req)
			})
			if err != nil {
				return fmt.Errorf("failed to create test run: %w", err)
			}

			output.PrintSuccess(cmd, "Test run created successfully (ID: %d):", run.ID)
			return output.OutputResultWithFlags(cmd, run)
		},
	}

	cmd.Flags().Int64P("suite-id", "s", 0, "Test suite ID (required)")
	cmd.Flags().String("name", "", "Test run name (required)")
	cmd.Flags().String("description", "", "Test run description")
	cmd.Flags().Int64("milestone-id", 0, "Milestone ID")
	cmd.Flags().Int64("assigned-to", 0, "User ID to assign")
	cmd.Flags().Int64Slice("case-ids", nil, "List of case IDs to include (comma-separated)")
	cmd.Flags().Int64Slice("config-ids", nil, "List of configuration IDs (comma-separated)")
	cmd.Flags().Bool("include-all", true, "Include all suite cases")
	cmd.Flags().Bool("dry-run", false, "Show what would be executed without making actual changes")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// createCmd is the exported command for registration.
var createCmd = newCreateCmd(getClientSafe)
