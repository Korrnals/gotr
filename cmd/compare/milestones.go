package compare

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newMilestonesCmd creates the 'compare milestones' subcommand.
func newMilestonesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "milestones",
		Short: "Сравнить milestones между проектами",
		Long: `Выполняет сравнение milestones между двумя проектами.

Примеры:
  # Сравнить milestones
  gotr compare milestones --pid1 30 --pid2 31

  # Сохранить результат
  gotr compare milestones --pid1 30 --pid2 31 --format json --save milestones_diff.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClientSafe(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			// Parse flags
			pid1, pid2, format, savePath, err := parseCommonFlags(cmd)
			if err != nil {
				return err
			}

			// Get project names
			project1Name, project2Name, err := GetProjectNames(cli, pid1, pid2)
			if err != nil {
				return err
			}

			// Compare milestones
			result, err := compareMilestonesInternal(cli, pid1, pid2)
			if err != nil {
				return fmt.Errorf("ошибка сравнения milestones: %w", err)
			}

			// Print or save result
			return PrintCompareResult(*result, project1Name, project2Name, format, savePath)
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// milestonesCmd — экспортированная команда
var milestonesCmd = newMilestonesCmd()

// compareMilestonesInternal compares milestones between two projects and returns the result.
func compareMilestonesInternal(cli client.ClientInterface, pid1, pid2 int64) (*CompareResult, error) {
	milestones1, err := fetchMilestoneItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения milestones проекта %d: %w", pid1, err)
	}

	milestones2, err := fetchMilestoneItems(cli, pid2)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения milestones проекта %d: %w", pid2, err)
	}

	return compareItemInfos("milestones", pid1, pid2, milestones1, milestones2), nil
}

// fetchMilestoneItems fetches all milestones for a project and returns them as ItemInfo slice.
func fetchMilestoneItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	milestones, err := cli.GetMilestones(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(milestones))
	for _, m := range milestones {
		items = append(items, ItemInfo{
			ID:   m.ID,
			Name: m.Name,
		})
	}

	return items, nil
}
