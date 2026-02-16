package groups

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/internal/output"
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

			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("project_id должен быть положительным числом")
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name обязателен")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				fmt.Printf("[DRY-RUN] Будет создана группа '%s' в проекте %d\n", name, projectID)
				return nil
			}

			group, err := client.AddGroup(projectID, name, nil)
			if err != nil {
				return err
			}

			return output.Result(cmd, group)
		},
	}

	cmd.Flags().StringP("name", "n", "", "Название группы (обязательно)")
	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано, без выполнения")

	return cmd
}
