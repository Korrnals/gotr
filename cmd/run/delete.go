package run

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [run-id]",
	Short: "Удалить test run",
	Long: `Удаляет test run по его ID.

⚠️ ВНИМАНИЕ: Это действие необратимо!

При удалении run:
- Все результаты тестов будут удалены
- Все тесты (tests) будут удалены
- Сама структура run будет удалена
- Кейсы в сьюте останутся нетронутыми

Рекомендуется сначала закрыть run (gotr run close), а не удалять.

Примеры:
	# Удалить run (без подтверждения — осторожно!)
	gotr run delete 12345

	# Удалить в тихом режиме (для скриптов)
	gotr run delete 12345 -q
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		httpClient := getClientSafe(cmd)
		if httpClient == nil {
			return fmt.Errorf("HTTP клиент не инициализирован")
		}

		svc := service.NewRunService(httpClient)
		runID, err := svc.ParseID(args, 0)
		if err != nil {
			return fmt.Errorf("некорректный ID test run: %w", err)
		}

		if err := svc.Delete(runID); err != nil {
			return fmt.Errorf("ошибка удаления test run: %w", err)
		}

		svc.PrintSuccess(cmd, "Test run %d удалён успешно", runID)
		return nil
	},
}
