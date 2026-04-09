package variables

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddCmd creates the 'variables add' command.
// Endpoint: POST /add_variable/{dataset_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [dataset_id]",
		Short: "Create a variable in a dataset",
		Long: `Creates a new variable (column) in the specified dataset.

A variable represents a parameter that will be used
in test cases for parameterized testing.

After creating a variable, you can add values through
the TestRail web interface.`,
		Example: `  # Create a variable "username"
  gotr variables add 123 --name="username"

  # Preview before creating
  gotr variables add 123 --name="password" --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var datasetID int64
			if len(args) > 0 {
				var err error
				datasetID, err = flags.ValidateRequiredID(args, 0, "dataset_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("dataset_id is required in non-interactive mode: gotr variables add [dataset_id]")
				}
				var err error
				datasetID, err = resolveDatasetIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("variables add")
				dr.PrintSimple("Create Variable", fmt.Sprintf("Dataset ID: %d, Name: %s", datasetID, name))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.AddVariable(ctx, datasetID, name)
			if err != nil {
				return fmt.Errorf("failed to create variable: %w", err)
			}

			ui.Successf(os.Stdout, "Variable created (ID: %d)", resp.ID)
			return output.OutputResult(cmd, resp, "variables")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without creating")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Variable name (required)")

	return cmd
}
