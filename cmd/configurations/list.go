package configurations

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'configurations list' command.
// Endpoint: GET /get_configs/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "List project configurations",
		Long: `Lists configurations available in the specified project.

Configurations represent test environments (browsers, OS, devices)
and are grouped by type. They are used when creating test plans
with multiple configurations.

Each configuration has an ID that is used to specify
parameters when creating plan entries with configurations.`,
		Example: `  # Get project configurations
  gotr configurations list 1

  # Save to file for analysis
  gotr configurations list 5 -o configs.json`,
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
					return fmt.Errorf("project_id is required in non-interactive mode: gotr configurations list [project_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr configurations list [project_id]")
				}

				projectID, err = resolveProjectIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetConfigs(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to get configurations: %w", err)
			}

			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
