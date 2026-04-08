package bdds

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddCmd creates the 'bdds add' command.
// Endpoint: POST /add_bdd/{test_case_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [case_id]",
		Short: "Add a BDD scenario to a test case",
		Long: `Add a Gherkin-format BDD scenario to the specified test case.

The scenario must follow the Given-When-Then format.
Content can be provided via a file or directly.`,
		Example: `  # Add BDD from a file
  gotr bdds add 12345 --file=scenario.feature

  # Add BDD from stdin
  cat scenario.feature | gotr bdds add 12345`,
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
					return fmt.Errorf("case_id is required in non-interactive mode: gotr bdds add [case_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("case_id is required in non-interactive mode: gotr bdds add [case_id]")
				}
				caseID, err = resolveCaseIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			// Read the BDD content
			content, err := readBDDContent(cmd)
			if err != nil {
				return err
			}
			if content == "" {
				return fmt.Errorf("BDD content cannot be empty (use --file or stdin)")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("bdds add")
				dr.PrintSimple("Add BDD", fmt.Sprintf("Case ID: %d", caseID))
				return nil
			}

			resp, err := cli.AddBDD(ctx, caseID, content)
			if err != nil {
				return fmt.Errorf("failed to add BDD: %w", err)
			}

			ui.Successf(os.Stdout, "BDD added to case %d", caseID)
			return output.OutputResult(cmd, resp, "bdds")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	output.AddFlag(cmd)
	cmd.Flags().String("file", "", "Path to a Gherkin scenario file")

	return cmd
}

// readBDDContent reads BDD content from a file or stdin.
func readBDDContent(cmd *cobra.Command) (string, error) {
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return string(data), nil
	}

	// TODO: Read from stdin when no file is specified.
	// For now, return empty string; validation will report the error.
	return "", nil
}
