package groups

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду для обновления группы
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <group_id>",
		Short: "Обновить группу",
		Long:  `Обновить существующую группу пользователей.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getClient(cmd)
			ctx := cmd.Context()

			groupID, err := flags.ValidateRequiredID(args, 0, "group_id")
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				ui.Infof(os.Stdout, "[DRY-RUN] Will update group %d, new name: '%s'", groupID, name)
				return nil
			}

			group, err := client.UpdateGroup(ctx, groupID, name, nil)
			if err != nil {
				return err
			}

			return output.OutputResult(cmd, group, "groups")
		},
	}

	cmd.Flags().StringP("name", "n", "", "Новое название группы (обязательно)")
	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано, без выполнения")
	output.AddFlag(cmd)

	return cmd
}
