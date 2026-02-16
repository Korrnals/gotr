package attachments

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// newAddCaseCmd создаёт команду 'attachments add case'
// Эндпоинт: POST /add_attachment_to_case/{case_id}
func newAddCaseCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "case <case_id> <file_path>",
		Short: "Добавить вложение к тест-кейсу",
		Long:  `Загружает файл и прикрепляет его к указанному тест-кейсу.`,
		Example: `  # Прикрепить скриншот к тест-кейсу
  gotr attachments add case 12345 ./screenshot.png

  # Проверить без реальной загрузки
  gotr attachments add case 99999 ./test-data.json --dry-run`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			caseID, err := parseID(args[0], "case_id")
			if err != nil {
				return err
			}
			filePath := args[1]

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("attachments add case")
				dr.PrintSimple("Add Attachment to Case", fmt.Sprintf("Case ID: %d, File: %s", caseID, filePath))
				return nil
			}

			// Validate file exists
			if err := validateFileExists(filePath); err != nil {
				return err
			}

			cli := getClient(cmd)
			resp, err := cli.AddAttachmentToCase(caseID, filePath)
			if err != nil {
				return fmt.Errorf("failed to add attachment: %w", err)
			}

			fmt.Printf("✅ Attachment added (ID: %d)\n   URL: %s\n", resp.AttachmentID, resp.URL)
			return outputResult(cmd, resp)
		},
	}
	save.AddFlag(cmd)
	return cmd
}

// newAddPlanCmd создаёт команду 'attachments add plan'
// Эндпоинт: POST /add_attachment_to_plan/{plan_id}
func newAddPlanCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan <plan_id> <file_path>",
		Short: "Добавить вложение к тест-плану",
		Long:  `Загружает файл и прикрепляет его к указанному тест-плану.`,
		Example: `  # Прикрепить отчёт к плану
  gotr attachments add plan 100 ./report.pdf

  # Прикрепить документ
  gotr attachments add plan 200 ./summary.docx`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := parseID(args[0], "plan_id")
			if err != nil {
				return err
			}
			filePath := args[1]

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("attachments add plan")
				dr.PrintSimple("Add Attachment to Plan", fmt.Sprintf("Plan ID: %d, File: %s", planID, filePath))
				return nil
			}

			if err := validateFileExists(filePath); err != nil {
				return err
			}

			cli := getClient(cmd)
			resp, err := cli.AddAttachmentToPlan(planID, filePath)
			if err != nil {
				return fmt.Errorf("failed to add attachment: %w", err)
			}

			fmt.Printf("✅ Attachment added (ID: %d)\n   URL: %s\n", resp.AttachmentID, resp.URL)
			return outputResult(cmd, resp)
		},
	}
	save.AddFlag(cmd)
	return cmd
}

// newAddPlanEntryCmd создаёт команду 'attachments add plan-entry'
// Эндпоинт: POST /add_attachment_to_plan_entry/{plan_id}/{entry_id}
func newAddPlanEntryCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan-entry <plan_id> <entry_id> <file_path>",
		Short: "Добавить вложение к записи плана",
		Long:  `Загружает файл и прикрепляет его к записи (entry) в тест-плане.`,
		Example: `  # Прикрепить данные к записи плана
  gotr attachments add plan-entry 100 entry-abc123 ./data.csv

  # Прикрепить заметки
  gotr attachments add plan-entry 200 def456 ./notes.txt`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := parseID(args[0], "plan_id")
			if err != nil {
				return err
			}
			entryID := args[1]
			filePath := args[2]

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("attachments add plan-entry")
				dr.PrintSimple("Add Attachment to Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s, File: %s", planID, entryID, filePath))
				return nil
			}

			if err := validateFileExists(filePath); err != nil {
				return err
			}

			cli := getClient(cmd)
			resp, err := cli.AddAttachmentToPlanEntry(planID, entryID, filePath)
			if err != nil {
				return fmt.Errorf("failed to add attachment: %w", err)
			}

			fmt.Printf("✅ Attachment added (ID: %d)\n   URL: %s\n", resp.AttachmentID, resp.URL)
			return outputResult(cmd, resp)
		},
	}
	save.AddFlag(cmd)
	return cmd
}

// newAddResultCmd создаёт команду 'attachments add result'
// Эндпоинт: POST /add_attachment_to_result/{result_id}
func newAddResultCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "result <result_id> <file_path>",
		Short: "Добавить вложение к результату теста",
		Long:  `Загружает файл и прикрепляет его к результату выполнения теста.`,
		Example: `  # Прикрепить лог к результату
  gotr attachments add result 98765 ./log.txt

  # Прикрепить скриншот ошибки
  gotr attachments add result 54321 ./screenshot.png`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resultID, err := parseID(args[0], "result_id")
			if err != nil {
				return err
			}
			filePath := args[1]

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("attachments add result")
				dr.PrintSimple("Add Attachment to Result", fmt.Sprintf("Result ID: %d, File: %s", resultID, filePath))
				return nil
			}

			if err := validateFileExists(filePath); err != nil {
				return err
			}

			cli := getClient(cmd)
			resp, err := cli.AddAttachmentToResult(resultID, filePath)
			if err != nil {
				return fmt.Errorf("failed to add attachment: %w", err)
			}

			fmt.Printf("✅ Attachment added (ID: %d)\n   URL: %s\n", resp.AttachmentID, resp.URL)
			return outputResult(cmd, resp)
		},
	}
	save.AddFlag(cmd)
	return cmd
}

// newAddRunCmd создаёт команду 'attachments add run'
// Эндпоинт: POST /add_attachment_to_run/{run_id}
func newAddRunCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <run_id> <file_path>",
		Short: "Добавить вложение к тестовому прогону",
		Long:  `Загружает файл и прикрепляет его к тестовому прогону.`,
		Example: `  # Прикрепить HTML-отчёт к прогону
  gotr attachments add run 555 ./report.html

  # Прикрепить PDF-сводку
  gotr attachments add run 777 ./summary.pdf`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID, err := parseID(args[0], "run_id")
			if err != nil {
				return err
			}
			filePath := args[1]

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("attachments add run")
				dr.PrintSimple("Add Attachment to Run", fmt.Sprintf("Run ID: %d, File: %s", runID, filePath))
				return nil
			}

			if err := validateFileExists(filePath); err != nil {
				return err
			}

			cli := getClient(cmd)
			resp, err := cli.AddAttachmentToRun(runID, filePath)
			if err != nil {
				return fmt.Errorf("failed to add attachment: %w", err)
			}

			fmt.Printf("✅ Attachment added (ID: %d)\n   URL: %s\n", resp.AttachmentID, resp.URL)
			return outputResult(cmd, resp)
		},
	}
	save.AddFlag(cmd)
	return cmd
}

// parseID преобразует строковый ID в int64
func parseID(s, name string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s: %s", name, s)
	}
	return id, nil
}

// validateFileExists проверяет существование файла
func validateFileExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}
	return nil
}

// outputResult выводит результат в JSON или сохраняет в файл
func outputResult(cmd *cobra.Command, data interface{}) error {
	_, err := save.Output(cmd, data, "attachments", "json")
	return err
}
