package tests

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'tests get' command for retrieving test details.
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [test_id]",
		Short: "Get test information",
		Long: `Get detailed information about a test by its ID.

A test represents a specific execution of a test case
within a test run.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getClient(cmd)
			ctx := cmd.Context()

			var testID int64
			var err error
			if len(args) > 0 {
				testID, err = flags.ValidateRequiredID(args, 0, "test_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr tests get [test_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr tests get [test_id]")
				}
				testID, err = resolveTestIDInteractive(ctx, client)
				if err != nil {
					return err
				}
			}

			// Check dry-run mode
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("tests get")
				dr.PrintSimple("Get test information", fmt.Sprintf("Test ID: %d", testID))
				return nil
			}

			test, err := client.GetTest(ctx, testID)
			if err != nil {
				return err
			}

			return output.OutputResult(cmd, test, "tests")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	output.AddFlag(cmd)

	return cmd
}
