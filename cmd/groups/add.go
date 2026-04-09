package groups

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddCmd creates the 'groups add' command.
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [project_id]",
		Short: "Create a new group",
		Long:  `Create a new user group in the specified project.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getClient(cmd)
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
					return fmt.Errorf("project_id required: gotr groups add [project_id] --name <name>")
				}
				projectID, err = resolveProjectIDInteractive(ctx, client)
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				ui.Infof(os.Stdout, "[DRY-RUN] Will create group '%s' in project %d", name, projectID)
				return nil
			}

			group, err := client.AddGroup(ctx, projectID, name, nil)
			if err != nil {
				return err
			}

			return output.OutputResult(cmd, group, "groups")
		},
	}

	cmd.Flags().StringP("name", "n", "", "Group name (required)")
	cmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	output.AddFlag(cmd)

	return cmd
}
