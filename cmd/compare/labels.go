package compare

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/spf13/cobra"
)

// newLabelsCmd creates the 'compare labels' subcommand.
func newLabelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "labels",
		Short: "Сравнить метки между проектами",
		Long: `Выполняет сравнение меток (labels) между двумя проектами.

Примеры:
  # Сравнить метки
  gotr compare labels --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare labels --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare labels --pid1 30 --pid2 31 --save-to labels_diff.json
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

			// Compare labels
			result, err := compareLabelsInternal(cli, pid1, pid2, pm)
			if err != nil {
				return fmt.Errorf("ошибка сравнения меток: %w", err)
			}

			// Print or save result
			return PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath)
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// labelsCmd — экспортированная команда
var labelsCmd = newLabelsCmd()

// compareLabelsInternal compares labels between two projects and returns the result.
func compareLabelsInternal(cli client.ClientInterface, pid1, pid2 int64, pm *progress.Manager) (*CompareResult, error) {
	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка меток из проекта %d...", pid1))
	labels1, err := fetchLabelItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения меток проекта %d: %w", pid1, err)
	}

	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка меток из проекта %d...", pid2))
	labels2, err := fetchLabelItems(cli, pid2)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения меток проекта %d: %w", pid2, err)
	}

	return compareItemInfos("labels", pid1, pid2, labels1, labels2), nil
}

// fetchLabelItems fetches all labels for a project and returns them as ItemInfo slice.
func fetchLabelItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	labels, err := cli.GetLabels(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(labels))
	for _, l := range labels {
		items = append(items, ItemInfo{
			ID:   l.ID,
			Name: l.Name,
		})
	}

	return items, nil
}
