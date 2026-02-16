package compare

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newPlansCmd creates the 'compare plans' subcommand.
func newPlansCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Short: "Сравнить test plans между проектами",
		Long: `Выполняет сравнение test plans между двумя проектами.

Примеры:
  # Сравнить test plans
  gotr compare plans --pid1 30 --pid2 31

  # Сохранить результат
  gotr compare plans --pid1 30 --pid2 31 --format json --save plans_diff.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClientSafe(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			// Parse flags
			pid1, pid2, format, saveFlag, err := parseCommonFlags(cmd)
			if err != nil {
				return err
			}

			// Get project names
			project1Name, project2Name, err := GetProjectNames(cli, pid1, pid2)
			if err != nil {
				return err
			}

			// Compare plans
			result, err := comparePlansInternal(cli, pid1, pid2)
			if err != nil {
				return fmt.Errorf("ошибка сравнения plans: %w", err)
			}

			// Print or save result
			return PrintCompareResult(cmd, *result, project1Name, project2Name, format, saveFlag)
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// plansCmd — экспортированная команда
var plansCmd = newPlansCmd()

// comparePlansInternal compares plans between two projects and returns the result.
func comparePlansInternal(cli client.ClientInterface, pid1, pid2 int64) (*CompareResult, error) {
	plans1, err := fetchPlanItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения plans проекта %d: %w", pid1, err)
	}

	plans2, err := fetchPlanItems(cli, pid2)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения plans проекта %d: %w", pid2, err)
	}

	return compareItemInfos("plans", pid1, pid2, plans1, plans2), nil
}

// fetchPlanItems fetches all plans for a project and returns them as ItemInfo slice.
func fetchPlanItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	plans, err := cli.GetPlans(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(plans))
	for _, p := range plans {
		items = append(items, ItemInfo{
			ID:   p.ID,
			Name: p.Name,
		})
	}

	return items, nil
}
