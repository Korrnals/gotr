package groups

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'groups list'
// Эндпоинт: GET /get_groups/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
		Short: "Список групп проекта",
		Long: `Выводит список групп пользователей, доступных в указанном проекте.

Каждая группа содержит ID, название и информацию о пользователях,
входящих в группу. Используется для просмотра структуры команд
и управления правами доступа в рамках проекта.`,
		Example: `  # Получить список групп проекта
  gotr groups list 1

  # Сохранить в файл
  gotr groups list 5 -o groups.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("некорректный project_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetGroups(projectID)
			if err != nil {
				return fmt.Errorf("не удалось получить список групп: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")

	return cmd
}

// outputResult выводит результат в JSON или сохраняет в файл
func outputResult(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if output != "" {
		return os.WriteFile(output, jsonBytes, 0644)
	}

	fmt.Println(string(jsonBytes))
	return nil
}
