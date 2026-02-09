package cases

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду 'cases add'
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <section_id>",
		Short: "Создать новый тест-кейс",
		Long:  `Создаёт новый тест-кейс в указанной секции.`,
		Example: `  # Создать тест-кейс с параметрами
  gotr cases add 100 --title="Тест авторизации" --template-id=1

  # Создать из JSON-файла
  gotr cases add 100 --json-file=case.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sectionID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || sectionID <= 0 {
				return fmt.Errorf("invalid section_id: %s", args[0])
			}

			// Check JSON file
			jsonFile, _ := cmd.Flags().GetString("json-file")
			var req data.AddCaseRequest

			if jsonFile != "" {
				jsonData, err := os.ReadFile(jsonFile)
				if err != nil {
					return fmt.Errorf("error reading JSON file: %w", err)
				}
				if err := json.Unmarshal(jsonData, &req); err != nil {
					return fmt.Errorf("error parsing JSON: %w", err)
				}
			} else {
				title, _ := cmd.Flags().GetString("title")
				if title == "" {
					return fmt.Errorf("--title is required (or use --json-file)")
				}
				req.Title = title
				req.TemplateID, _ = cmd.Flags().GetInt64("template-id")
				req.TypeID, _ = cmd.Flags().GetInt64("type-id")
				req.PriorityID, _ = cmd.Flags().GetInt64("priority-id")
				req.Refs, _ = cmd.Flags().GetString("refs")
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases add")
				dr.PrintSimple("Create Case", fmt.Sprintf("Section ID: %d, Title: %s", sectionID, req.Title))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.AddCase(sectionID, &req)
			if err != nil {
				return fmt.Errorf("failed to create case: %w", err)
			}

			fmt.Printf("✅ Case created (ID: %d)\n", resp.ID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без создания")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")
	cmd.Flags().String("json-file", "", "Путь к JSON-файлу с данными кейса")
	cmd.Flags().String("title", "", "Название тест-кейса")
	cmd.Flags().Int64("template-id", 0, "ID шаблона")
	cmd.Flags().Int64("type-id", 0, "ID типа теста")
	cmd.Flags().Int64("priority-id", 0, "ID приоритета")
	cmd.Flags().String("refs", "", "Ссылки (референсы)")

	return cmd
}
