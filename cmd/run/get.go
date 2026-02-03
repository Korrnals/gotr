package run

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [run-id]",
	Short: "Получить информацию о test run",
	Long: `Получает детальную информацию о test run по его ID.

Test run — это экземпляр тест-сюиты, запущенный для выполнения тестов.
В ответе содержится: название, описание, статистика прохождения,
даты создания/обновления, assignedto_id и другие поля.

Примеры:
	# Получить информацию о run
	gotr run get 12345

	# Сохранить результат в файл
	gotr run get 12345 -o run_info.json
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

		run, err := svc.Get(runID)
		if err != nil {
			return fmt.Errorf("ошибка получения test run: %w", err)
		}

		return svc.Output(cmd, run)
	},
}
