package run

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду 'run update'
func newUpdateCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [run-id]",
		Short: "Обновить test run",
		Long: `Обновляет существующий test run.

Можно обновлять только открытые runs. Для обновления используйте флаги.
Только изменённые поля будут отправлены в API.

Примеры:
	# Изменить название и описание
	gotr run update 12345 --name "Updated Name" --description "New description"

	# Переназначить на другого пользователя
	gotr run update 12345 --assigned-to 10

	# Изменить набор кейсов в run
	gotr run update 12345 --case-ids 100,200,300 --include-all=false

	# Dry-run режим
	gotr run update 12345 --name "Test" --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			svc := newRunServiceFromInterface(cli)
			runID, err := svc.ParseID(args, 0)
			if err != nil {
				return fmt.Errorf("некорректный ID test run: %w", err)
			}

			// Собираем параметры из флагов (только изменённые)
			req := &data.UpdateRunRequest{}

			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = &name
			}
			if cmd.Flags().Changed("description") {
				description, _ := cmd.Flags().GetString("description")
				req.Description = &description
			}
			if cmd.Flags().Changed("milestone-id") {
				milestoneID, _ := cmd.Flags().GetInt64("milestone-id")
				req.MilestoneID = &milestoneID
			}
			if cmd.Flags().Changed("assigned-to") {
				assignedTo, _ := cmd.Flags().GetInt64("assigned-to")
				req.AssignedTo = &assignedTo
			}
			if cmd.Flags().Changed("case-ids") {
				caseIDs, _ := cmd.Flags().GetInt64Slice("case-ids")
				req.CaseIDs = caseIDs
			}
			if cmd.Flags().Changed("include-all") {
				includeAll, _ := cmd.Flags().GetBool("include-all")
				req.IncludeAll = &includeAll
			}

			// Проверяем dry-run режим
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("run update")
				dr.PrintOperation(
					fmt.Sprintf("Update Run %d", runID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/update_run/%d", runID),
					req,
				)
				return nil
			}

			run, err := svc.Update(runID, req)
			if err != nil {
				return fmt.Errorf("ошибка обновления test run: %w", err)
			}

			svc.PrintSuccess(cmd, "Test run обновлён успешно:")
			return svc.Output(cmd, run)
		},
	}

	cmd.Flags().String("name", "", "Новое название")
	cmd.Flags().String("description", "", "Новое описание")
	cmd.Flags().Int64("milestone-id", 0, "ID milestone")
	cmd.Flags().Int64("assigned-to", 0, "ID пользователя для назначения")
	cmd.Flags().Int64Slice("case-ids", nil, "Список ID кейсов (через запятую)")
	cmd.Flags().Bool("include-all", false, "Включить все кейсы сьюты")
	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")

	return cmd
}

// updateCmd — экспортированная команда
var updateCmd = newUpdateCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClientSafe(cmd)
})
