package bdds

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду 'bdds add'
// Эндпоинт: POST /add_bdd/{test_case_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <case_id>",
		Short: "Добавить BDD сценарий к тест-кейсу",
		Long: `Добавляет BDD сценарий в формате Gherkin к указанному тест-кейсу.

Сценарий должен быть в формате Given-When-Then (Дано-Когда-Тогда).
Можно передать содержимое через файл или напрямую.`,
		Example: `  # Добавить BDD из файла
  gotr bdds add 12345 --file=scenario.feature

  # Добавить BDD из stdin
  cat scenario.feature | gotr bdds add 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			caseID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || caseID <= 0 {
				return fmt.Errorf("некорректный case_id: %s", args[0])
			}

			// Читаем содержимое BDD
			content, err := readBDDContent(cmd)
			if err != nil {
				return err
			}
			if content == "" {
				return fmt.Errorf("BDD содержимое не может быть пустым (используйте --file или stdin)")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("bdds add")
				dr.PrintSimple("Добавить BDD", fmt.Sprintf("Case ID: %d", caseID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.AddBDD(caseID, content)
			if err != nil {
				return fmt.Errorf("не удалось добавить BDD: %w", err)
			}

			fmt.Printf("✅ BDD добавлен к кейсу %d\n", caseID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без добавления")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")
	cmd.Flags().String("file", "", "Путь к файлу с Gherkin сценарием")

	return cmd
}

// readBDDContent читает содержимое BDD из файла или stdin
func readBDDContent(cmd *cobra.Command) (string, error) {
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("не удалось прочитать файл: %w", err)
		}
		return string(data), nil
	}

	// TODO: Чтение из stdin если файл не указан
	// Пока возвращаем пустую строку, будет ошибка валидации
	return "", nil
}
