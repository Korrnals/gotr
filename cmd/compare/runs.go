package compare

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/spf13/cobra"
)

// newRunsCmd creates the 'compare runs' subcommand.
func newRunsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runs",
		Short: "Сравнить test runs между проектами",
		Long: `Выполняет сравнение test runs между двумя проектами.

Примеры:
  # Сравнить test runs
  gotr compare runs --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare runs --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare runs --pid1 30 --pid2 31 --save-to runs_diff.json
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

			// Compare runs
			result, err := compareRunsInternal(cli, pid1, pid2, pm)
			if err != nil {
				return fmt.Errorf("ошибка сравнения runs: %w", err)
			}

			// Print or save result
			return PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath)
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// runsCmd — экспортированная команда
var runsCmd = newRunsCmd()

// compareRunsInternal compares runs between two projects and returns the result.
func compareRunsInternal(cli client.ClientInterface, pid1, pid2 int64, pm *progress.Manager) (*CompareResult, error) {
	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка runs из проекта %d...", pid1))
	runs1, err := fetchRunItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения runs проекта %d: %w", pid1, err)
	}

	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка runs из проекта %d...", pid2))
	runs2, err := fetchRunItems(cli, pid2)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения runs проекта %d: %w", pid2, err)
	}

	return compareItemInfos("runs", pid1, pid2, runs1, runs2), nil
}

// fetchRunItems fetches all runs for a project and returns them as ItemInfo slice.
func fetchRunItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	runs, err := cli.GetRuns(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(runs))
	for _, r := range runs {
		items = append(items, ItemInfo{
			ID:   r.ID,
			Name: r.Name,
		})
	}

	return items, nil
}
