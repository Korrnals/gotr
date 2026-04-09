package variables

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'variables list' command.
// Endpoint: GET /get_variables/{dataset_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [dataset_id]",
		Short: "List dataset variables",
		Long: `Displays the list of variables defined in the specified dataset.

Variables are used for parameterized testing
and represent columns in the test data table.`,
		Example: `  # Get list of dataset variables
  gotr variables list 123

  # Save to a file
  gotr variables list 456 -o vars.json`,
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
					return fmt.Errorf("dataset_id is required in non-interactive mode: gotr variables list [dataset_id]")
				}
				var err error
				datasetID, err = resolveDatasetIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetVariables(ctx, datasetID)
			if err != nil {
				return fmt.Errorf("failed to get variables list: %w", err)
			}

			return output.OutputResult(cmd, resp, "variables")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
