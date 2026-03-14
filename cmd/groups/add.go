package groups

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду для добавления группы
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <project_id>",
		Short: "Создать новую группу",
		Long:  `Создать новую группу пользователей в указанном проекте.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getClient(cmd)
			ctx := cmd.Context()

			projectID, err := flags.ValidateRequiredID(args, 0, "project_id")
			if err != nil {
				return err
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

	cmd.Flags().StringP("name", "n", "", "Название группы (обязательно)")
	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано, без выполнения")
	output.AddFlag(cmd)

	return cmd
}
