package run

import (
	"fmt"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [project-id]",
	Short: "Создать новый test run",
	Long: `Создаёт новый test run в указанном проекте.

Test run создаётся на основе тест-сюиты (suite). Можно указать:
- название и описание
- milestone для привязки
- пользователя для назначения (assignedto_id)
- конкретные case_ids (если не нужны все кейсы сьюты)
- config_ids для конфигурационного тестирования

Примеры:
	# Создать run с минимальными параметрами
	gotr run create 30 --suite-id 20069 --name "Smoke Tests"

	# Создать run с описанием и назначением
	gotr run create 30 --suite-id 20069 --name "Regression" \\
		--description "Full regression suite" --assigned-to 5

	# Создать run только с определёнными кейсами
	gotr run create 30 --suite-id 20069 --name "Critical Path" \\
		--case-ids 123,456,789

	# Dry-run режим
	gotr run create 30 --suite-id 20069 --name "Test" --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		httpClient := getClientSafe(cmd)
		if httpClient == nil {
			return fmt.Errorf("HTTP клиент не инициализирован")
		}

		svc := service.NewRunService(httpClient)
		projectID, err := svc.ParseID(args, 0)
		if err != nil {
			return fmt.Errorf("некорректный ID проекта: %w", err)
		}

		// Собираем параметры из флагов
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		suiteID, _ := cmd.Flags().GetInt64("suite-id")
		milestoneID, _ := cmd.Flags().GetInt64("milestone-id")
		assignedTo, _ := cmd.Flags().GetInt64("assigned-to")
		caseIDs, _ := cmd.Flags().GetInt64Slice("case-ids")
		configIDs, _ := cmd.Flags().GetInt64Slice("config-ids")
		includeAll, _ := cmd.Flags().GetBool("include-all")

		req := &data.AddRunRequest{
			Name:        name,
			Description: description,
			SuiteID:     suiteID,
			MilestoneID: milestoneID,
			AssignedTo:  assignedTo,
			CaseIDs:     caseIDs,
			ConfigIDs:   configIDs,
			IncludeAll:  includeAll,
		}

		// Проверяем dry-run режим
		isDryRun, _ := cmd.Flags().GetBool("dry-run")
		if isDryRun {
			dr := dryrun.New("run create")
			dr.PrintOperation(
				fmt.Sprintf("Create Run in Project %d", projectID),
				"POST",
				fmt.Sprintf("/index.php?/api/v2/add_run/%d", projectID),
				req,
			)
			return nil
		}

		run, err := svc.Create(projectID, req)
		if err != nil {
			return fmt.Errorf("ошибка создания test run: %w", err)
		}

		svc.PrintSuccess(cmd, "Test run создан успешно (ID: %d):", run.ID)
		return svc.Output(cmd, run)
	},
}
