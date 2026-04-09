package run

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newUpdateCmd creates the 'run update' command.
func newUpdateCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [run-id]",
		Short: "Update a test run",
		Long: `Updates an existing test run.

Only open runs can be updated. Use flags to specify changes.
Only modified fields will be sent to the API.

Examples:
	# Change name and description
	gotr run update 12345 --name "Updated Name" --description "New description"

	# Reassign to another user
	gotr run update 12345 --assigned-to 10

	# Change the set of cases in the run
	gotr run update 12345 --case-ids 100,200,300 --include-all=false

	# Dry-run mode
	gotr run update 12345 --name "Test" --dry-run`,
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

			// Collect parameters from flags (changed only)
			req := &data.UpdateRunRequest{}

			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = &name
			}
			if cmd.Flags().Changed("description") {
				description, _ := cmd.Flags().GetString("description")
				req.Description = &description
			}
			if cmd.Flags().Changed("milestone-id") {
				milestoneID, _ := cmd.Flags().GetInt64("milestone-id")
				req.MilestoneID = &milestoneID
			}
			if cmd.Flags().Changed("assigned-to") {
				assignedTo, _ := cmd.Flags().GetInt64("assigned-to")
				req.AssignedTo = &assignedTo
			}
			if cmd.Flags().Changed("case-ids") {
				caseIDs, _ := cmd.Flags().GetInt64Slice("case-ids")
				req.CaseIDs = caseIDs
			}
			if cmd.Flags().Changed("include-all") {
				includeAll, _ := cmd.Flags().GetBool("include-all")
				req.IncludeAll = &includeAll
			}

			// Check dry-run mode
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("run update")
				dr.PrintOperation(
					fmt.Sprintf("Update Run %d", runID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/update_run/%d", runID),
					req,
				)
				return nil
			}

			run, err := svc.Update(ctx, runID, req)
			if err != nil {
				return fmt.Errorf("failed to update test run: %w", err)
			}

			output.PrintSuccess(cmd, "Test run updated successfully:")
			return output.OutputResultWithFlags(cmd, run)
		},
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().Int64("milestone-id", 0, "Milestone ID")
	cmd.Flags().Int64("assigned-to", 0, "User ID to assign")
	cmd.Flags().Int64Slice("case-ids", nil, "List of case IDs (comma-separated)")
	cmd.Flags().Bool("include-all", false, "Include all suite cases")
	cmd.Flags().Bool("dry-run", false, "Show what would be executed without making actual changes")

	return cmd
}

// updateCmd is the exported command.
var updateCmd = newUpdateCmd(getClientSafe)
