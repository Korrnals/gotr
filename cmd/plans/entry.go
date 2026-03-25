package plans

import (
	"fmt"
	"os"
	"strings"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
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
		Use:   "add [plan_id]",
		Short: "Добавить запись в план",
		Long:  `Добавляет новую запись (тестовый прогон) в существующий план.`,
		Example: `  # Добавить прогон с названием
  gotr plans entry add 100 --suite-id=50 --name="Прогон 1"

  # Добавить с конфигурациями
  gotr plans entry add 100 --suite-id=50 --config-ids="1,2,3"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var planID int64
			if len(args) > 0 {
				var err error
				planID, err = flags.ValidateRequiredID(args, 0, "plan_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("plan_id is required in non-interactive mode: gotr plans entry add [plan_id]")
				}
				var err error
				planID, err = resolvePlanIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
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
				dr := output.NewDryRunPrinter("plans entry add")
				dr.PrintSimple("Add Plan Entry", fmt.Sprintf("Plan ID: %d, Suite ID: %d", planID, suiteID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.AddPlanEntry(ctx, planID, &req)
			if err != nil {
				return fmt.Errorf("failed to add plan entry: %w", err)
			}

			ui.Successf(os.Stdout, "Entry added to plan %d", planID)
			return output.OutputResult(cmd, resp, "plans")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без добавления")
	output.AddFlag(cmd)
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
			planID, err := flags.ValidateRequiredID(args, 0, "plan_id")
			if err != nil {
				return err
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
				dr := output.NewDryRunPrinter("plans entry update")
				dr.PrintSimple("Update Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s", planID, entryID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.UpdatePlanEntry(ctx, planID, entryID, &req)
			if err != nil {
				return fmt.Errorf("failed to update plan entry: %w", err)
			}

			ui.Successf(os.Stdout, "Entry %s updated in plan %d", entryID, planID)
			return output.OutputResult(cmd, resp, "plans")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)
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
			planID, err := flags.ValidateRequiredID(args, 0, "plan_id")
			if err != nil {
				return err
			}

			entryID := args[1]
			if entryID == "" {
				return fmt.Errorf("entry_id is required")
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("plans entry delete")
				dr.PrintSimple("Delete Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s", planID, entryID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeletePlanEntry(ctx, planID, entryID); err != nil {
				return fmt.Errorf("failed to delete plan entry: %w", err)
			}

			ui.Successf(os.Stdout, "Entry %s deleted from plan %d", entryID, planID)
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
		id, err := flags.ParseID(part)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}
