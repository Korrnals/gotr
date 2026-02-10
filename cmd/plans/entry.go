package plans

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newEntryCmd создаёт родительскую команду 'plans entry'
// Родительская команда для управления записями плана
func newEntryCmd(getClient GetClientFunc) *cobra.Command {
	entryCmd := &cobra.Command{
		Use:   "entry",
		Short: "Управление записями плана",
		Long: `Управление записями (entries) тест-плана — тестовыми прогонами внутри плана.

Подкоманды:
  • add    — добавить запись (тестовый прогон) в план
  • update — обновить существующую запись
  • delete — удалить запись из плана`,
	}

	entryCmd.AddCommand(newEntryAddCmd(getClient))
	entryCmd.AddCommand(newEntryUpdateCmd(getClient))
	entryCmd.AddCommand(newEntryDeleteCmd(getClient))

	return entryCmd
}

// newEntryAddCmd создаёт команду 'plans entry add'
// Эндпоинт: POST /add_plan_entry/{plan_id}
func newEntryAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <plan_id>",
		Short: "Добавить запись в план",
		Long:  `Добавляет новую запись (тестовый прогон) в существующий план.`,
		Example: `  # Добавить прогон с названием
  gotr plans entry add 100 --suite-id=50 --name="Прогон 1"

  # Добавить с конфигурациями
  gotr plans entry add 100 --suite-id=50 --config-ids="1,2,3"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			if suiteID <= 0 {
				return fmt.Errorf("--suite-id is required")
			}

			req := data.AddPlanEntryRequest{
				SuiteID: suiteID,
			}

			if v, _ := cmd.Flags().GetString("name"); v != "" {
				req.Name = v
			}
			if v, _ := cmd.Flags().GetString("config-ids"); v != "" {
				req.ConfigIDs = parseIntList(v)
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans entry add")
				dr.PrintSimple("Add Plan Entry", fmt.Sprintf("Plan ID: %d, Suite ID: %d", planID, suiteID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.AddPlanEntry(planID, &req)
			if err != nil {
				return fmt.Errorf("failed to add plan entry: %w", err)
			}

			fmt.Printf("✅ Entry added to plan %d\n", planID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без добавления")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")
	cmd.Flags().Int64("suite-id", 0, "ID сьюты (обязательно)")
	cmd.Flags().String("name", "", "Название записи")
	cmd.Flags().String("config-ids", "", "ID конфигураций через запятую")

	return cmd
}

// newEntryUpdateCmd создаёт команду 'plans entry update'
// Эндпоинт: POST /update_plan_entry/{plan_id}/{entry_id}
func newEntryUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <plan_id> <entry_id>",
		Short: "Обновить запись плана",
		Long:  `Обновляет существующую запись в тест-плане.`,
		Example: `  # Изменить название записи
  gotr plans entry update 100 abc123 --name="Обновлённая запись"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			entryID := args[1]
			if entryID == "" {
				return fmt.Errorf("entry_id is required")
			}

			req := data.UpdatePlanEntryRequest{}

			if v, _ := cmd.Flags().GetString("name"); v != "" {
				req.Name = v
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans entry update")
				dr.PrintSimple("Update Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s", planID, entryID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdatePlanEntry(planID, entryID, &req)
			if err != nil {
				return fmt.Errorf("failed to update plan entry: %w", err)
			}

			fmt.Printf("✅ Entry %s updated in plan %d\n", entryID, planID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")
	cmd.Flags().String("name", "", "Новое название записи")

	return cmd
}

// newEntryDeleteCmd создаёт команду 'plans entry delete'
// Эндпоинт: POST /delete_plan_entry/{plan_id}/{entry_id}
func newEntryDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <plan_id> <entry_id>",
		Short: "Удалить запись плана",
		Long:  `Удаляет запись из тест-плана.`,
		Example: `  # Удалить запись из плана
  gotr plans entry delete 100 abc123

  # Проверить перед удалением
  gotr plans entry delete 100 abc123 --dry-run`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			entryID := args[1]
			if entryID == "" {
				return fmt.Errorf("entry_id is required")
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans entry delete")
				dr.PrintSimple("Delete Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s", planID, entryID))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeletePlanEntry(planID, entryID); err != nil {
				return fmt.Errorf("failed to delete plan entry: %w", err)
			}

			fmt.Printf("✅ Entry %s deleted from plan %d\n", entryID, planID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено")

	return cmd
}

// parseIntList разбирает список чисел, разделённых запятыми
func parseIntList(s string) []int64 {
	var ids []int64
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}
