package test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду для получения информации о тесте
func newGetCmd(getClient func(cmd *cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [test-id]",
		Short: "Получить информацию о тесте",
		Long: `Получает детальную информацию о тесте по его ID.

Примеры:
	# Получить тест по ID
	gotr test get 12345

	# Получить и сохранить в файл
	gotr test get 12345 -o test.json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			httpClient := getClient(cmd)
			if httpClient == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			svc := service.NewTestService(httpClient)
			testID, err := svc.ParseID(args, 0)
			if err != nil {
				return fmt.Errorf("некорректный ID теста: %w", err)
			}

			test, err := svc.Get(testID)
			if err != nil {
				return fmt.Errorf("ошибка получения теста: %w", err)
			}

			// Проверяем нужно ли сохранить в файл
			output, _ := cmd.Flags().GetString("output")
			if output != "" {
				jsonBytes, err := json.MarshalIndent(test, "", "  ")
				if err != nil {
					return fmt.Errorf("ошибка сериализации: %w", err)
				}
				if err := os.WriteFile(output, jsonBytes, 0644); err != nil {
					return fmt.Errorf("ошибка записи файла: %w", err)
				}
				svc.PrintSuccess(cmd, "Тест сохранён в %s", output)
				return nil
			}

			return svc.Output(cmd, test)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл")
	cmd.Flags().BoolP("quiet", "q", false, "Тихий режим")

	return cmd
}
