package bdds

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду 'bdds add'
// Эндпоинт: POST /add_bdd/{test_case_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [case_id]",
		Short: "Добавить BDD сценарий к тест-кейсу",
		Long: `Добавляет BDD сценарий в формате Gherkin к указанному тест-кейсу.

Сценарий должен быть в формате Given-When-Then (Дано-Когда-Тогда).
Можно передать содержимое через файл или напрямую.`,
		Example: `  # Добавить BDD из файла
  gotr bdds add 12345 --file=scenario.feature

  # Добавить BDD из stdin
  cat scenario.feature | gotr bdds add 12345`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var caseID int64
			var err error
			if len(args) > 0 {
				caseID, err = flags.ValidateRequiredID(args, 0, "case_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("case_id is required in non-interactive mode: gotr bdds add [case_id]")
				}
				if _, ok := interactive.PrompterFromContext(ctx).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("case_id is required in non-interactive mode: gotr bdds add [case_id]")
				}
				caseID, err = resolveCaseIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			// Читаем содержимое BDD
			content, err := readBDDContent(cmd)
			if err != nil {
				return err
			}
			if content == "" {
				return fmt.Errorf("BDD content cannot be empty (use --file or stdin)")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("bdds add")
				dr.PrintSimple("Добавить BDD", fmt.Sprintf("Case ID: %d", caseID))
				return nil
			}

			resp, err := cli.AddBDD(ctx, caseID, content)
			if err != nil {
				return fmt.Errorf("failed to add BDD: %w", err)
			}

			ui.Successf(os.Stdout, "BDD added to case %d", caseID)
			return output.OutputResult(cmd, resp, "bdds")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без добавления")
	output.AddFlag(cmd)
	cmd.Flags().String("file", "", "Путь к файлу с Gherkin сценарием")

	return cmd
}

// readBDDContent читает содержимое BDD из файла или stdin
func readBDDContent(cmd *cobra.Command) (string, error) {
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return string(data), nil
	}

	// TODO: Чтение из stdin если файл не указан
	// Пока возвращаем пустую строку, будет ошибка валидации
	return "", nil
}
