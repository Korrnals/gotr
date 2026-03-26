package cases

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'cases get'
// Эндпоинт: GET /get_case/{case_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "get [case_id]",
		Short: "Получить тест-кейс по ID",
		Long:  `Получает детальную информацию о тест-кейсе по его идентификатору.`,
		Example: `  # Получить информацию о кейсе
  gotr cases get 12345

  # Сохранить в файл
  gotr cases get 12345 -o case.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var caseID int64
			var err error
			if len(args) > 0 {
				caseID, err = flags.ValidateRequiredID(args, 0, "case_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("case_id required: gotr cases get [case_id]")
				}

				caseID, err = resolveCaseIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetCase(ctx, caseID)
			if err != nil {
				return fmt.Errorf("failed to get case: %w", err)
			}

			return output.OutputResult(cmd, resp, "cases")
		},
	}
}
