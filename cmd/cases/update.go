package cases

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду 'cases update'
// Эндпоинт: POST /update_case/{case_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <case_id>",
		Short: "Обновить тест-кейс",
		Long:  `Обновляет существующий тест-кейс.`,
		Example: `  # Обновить название и приоритет
  gotr cases update 12345 --title="Новое название" --priority-id=2

  # Обновить из JSON-файла
  gotr cases update 12345 --json-file=update.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			caseID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || caseID <= 0 {
				return fmt.Errorf("invalid case_id: %s", args[0])
			}

			jsonFile, _ := cmd.Flags().GetString("json-file")
			var req data.UpdateCaseRequest

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
				if title != "" {
					req.Title = &title
				}
				if v, _ := cmd.Flags().GetInt64("type-id"); v > 0 {
					req.TypeID = &v
				}
				if v, _ := cmd.Flags().GetInt64("priority-id"); v > 0 {
					req.PriorityID = &v
				}
				if v, _ := cmd.Flags().GetString("refs"); v != "" {
					req.Refs = &v
				}
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases update")
				dr.PrintSimple("Update Case", fmt.Sprintf("Case ID: %d", caseID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdateCase(caseID, &req)
			if err != nil {
				return fmt.Errorf("failed to update case: %w", err)
			}

			fmt.Printf("✅ Case %d updated\n", caseID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	save.AddFlag(cmd)
	cmd.Flags().String("json-file", "", "Путь к JSON-файлу с данными для обновления")
	cmd.Flags().String("title", "", "Новое название")
	cmd.Flags().Int64("type-id", 0, "Новый ID типа")
	cmd.Flags().Int64("priority-id", 0, "Новый ID приоритета")
	cmd.Flags().String("refs", "", "Новые ссылки")

	return cmd
}
