package result

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newFieldsCmd создаёт команду 'result fields'
func newFieldsCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fields",
		Short: "Получить список полей результатов",
		Long: `Получает список доступных полей для результатов тестов.

Эта команда полезна для понимания структуры данных результатов
и доступных полей для заполнения.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			// Проверяем dry-run режим
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := dryrun.New("result fields")
				dr.PrintOperation(
					"Get Result Fields",
					"GET",
					"/index.php?/api/v2/get_result_fields",
					nil,
				)
				return nil
			}

			fields, err := cli.GetResultFields()
			if err != nil {
				return fmt.Errorf("ошибка получения полей результатов: %w", err)
			}

			return outputResult(cmd, fields)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")

	return cmd
}

// outputResult выводит результат в JSON
func outputResult(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output != "" {
		return saveToFile(data, output)
	}
	return printJSON(data)
}

// saveToFile сохраняет данные в файл
func saveToFile(data interface{}, filename string) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации JSON: %w", err)
	}
	return os.WriteFile(filename, jsonBytes, 0644)
}

// printJSON выводит данные в JSON в stdout
func printJSON(data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

// fieldsCmd — экспортированная команда
var fieldsCmd = newFieldsCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClientSafe(cmd)
})
