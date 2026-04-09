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

// newAddCmd creates the 'datasets add' command.
// Endpoint: POST /add_dataset/{project_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [project_id]",
		Short: "Create a new dataset",
		Long: `Creates a new dataset (test data set) in the specified project.

The dataset is created with the given name. After creation, you can
add columns (parameters) and rows (values) via the web interface
or other API methods.`,
		Example: `  # Create a dataset with a name
  gotr datasets add 1 --name="Login Test Data"

  # Preview before creating
  gotr datasets add 1 --name="Test Data" --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cli := getClient(cmd)

			var projectID int64
			var err error
			if len(args) > 0 {
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id required: gotr datasets add [project_id] --name <name>")
				}
				projectID, err = resolveProjectIDInteractive(ctx, cli)
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
				dr := output.NewDryRunPrinter("datasets add")
				dr.PrintSimple("Create dataset", fmt.Sprintf("Project ID: %d, Name: %s", projectID, name))
				return nil
			}

			resp, err := cli.AddDataset(ctx, projectID, name)
			if err != nil {
				return fmt.Errorf("failed to create dataset: %w", err)
			}

			ui.Successf(os.Stdout, "Dataset created (ID: %d)", resp.ID)
			return output.OutputResult(cmd, resp, "datasets")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without creating")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Dataset name (required)")

	return cmd
}
