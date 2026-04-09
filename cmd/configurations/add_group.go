package configurations

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddGroupCmd creates the 'configurations add-group' command.
// Endpoint: POST /add_config_group/{project_id}
func newAddGroupCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-group [project_id]",
		Short: "Create a configuration group",
		Long: `Creates a new configuration group in the specified project.

A group is a container for related configurations (e.g., "Browsers",
"Operating Systems", "Devices"). After creating a group, you can add
individual configurations to it.`,
		Example: `  # Create a "Browsers" group
  gotr configurations add-group 1 --name="Browsers"

  # Preview before creating
  gotr configurations add-group 1 --name="OS" --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var projectID int64
			var err error
			if len(args) > 0 {
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr configurations add-group [project_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr configurations add-group [project_id]")
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

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations add-group")
				dr.PrintSimple("Create group", fmt.Sprintf("Project ID: %d, Name: %s", projectID, name))
				return nil
			}

			req := data.AddConfigGroupRequest{Name: name}
			resp, err := cli.AddConfigGroup(ctx, projectID, &req)
			if err != nil {
				return fmt.Errorf("failed to create group: %w", err)
			}

			ui.Successf(os.Stdout, "Group created (ID: %d)", resp.ID)
			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview what would be done without creating")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Group name (required)")

	return cmd
}
