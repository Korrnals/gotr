package compare

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/progress"
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

  # Сохранить результат в файл по умолчанию
  gotr compare milestones --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare milestones --pid1 30 --pid2 31 --save-to milestones_diff.json
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

			// Create progress manager
			pm := progress.NewManager()

			// Start timer
			startTime := time.Now()

			// Compare milestones
			result, err := compareMilestonesInternal(cli, pid1, pid2, pm)
			if err != nil {
				return fmt.Errorf("ошибка сравнения milestones: %w", err)
			}

			elapsed := time.Since(startTime)

			// Print or save result
			if err := PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath); err != nil {
				return err
			}

			// Print statistics
			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				PrintCompareStats("milestones", pid1, pid2,
					len(result.OnlyInFirst), len(result.OnlyInSecond), len(result.Common), elapsed)
			}

			return nil
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// milestonesCmd — экспортированная команда
var milestonesCmd = newMilestonesCmd()

// compareMilestonesInternal compares milestones between two projects and returns the result.
func compareMilestonesInternal(cli client.ClientInterface, pid1, pid2 int64, pm ...*progress.Manager) (*CompareResult, error) {
	var p *progress.Manager
	if len(pm) > 0 {
		p = pm[0]
	}
	progress.Describe(p.NewSpinner(""), fmt.Sprintf("Загрузка milestones из проекта %d...", pid1))
	milestones1, err := fetchMilestoneItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения milestones проекта %d: %w", pid1, err)
	}

	progress.Describe(p.NewSpinner(""), fmt.Sprintf("Загрузка milestones из проекта %d...", pid2))
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
