package attachments

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

type attachmentUploadFunc func(context.Context) (*data.AttachmentResponse, error)

func runAttachmentUpload(cmd *cobra.Command, upload attachmentUploadFunc) (*data.AttachmentResponse, error) {
	return ui.RunWithStatus(cmd.Context(), ui.StatusConfig{
		Title:  "Uploading attachment...",
		Writer: os.Stderr,
	}, upload)
}

// newAddCaseCmd creates the 'attachments add case' command.
// Endpoint: POST /add_attachment_to_case/{case_id}
func newAddCaseCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "case [case_id] <file_path>",
		Short: "Добавить вложение к тест-кейсу",
		Long:  `Загружает файл и прикрепляет его к указанному тест-кейсу.`,
		Example: `  # Прикрепить скриншот к тест-кейсу
  gotr attachments add case 12345 ./screenshot.png

  # Проверить без реальной загрузки
  gotr attachments add case 99999 ./test-data.json --dry-run`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cli := getClient(cmd)

			var caseID int64
			var filePath string
			var err error
			if len(args) == 2 {
				caseID, err = flags.ValidateRequiredID(args, 0, "case_id")
				if err != nil {
					return err
				}
				filePath = args[1]
			} else {
				filePath = args[0]
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("case_id required: gotr attachments add case [case_id] <file_path>")
				}
				caseID, err = resolveCaseIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("attachments add case")
				dr.PrintSimple("Add Attachment to Case", fmt.Sprintf("Case ID: %d, File: %s", caseID, filePath))
				return nil
			}

			// Validate file exists
			if err := validateFileExists(filePath); err != nil {
				return err
			}

			resp, err := runAttachmentUpload(cmd, func(ctx context.Context) (*data.AttachmentResponse, error) {
				return cli.AddAttachmentToCase(ctx, caseID, filePath)
			})
			if err != nil {
				return fmt.Errorf("failed to add attachment: %w", err)
			}

			ui.Successf(os.Stdout, "Attachment added (ID: %d)\n   URL: %s", resp.AttachmentID, resp.URL)
			return output.OutputResult(cmd, resp, "attachments")
		},
	}
	output.AddFlag(cmd)
	return cmd
}

// newAddPlanCmd creates the 'attachments add plan' command.
// Endpoint: POST /add_attachment_to_plan/{plan_id}
func newAddPlanCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [plan_id] <file_path>",
		Short: "Добавить вложение к тест-плану",
		Long:  `Загружает файл и прикрепляет его к указанному тест-плану.`,
		Example: `  # Прикрепить отчёт к плану
  gotr attachments add plan 100 ./report.pdf

  # Прикрепить документ
  gotr attachments add plan 200 ./summary.docx`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cli := getClient(cmd)

			var planID int64
			var filePath string
			var err error
			if len(args) == 2 {
				planID, err = flags.ValidateRequiredID(args, 0, "plan_id")
				if err != nil {
					return err
				}
				filePath = args[1]
			} else {
				filePath = args[0]
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("plan_id required: gotr attachments add plan [plan_id] <file_path>")
				}
				planID, err = resolvePlanIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("attachments add plan")
				dr.PrintSimple("Add Attachment to Plan", fmt.Sprintf("Plan ID: %d, File: %s", planID, filePath))
				return nil
			}

			if err := validateFileExists(filePath); err != nil {
				return err
			}

			resp, err := runAttachmentUpload(cmd, func(ctx context.Context) (*data.AttachmentResponse, error) {
				return cli.AddAttachmentToPlan(ctx, planID, filePath)
			})
			if err != nil {
				return fmt.Errorf("failed to add attachment: %w", err)
			}

			ui.Successf(os.Stdout, "Attachment added (ID: %d)\n   URL: %s", resp.AttachmentID, resp.URL)
			return output.OutputResult(cmd, resp, "attachments")
		},
	}
	output.AddFlag(cmd)
	return cmd
}

// newAddPlanEntryCmd creates the 'attachments add plan-entry' command.
// Endpoint: POST /add_attachment_to_plan_entry/{plan_id}/{entry_id}
func newAddPlanEntryCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan-entry [plan_id] [entry_id] <file_path>",
		Short: "Добавить вложение к записи плана",
		Long:  `Загружает файл и прикрепляет его к записи (entry) в тест-плане.`,
		Example: `  # Прикрепить данные к записи плана
  gotr attachments add plan-entry 100 entry-abc123 ./data.csv

  # Прикрепить заметки
  gotr attachments add plan-entry 200 def456 ./notes.txt`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cli := getClient(cmd)

			var planID int64
			var entryID string
			var filePath string
			var err error

			switch len(args) {
			case 3:
				planID, err = flags.ValidateRequiredID(args, 0, "plan_id")
				if err != nil {
					return err
				}
				entryID = args[1]
				filePath = args[2]
			case 2:
				planID, err = flags.ValidateRequiredID(args, 0, "plan_id")
				if err != nil {
					return err
				}
				filePath = args[1]
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("entry_id required: gotr attachments add plan-entry [plan_id] [entry_id] <file_path>")
				}
				entryID, err = resolvePlanEntryIDInteractive(ctx, cli, planID)
				if err != nil {
					return err
				}
			default:
				filePath = args[0]
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("plan_id required: gotr attachments add plan-entry [plan_id] [entry_id] <file_path>")
				}
				planID, entryID, err = resolvePlanAndEntryIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("attachments add plan-entry")
				dr.PrintSimple("Add Attachment to Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s, File: %s", planID, entryID, filePath))
				return nil
			}

			if err := validateFileExists(filePath); err != nil {
				return err
			}

			resp, err := runAttachmentUpload(cmd, func(ctx context.Context) (*data.AttachmentResponse, error) {
				return cli.AddAttachmentToPlanEntry(ctx, planID, entryID, filePath)
			})
			if err != nil {
				return fmt.Errorf("failed to add attachment: %w", err)
			}

			ui.Successf(os.Stdout, "Attachment added (ID: %d)\n   URL: %s", resp.AttachmentID, resp.URL)
			return output.OutputResult(cmd, resp, "attachments")
		},
	}
	output.AddFlag(cmd)
	return cmd
}

// newAddResultCmd creates the 'attachments add result' command.
// Endpoint: POST /add_attachment_to_result/{result_id}
func newAddResultCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "result [result_id] <file_path>",
		Short: "Добавить вложение к результату теста",
		Long:  `Загружает файл и прикрепляет его к результату выполнения теста.`,
		Example: `  # Прикрепить лог к результату
  gotr attachments add result 98765 ./log.txt

  # Прикрепить скриншот ошибки
  gotr attachments add result 54321 ./screenshot.png`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cli := getClient(cmd)

			var resultID int64
			var filePath string
			var err error
			if len(args) == 2 {
				resultID, err = flags.ValidateRequiredID(args, 0, "result_id")
				if err != nil {
					return err
				}
				filePath = args[1]
			} else {
				filePath = args[0]
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("result_id required: gotr attachments add result [result_id] <file_path>")
				}
				resultID, err = resolveResultIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("attachments add result")
				dr.PrintSimple("Add Attachment to Result", fmt.Sprintf("Result ID: %d, File: %s", resultID, filePath))
				return nil
			}

			if err := validateFileExists(filePath); err != nil {
				return err
			}

			resp, err := runAttachmentUpload(cmd, func(ctx context.Context) (*data.AttachmentResponse, error) {
				return cli.AddAttachmentToResult(ctx, resultID, filePath)
			})
			if err != nil {
				return fmt.Errorf("failed to add attachment: %w", err)
			}

			ui.Successf(os.Stdout, "Attachment added (ID: %d)\n   URL: %s", resp.AttachmentID, resp.URL)
			return output.OutputResult(cmd, resp, "attachments")
		},
	}
	output.AddFlag(cmd)
	return cmd
}

// newAddRunCmd creates the 'attachments add run' command.
// Endpoint: POST /add_attachment_to_run/{run_id}
func newAddRunCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [run_id] <file_path>",
		Short: "Добавить вложение к тестовому прогону",
		Long:  `Загружает файл и прикрепляет его к тестовому прогону.`,
		Example: `  # Прикрепить HTML-отчёт к прогону
  gotr attachments add run 555 ./report.html

  # Прикрепить PDF-сводку
  gotr attachments add run 777 ./summary.pdf`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cli := getClient(cmd)

			var runID int64
			var filePath string
			var err error
			if len(args) == 2 {
				runID, err = flags.ValidateRequiredID(args, 0, "run_id")
				if err != nil {
					return err
				}
				filePath = args[1]
			} else {
				filePath = args[0]
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("run_id required: gotr attachments add run [run_id] <file_path>")
				}
				runID, err = resolveRunIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("attachments add run")
				dr.PrintSimple("Add Attachment to Run", fmt.Sprintf("Run ID: %d, File: %s", runID, filePath))
				return nil
			}

			if err := validateFileExists(filePath); err != nil {
				return err
			}

			resp, err := runAttachmentUpload(cmd, func(ctx context.Context) (*data.AttachmentResponse, error) {
				return cli.AddAttachmentToRun(ctx, runID, filePath)
			})
			if err != nil {
				return fmt.Errorf("failed to add attachment: %w", err)
			}

			ui.Successf(os.Stdout, "Attachment added (ID: %d)\n   URL: %s", resp.AttachmentID, resp.URL)
			return output.OutputResult(cmd, resp, "attachments")
		},
	}
	output.AddFlag(cmd)
	return cmd
}

// validateFileExists checks that the file exists at the given path.
func validateFileExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}
	return nil
}
