package configurations

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newDeleteConfigCmd creates the 'configurations delete-config' command.
// Endpoint: POST /delete_config/{config_id}
func newDeleteConfigCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-config [config_id]",
		Short: "Delete a configuration",
		Long: `Deletes a configuration from a group.

⚠️ Warning: deletion cannot be undone! Make sure the configuration
is not used in active test plans.`,
		Example: `  # Delete a configuration
  gotr configurations delete-config 10

  # Preview before deleting
  gotr configurations delete-config 10 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var configID int64
			var err error
			if len(args) > 0 {
				configID, err = flags.ValidateRequiredID(args, 0, "config_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("config_id is required in non-interactive mode: gotr configurations delete-config [config_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("config_id is required in non-interactive mode: gotr configurations delete-config [config_id]")
				}

				configID, err = resolveConfigIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations delete-config")
				dr.PrintSimple("Delete configuration", fmt.Sprintf("Config ID: %d", configID))
				return nil
			}

			if err := cli.DeleteConfig(ctx, configID); err != nil {
				return fmt.Errorf("failed to delete configuration: %w", err)
			}

			ui.Successf(os.Stdout, "Configuration %d deleted", configID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview what would be deleted without actually deleting")

	return cmd
}
