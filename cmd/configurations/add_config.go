package configurations

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddConfigCmd creates the 'configurations add-config' command.
// Endpoint: POST /add_config/{group_id}
func newAddConfigCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-config [group_id]",
		Short: "Add a configuration to a group",
		Long: `Adds a new configuration to an existing group.

A configuration is a specific value (e.g., "Chrome", "Windows 10",
"iPhone 12") within a group. Configurations are used when creating
test plans with multiple configurations.`,
		Example: `  # Add "Chrome" to group 5
  gotr configurations add-config 5 --name="Chrome"

  # Preview before adding
  gotr configurations add-config 5 --name="Firefox" --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var groupID int64
			var err error
			if len(args) > 0 {
				groupID, err = flags.ValidateRequiredID(args, 0, "group_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations add-config [group_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations add-config [group_id]")
				}

				groupID, err = resolveGroupIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations add-config")
				dr.PrintSimple("Add configuration", fmt.Sprintf("Group ID: %d, Name: %s", groupID, name))
				return nil
			}

			req := data.AddConfigRequest{Name: name}
			quiet, _ := cmd.Flags().GetBool("quiet")
			resp, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Creating configuration",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*data.Config, error) {
				return cli.AddConfig(ctx, groupID, &req)
			})
			if err != nil {
				return fmt.Errorf("failed to add configuration: %w", err)
			}

			ui.Successf(os.Stdout, "Configuration added (ID: %d)", resp.ID)
			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview what would be done without adding")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Configuration name (required)")

	return cmd
}
