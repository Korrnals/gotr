package datasets

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates the 'datasets delete' command.
// Endpoint: POST /delete_dataset/{dataset_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [dataset_id]",
		Short: "Delete a dataset",
		Long: `Deletes a dataset by its ID.

⚠️ Warning: deletion cannot be undone! All data in the dataset
will be permanently removed. Make sure the dataset is not used
in active test plans before deleting.

Use --dry-run to preview before deleting.`,
		Example: `  # Delete a dataset
  gotr datasets delete 123

  # Preview before deleting
  gotr datasets delete 123 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cli := getClient(cmd)

			var datasetID int64
			var err error
			if len(args) > 0 {
				datasetID, err = flags.ValidateRequiredID(args, 0, "dataset_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("dataset_id required: gotr datasets delete [dataset_id]")
				}
				datasetID, err = resolveDatasetIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("datasets delete")
				dr.PrintSimple("Delete dataset", fmt.Sprintf("Dataset ID: %d", datasetID))
				return nil
			}

			if err := cli.DeleteDataset(ctx, datasetID); err != nil {
				return fmt.Errorf("failed to delete dataset: %w", err)
			}

			ui.Successf(os.Stdout, "Dataset %d deleted", datasetID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be deleted without actually deleting")

	return cmd
}
