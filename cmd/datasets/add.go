package datasets

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду 'datasets add'
// Эндпоинт: POST /add_dataset/{project_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [project_id]",
		Short: "Создать новый датасет",
		Long: `Создаёт новый датасет (набор тестовых данных) в указанном проекте.

Датасет создаётся с указанным названием. После создания можно
добавлять колонки (параметры) и строки (значения) через веб-интерфейс
или другие API методы.`,
		Example: `  # Создать датасет с названием
  gotr datasets add 1 --name="Login Test Data"

  # Проверить перед созданием
  gotr datasets add 1 --name="Test Data" --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cli := getClient(cmd)

			var projectID int64
			var err error
			if len(args) > 0 {
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id required: gotr datasets add [project_id] --name <name>")
				}
				projectID, err = resolveProjectIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("datasets add")
				dr.PrintSimple("Создать датасет", fmt.Sprintf("Project ID: %d, Name: %s", projectID, name))
				return nil
			}

			resp, err := cli.AddDataset(ctx, projectID, name)
			if err != nil {
				return fmt.Errorf("failed to create dataset: %w", err)
			}

			ui.Successf(os.Stdout, "Dataset created (ID: %d)", resp.ID)
			return output.OutputResult(cmd, resp, "datasets")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без создания")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Название датасета (обязательно)")

	return cmd
}
