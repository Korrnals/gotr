package groups

import (
	"fmt"
	"strconv"

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

			groupID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || groupID <= 0 {
				return fmt.Errorf("group_id должен быть положительным числом")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				fmt.Printf("[DRY-RUN] Будет удалена группа %d\n", groupID)
				return nil
			}

			if err := client.DeleteGroup(groupID); err != nil {
				return err
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				color.New(color.FgGreen).Fprintf(cmd.OutOrStdout(), "✓ Группа %d удалена\n", groupID)
			}

			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано, без выполнения")

	return cmd
}
