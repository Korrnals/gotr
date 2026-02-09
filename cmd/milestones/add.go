package milestones

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду 'milestones add'
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <project_id>",
		Short: "Создать новый майлстон",
		Long: `Создаёт новый майлстон (веху) в указанном проекте.

Майлстон — это этап разработки, к которому привязываются тестовые прогоны.
Можно указать дедлайн, описание и родительский майлстон для иерархии.

Примеры использования:
  # Создать простой майлстон
  gotr milestones add 1 --name="Релиз 1.0"

  # Майлстон с дедлайном и описанием
  gotr milestones add 1 --name="Спринт 5" --due-on="2026-03-15" --description="Цель спринта"

  # Вложенный майлстон (подэтап)
  gotr milestones add 1 --name="Итерация 1.1" --parent-id=123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("invalid project_id: %s", args[0])
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			req := data.AddMilestoneRequest{
				Name: name,
			}

			if v, _ := cmd.Flags().GetString("description"); v != "" {
				req.Description = v
			}
			if v, _ := cmd.Flags().GetString("due-on"); v != "" {
				req.DueOn = v
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("milestones add")
				dr.PrintSimple("Create Milestone", fmt.Sprintf("Project ID: %d, Name: %s", projectID, req.Name))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.AddMilestone(projectID, &req)
			if err != nil {
				return fmt.Errorf("failed to create milestone: %w", err)
			}

			fmt.Printf("✅ Milestone created (ID: %d)\n", resp.ID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без реального выполнения")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")
	cmd.Flags().String("name", "", "Название майлстона (обязательно)")
	cmd.Flags().String("description", "", "Описание майлстона")
	cmd.Flags().String("due-on", "", "Дедлайн в формате YYYY-MM-DD")

	return cmd
}
