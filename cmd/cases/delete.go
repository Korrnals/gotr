package cases

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates the 'cases delete' command.
// Endpoint: POST /delete_case/{case_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [case_id]",
		Short: "Delete a test case",
		Long:  `Deletes a test case by its ID.`,
		Example: `  # Delete a test case
  gotr cases delete 12345

  # Preview before deleting
  gotr cases delete 12345 --dry-run`,
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
					return fmt.Errorf("case_id required: gotr cases delete [case_id]")
				}

				caseID, err = resolveCaseIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("cases delete")
				dr.PrintSimple("Delete Case", fmt.Sprintf("Case ID: %d", caseID))
				return nil
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			_, err = ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Deleting case",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (struct{}, error) {
				return struct{}{}, cli.DeleteCase(ctx, caseID)
			})
			if err != nil {
				return fmt.Errorf("failed to delete case: %w", err)
			}

			ui.Successf(os.Stdout, "Case %d deleted", caseID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview what will be deleted without actually deleting")

	return cmd
}
