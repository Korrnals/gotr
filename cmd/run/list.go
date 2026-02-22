package run

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'run list'
func newListCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project-id]",
		Short: "Получить список test runs проекта",
		Long: `Получает список всех test runs для указанного проекта.

В списке содержатся активные и завершённые runs с базовой информацией:
ID, название, описание, статистика тестов (passed/failed/blocked).

Если project-id не указан, будет предложен интерактивный выбор из списка проектов.

Примеры:
	# Получить список runs проекта (с интерактивным выбором)
	gotr run list

	# Получить список runs проекта (с явным ID)
	gotr run list 30

	# Сохранить в файл для дальнейшей обработки
	gotr run list 30 -o runs.json

	# Dry-run режим
	gotr run list 30 --dry-run
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			svc := newRunServiceFromInterface(cli)

			var projectID int64
			var err error

			if len(args) > 0 {
				// Явно указан project-id
				projectID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("некорректный ID проекта: %w", err)
				}
			} else {
				// Интерактивный выбор проекта
				// Нужен *client.HTTPClient для интерактивного режима
				httpClient, ok := cli.(*client.HTTPClient)
				if !ok {
					return fmt.Errorf("интерактивный режим недоступен в тестовом режиме, укажите project-id")
				}
				projectID, err = interactive.SelectProjectInteractively(httpClient)
				if err != nil {
					return err
				}
			}

			// Проверяем dry-run режим
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("run list")
				dr.PrintOperation(
					fmt.Sprintf("List Runs for Project %d", projectID),
					"GET",
					fmt.Sprintf("/index.php?/api/v2/get_runs/%d", projectID),
					nil,
				)
				return nil
			}

			runs, err := svc.GetByProject(projectID)
			if err != nil {
				return fmt.Errorf("ошибка получения списка test runs: %w", err)
			}

			return svc.Output(cmd, runs)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")

	return cmd
}

// listCmd — экспортированная команда
var listCmd = newListCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClientSafe(cmd)
})
