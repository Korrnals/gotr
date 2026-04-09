package bdds

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'bdds get' command.
// Endpoint: GET /get_bdd/{test_case_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [case_id]",
		Short: "Get a BDD scenario for a test case",
		Long: `Retrieve the BDD scenario linked to the specified test case.

Returns the Gherkin scenario in Given-When-Then format,
if one has been added to the test case.`,
		Example: `  # Get BDD for a case
  gotr bdds get 12345

  # Save to a file
  gotr bdds get 12345 -o bdd.feature`,
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
					return fmt.Errorf("case_id is required in non-interactive mode: gotr bdds get [case_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("case_id is required in non-interactive mode: gotr bdds get [case_id]")
				}
				caseID, err = resolveCaseIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetBDD(ctx, caseID)
			if err != nil {
				return fmt.Errorf("failed to get BDD: %w", err)
			}

			return output.OutputResult(cmd, resp, "bdds")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
