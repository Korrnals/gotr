package groups

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/internal/output"
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

			groupID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || groupID <= 0 {
				return fmt.Errorf("group_id должен быть положительным числом")
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name обязателен")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				fmt.Printf("[DRY-RUN] Будет обновлена группа %d, новое название: '%s'\n", groupID, name)
				return nil
			}

			group, err := client.UpdateGroup(groupID, name, nil)
			if err != nil {
				return err
			}

			return output.Result(cmd, group)
		},
	}

	cmd.Flags().StringP("name", "n", "", "Новое название группы (обязательно)")
	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано, без выполнения")

	return cmd
}
