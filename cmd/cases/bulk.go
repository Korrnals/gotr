package cases

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newBulkCmd создаёт родительскую команду 'cases bulk'
// Родительская команда для массовых операций над кейсами
func newBulkCmd(getClient GetClientFunc) *cobra.Command {
	bulkCmd := &cobra.Command{
		Use:   "bulk",
		Short: "Массовые операции над тест-кейсами",
		Long: `Массовые операции: обновление, удаление, копирование или перемещение нескольких тест-кейсов.

Подкоманды:
  • update — массовое обновление полей
  • delete — массовое удаление
  • copy   — копирование в другую секцию
  • move   — перемещение в другую секцию`,
	}

	bulkCmd.AddCommand(newBulkUpdateCmd(getClient))
	bulkCmd.AddCommand(newBulkDeleteCmd(getClient))
	bulkCmd.AddCommand(newBulkCopyCmd(getClient))
	bulkCmd.AddCommand(newBulkMoveCmd(getClient))

	return bulkCmd
}

// newBulkUpdateCmd создаёт команду 'cases bulk update'
// Эндпоинт: POST /update_cases/{suite_id}
func newBulkUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <case_ids...>",
		Short: "Массовое обновление кейсов",
		Long:  `Обновляет несколько тест-кейсов одновременно с одинаковыми значениями полей.`,
		Example: `  # Обновить приоритет нескольких кейсов
  gotr cases bulk update 1,2,3 --suite-id=100 --priority-id=1

  # Обновить оценку времени
  gotr cases bulk update 1 2 3 --suite-id=100 --estimate="1h 30m"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("case IDs required")
			}

			caseIDs := parseIDList(args)
			if len(caseIDs) == 0 {
				return fmt.Errorf("no valid case IDs provided")
			}

			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			if suiteID <= 0 {
				return fmt.Errorf("--suite-id is required")
			}

			req := data.UpdateCasesRequest{CaseIDs: caseIDs}
			if v, _ := cmd.Flags().GetInt64("priority-id"); v > 0 {
				req.PriorityID = v
			}
			if v, _ := cmd.Flags().GetString("estimate"); v != "" {
				req.Estimate = v
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases bulk update")
				dr.PrintSimple("Bulk Update Cases", fmt.Sprintf("Suite: %d, Cases: %v", suiteID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdateCases(suiteID, &req)
			if err != nil {
				return fmt.Errorf("failed to update cases: %w", err)
			}

			fmt.Printf("✅ Updated %d cases\n", len(caseIDs))
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	save.AddFlag(cmd)
	cmd.Flags().Int64("suite-id", 0, "ID сьюты (обязательно)")
	cmd.Flags().Int64("priority-id", 0, "ID приоритета для установки")
	cmd.Flags().String("estimate", "", "Оценка времени (например: '1h 30m')")

	return cmd
}

// newBulkDeleteCmd создаёт команду 'cases bulk delete'
// Эндпоинт: POST /delete_cases/{suite_id}
func newBulkDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <case_ids...>",
		Short: "Массовое удаление кейсов",
		Long:  `Удаляет несколько тест-кейсов одновременно.`,
		Example: `  # Удалить несколько кейсов
  gotr cases bulk delete 1,2,3 --suite-id=100

  # Проверить перед удалением
  gotr cases bulk delete 1,2,3 --suite-id=100 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("case IDs required")
			}

			caseIDs := parseIDList(args)
			if len(caseIDs) == 0 {
				return fmt.Errorf("no valid case IDs provided")
			}

			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			if suiteID <= 0 {
				return fmt.Errorf("--suite-id is required")
			}

			req := data.DeleteCasesRequest{CaseIDs: caseIDs}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases bulk delete")
				dr.PrintSimple("Bulk Delete Cases", fmt.Sprintf("Suite: %d, Cases: %v", suiteID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeleteCases(suiteID, &req); err != nil {
				return fmt.Errorf("failed to delete cases: %w", err)
			}

			fmt.Printf("✅ Deleted %d cases\n", len(caseIDs))
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено")
	cmd.Flags().Int64("suite-id", 0, "ID сьюты (обязательно)")

	return cmd
}

// newBulkCopyCmd создаёт команду 'cases bulk copy'
// Эндпоинт: POST /copy_cases_to_section/{section_id}
func newBulkCopyCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy <case_ids...>",
		Short: "Копировать кейсы в секцию",
		Long:  `Копирует несколько тест-кейсов в другую секцию.`,
		Example: `  # Копировать кейсы в другую секцию
  gotr cases bulk copy 1,2,3 --section-id=50

  # Проверить перед копированием
  gotr cases bulk copy 1,2,3 --section-id=50 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("case IDs required")
			}

			caseIDs := parseIDList(args)
			if len(caseIDs) == 0 {
				return fmt.Errorf("no valid case IDs provided")
			}

			sectionID, _ := cmd.Flags().GetInt64("section-id")
			if sectionID <= 0 {
				return fmt.Errorf("--section-id is required")
			}

			req := data.CopyCasesRequest{CaseIDs: caseIDs}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases bulk copy")
				dr.PrintSimple("Copy Cases", fmt.Sprintf("Section: %d, Cases: %v", sectionID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.CopyCasesToSection(sectionID, &req); err != nil {
				return fmt.Errorf("failed to copy cases: %w", err)
			}

			fmt.Printf("✅ Copied %d cases to section %d\n", len(caseIDs), sectionID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано")
	cmd.Flags().Int64("section-id", 0, "ID целевой секции (обязательно)")

	return cmd
}

// newBulkMoveCmd создаёт команду 'cases bulk move'
// Эндпоинт: POST /move_cases_to_section/{section_id}
func newBulkMoveCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <case_ids...>",
		Short: "Переместить кейсы в секцию",
		Long:  `Перемещает несколько тест-кейсов в другую секцию.`,
		Example: `  # Переместить кейсы в другую секцию
  gotr cases bulk move 1,2,3 --section-id=50

  # Проверить перед перемещением
  gotr cases bulk move 1,2,3 --section-id=50 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("case IDs required")
			}

			caseIDs := parseIDList(args)
			if len(caseIDs) == 0 {
				return fmt.Errorf("no valid case IDs provided")
			}

			sectionID, _ := cmd.Flags().GetInt64("section-id")
			if sectionID <= 0 {
				return fmt.Errorf("--section-id is required")
			}

			req := data.MoveCasesRequest{CaseIDs: caseIDs}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases bulk move")
				dr.PrintSimple("Move Cases", fmt.Sprintf("Section: %d, Cases: %v", sectionID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.MoveCasesToSection(sectionID, &req); err != nil {
				return fmt.Errorf("failed to move cases: %w", err)
			}

			fmt.Printf("✅ Moved %d cases to section %d\n", len(caseIDs), sectionID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано")
	cmd.Flags().Int64("section-id", 0, "ID целевой секции (обязательно)")

	return cmd
}

// parseIDList разбирает ID, разделённые запятыми или пробелами
func parseIDList(args []string) []int64 {
	var ids []int64
	for _, arg := range args {
		for _, part := range strings.Split(arg, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.ParseInt(part, 10, 64)
			if err == nil && id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// outputResult выводит результат в JSON или сохраняет в файл
func outputResult(cmd *cobra.Command, data interface{}) error {
	_, err := save.Output(cmd, data, "cases", "json")
	return err
}
