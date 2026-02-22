package bdds

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'bdds get'
// Эндпоинт: GET /get_bdd/{test_case_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <case_id>",
		Short: "Получить BDD сценарий для тест-кейса",
		Long: `Получает BDD сценарий, привязанный к указанному тест-кейсу.

Возвращает Gherkin сценарий в формате Given-When-Then,
если он был добавлен к тест-кейсу.`,
		Example: `  # Получить BDD для кейса
  gotr bdds get 12345

  # Сохранить в файл
  gotr bdds get 12345 -o bdd.feature`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			caseID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || caseID <= 0 {
				return fmt.Errorf("некорректный case_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetBDD(caseID)
			if err != nil {
				return fmt.Errorf("не удалось получить BDD: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	output.AddFlag(cmd)

	return cmd
}

func outputResult(cmd *cobra.Command, data interface{}) error {
	_, err := output.Output(cmd, data, "bdds", "json")
	return err
}
