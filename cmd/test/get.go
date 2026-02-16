package test

import (
	"fmt"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
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
			saveFlag, _ := cmd.Flags().GetBool("save")
			if saveFlag {
				filepath, err := save.Output(cmd, test, "test", "json")
				if err != nil {
					return fmt.Errorf("ошибка сохранения: %w", err)
				}
				if filepath != "" {
					svc.PrintSuccess(cmd, "Тест сохранён в %s", filepath)
				}
				return nil
			}

			return svc.Output(cmd, test)
		},
	}

	save.AddFlag(cmd)
	cmd.Flags().BoolP("quiet", "q", false, "Тихий режим")

	return cmd
}
