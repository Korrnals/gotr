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

// newUpdateCmd creates the 'datasets update' command.
// Endpoint: POST /update_dataset/{dataset_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [dataset_id]",
		Short: "Update a dataset",
		Long: `Updates the name of an existing dataset.

⚠️ Note: only the dataset name can be updated via the API.
To modify the table structure (add/change columns and rows),
use the TestRail web interface.`,
		Example: `  # Change dataset name
  gotr datasets update 123 --name="New Name"

  # Preview before updating
  gotr datasets update 123 --name="New Name" --dry-run`,
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
					return fmt.Errorf("dataset_id required: gotr datasets update [dataset_id] --name <name>")
				}
				datasetID, err = resolveDatasetIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("datasets update")
				dr.PrintSimple("Update dataset", fmt.Sprintf("Dataset ID: %d, New Name: %s", datasetID, name))
				return nil
			}

			resp, err := cli.UpdateDataset(ctx, datasetID, name)
			if err != nil {
				return fmt.Errorf("failed to update dataset: %w", err)
			}

			ui.Successf(os.Stdout, "Dataset %d updated", datasetID)
			return output.OutputResult(cmd, resp, "datasets")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "New dataset name (required)")

	return cmd
}
