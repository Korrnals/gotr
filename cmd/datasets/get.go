package datasets

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'datasets get' command.
// Endpoint: GET /get_dataset/{dataset_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [dataset_id]",
		Short: "Get a dataset by ID",
		Long: `Retrieves detailed information about a dataset by its ID.

Includes the name, table structure (columns), and all rows
with test data for parameterized testing.`,
		Example: `  # Get dataset information
  gotr datasets get 123

  # Save to file
  gotr datasets get 456 -o dataset.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var datasetID int64
			var err error
			if len(args) > 0 {
				datasetID, err = flags.ValidateRequiredID(args, 0, "dataset_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("dataset_id required: gotr datasets get [dataset_id]")
				}
				datasetID, err = resolveDatasetIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetDataset(ctx, datasetID)
			if err != nil {
				return fmt.Errorf("failed to get dataset: %w", err)
			}

			return output.OutputResult(cmd, resp, "datasets")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
