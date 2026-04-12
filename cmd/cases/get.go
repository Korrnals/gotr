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

// newGetCmd creates the 'cases get' command.
// Endpoint: GET /get_case/{case_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "get [case_id]",
		Short: "Get a test case by ID",
		Long:  `Retrieves detailed information about a test case by its identifier.`,
		Example: `  # Get case information
  gotr cases get 12345

  # Save to file
  gotr cases get 12345 -o case.json`,
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
					return fmt.Errorf("case_id required: gotr cases get [case_id]")
				}

				caseID, err = resolveCaseIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			resp, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Loading case",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*data.Case, error) {
				return cli.GetCase(ctx, caseID)
			})
			if err != nil {
				return fmt.Errorf("failed to get case: %w", err)
			}

			return output.OutputResult(cmd, resp, "cases")
		},
	}
}
