package cases

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'cases get'
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "get <case_id>",
		Short: "Получить тест-кейс по ID",
		Long:  `Получает детальную информацию о тест-кейсе по его идентификатору.`,
		Example: `  # Получить информацию о кейсе
  gotr cases get 12345

  # Сохранить в файл
  gotr cases get 12345 -o case.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			caseID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || caseID <= 0 {
				return fmt.Errorf("invalid case_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetCase(caseID)
			if err != nil {
				return fmt.Errorf("failed to get case: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}
}
