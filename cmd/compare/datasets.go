package compare

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newDatasetsCmd creates the 'compare datasets' subcommand.
func newDatasetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datasets",
		Short: "Сравнить datasets между проектами",
		Long: `Выполняет сравнение datasets между двумя проектами.

Примеры:
  # Сравнить datasets
  gotr compare datasets --pid1 30 --pid2 31

  # Сохранить результат
  gotr compare datasets --pid1 30 --pid2 31 --format json --save datasets_diff.json
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

			// Compare datasets
			result, err := compareDatasetsInternal(cli, pid1, pid2)
			if err != nil {
				return fmt.Errorf("ошибка сравнения datasets: %w", err)
			}

			// Print or save result
			return PrintCompareResult(cmd, *result, project1Name, project2Name, format, saveFlag)
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// datasetsCmd — экспортированная команда
var datasetsCmd = newDatasetsCmd()

// compareDatasetsInternal compares datasets between two projects and returns the result.
func compareDatasetsInternal(cli client.ClientInterface, pid1, pid2 int64) (*CompareResult, error) {
	datasets1, err := fetchDatasetItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения datasets проекта %d: %w", pid1, err)
	}

	datasets2, err := fetchDatasetItems(cli, pid2)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения datasets проекта %d: %w", pid2, err)
	}

	return compareItemInfos("datasets", pid1, pid2, datasets1, datasets2), nil
}

// fetchDatasetItems fetches all datasets for a project and returns them as ItemInfo slice.
func fetchDatasetItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	datasets, err := cli.GetDatasets(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(datasets))
	for _, d := range datasets {
		items = append(items, ItemInfo{
			ID:   d.ID,
			Name: d.Name,
		})
	}

	return items, nil
}
