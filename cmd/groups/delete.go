package groups

import (
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// newDeleteCmd создаёт команду для удаления группы
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <group_id>",
		Short: "Удалить группу",
		Long:  `Удалить группу пользователей.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getClient(cmd)
			ctx := cmd.Context()

			groupID, err := flags.ValidateRequiredID(args, 0, "group_id")
			if err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				ui.Infof(os.Stdout, "[DRY-RUN] Will delete group %d", groupID)
				return nil
			}

			if err := client.DeleteGroup(ctx, groupID); err != nil {
				return err
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				color.New(color.FgGreen).Fprintf(cmd.OutOrStdout(), "✓ Group %d deleted\n", groupID)
			}

			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано, без выполнения")

	return cmd
}
