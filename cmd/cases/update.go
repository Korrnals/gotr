package cases

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newUpdateCmd creates the 'cases update' command.
// Endpoint: POST /update_case/{case_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [case_id]",
		Short: "Update a test case",
		Long:  `Updates an existing test case.`,
		Example: `  # Update title and priority
  gotr cases update 12345 --title="New title" --priority-id=2

  # Update from a JSON file
  gotr cases update 12345 --json-file=update.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var caseID int64
			var err error
			if len(args) > 0 {
				caseID, err = flags.ValidateRequiredID(args, 0, "case_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("case_id required: gotr cases update [case_id]")
				}

				caseID, err = resolveCaseIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			jsonFile, _ := cmd.Flags().GetString("json-file")
			var req data.UpdateCaseRequest

			if jsonFile != "" {
				jsonData, err := os.ReadFile(jsonFile)
				if err != nil {
					return fmt.Errorf("error reading JSON file: %w", err)
				}
				if err := json.Unmarshal(jsonData, &req); err != nil {
					return fmt.Errorf("error parsing JSON: %w", err)
				}
			} else {
				title, _ := cmd.Flags().GetString("title")
				if title != "" {
					req.Title = &title
				}
				if v, _ := cmd.Flags().GetInt64("type-id"); v > 0 {
					req.TypeID = &v
				}
				if v, _ := cmd.Flags().GetInt64("priority-id"); v > 0 {
					req.PriorityID = &v
				}
				if v, _ := cmd.Flags().GetString("refs"); v != "" {
					req.Refs = &v
				}
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("cases update")
				dr.PrintSimple("Update Case", fmt.Sprintf("Case ID: %d", caseID))
				return nil
			}

			resp, err := cli.UpdateCase(ctx, caseID, &req)
			if err != nil {
				return fmt.Errorf("failed to update case: %w", err)
			}

			ui.Successf(os.Stdout, "Case %d updated", caseID)
			return output.OutputResult(cmd, resp, "cases")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	output.AddFlag(cmd)
	cmd.Flags().String("json-file", "", "Path to a JSON file with update data")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().Int64("type-id", 0, "New type ID")
	cmd.Flags().Int64("priority-id", 0, "New priority ID")
	cmd.Flags().String("refs", "", "New references")

	return cmd
}
