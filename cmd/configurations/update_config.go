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

// newUpdateConfigCmd creates the 'configurations update-config' command.
// Endpoint: POST /update_config/{config_id}
func newUpdateConfigCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-config [config_id]",
		Short: "Update a configuration",
		Long:  `Updates the name of an existing configuration.`,
		Example: `  # Change configuration name
  gotr configurations update-config 10 --name="Chrome 120"

  # Preview before updating
  gotr configurations update-config 10 --name="Chrome 120" --dry-run`,
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
					return fmt.Errorf("config_id is required in non-interactive mode: gotr configurations update-config [config_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("config_id is required in non-interactive mode: gotr configurations update-config [config_id]")
				}

				configID, err = resolveConfigIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations update-config")
				dr.PrintSimple("Update configuration", fmt.Sprintf("Config ID: %d, New Name: %s", configID, name))
				return nil
			}

			req := data.UpdateConfigRequest{Name: name}
			resp, err := cli.UpdateConfig(ctx, configID, &req)
			if err != nil {
				return fmt.Errorf("failed to update configuration: %w", err)
			}

			ui.Successf(os.Stdout, "Configuration %d updated", configID)
			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview what would be done without applying changes")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "New configuration name (required)")

	return cmd
}
