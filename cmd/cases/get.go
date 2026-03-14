package cases

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'cases get'
// Эндпоинт: GET /get_case/{case_id}
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
			caseID, err := flags.ValidateRequiredID(args, 0, "case_id")
			if err != nil {
				return err
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetCase(ctx, caseID)
			if err != nil {
				return fmt.Errorf("failed to get case: %w", err)
			}

			return output.OutputResult(cmd, resp, "cases")
		},
	}
}
