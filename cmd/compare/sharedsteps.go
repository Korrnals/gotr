package compare

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newSharedStepsCmd creates the 'compare sharedsteps' subcommand.
func newSharedStepsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sharedsteps",
		Short: "Сравнить shared steps между проектами",
		Long: `Выполняет сравнение shared steps (общих шагов) между двумя проектами.

Примеры:
  # Сравнить shared steps
  gotr compare sharedsteps --pid1 30 --pid2 31

  # Сохранить результат
  gotr compare sharedsteps --pid1 30 --pid2 31 --format json --save sharedsteps_diff.json
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

			// Compare shared steps
			result, err := compareSharedStepsInternal(cli, pid1, pid2)
			if err != nil {
				return fmt.Errorf("ошибка сравнения shared steps: %w", err)
			}

			// Print or save result
			return PrintCompareResult(cmd, *result, project1Name, project2Name, format, saveFlag)
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// sharedStepsCmd — экспортированная команда
var sharedStepsCmd = newSharedStepsCmd()

// compareSharedStepsInternal compares shared steps between two projects and returns the result.
func compareSharedStepsInternal(cli client.ClientInterface, pid1, pid2 int64) (*CompareResult, error) {
	steps1, err := fetchSharedStepItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения shared steps проекта %d: %w", pid1, err)
	}

	steps2, err := fetchSharedStepItems(cli, pid2)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения shared steps проекта %d: %w", pid2, err)
	}

	return compareItemInfos("sharedsteps", pid1, pid2, steps1, steps2), nil
}

// fetchSharedStepItems fetches all shared steps for a project and returns them as ItemInfo slice.
func fetchSharedStepItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	steps, err := cli.GetSharedSteps(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(steps))
	for _, s := range steps {
		items = append(items, ItemInfo{
			ID:   s.ID,
			Name: s.Title,
		})
	}

	return items, nil
}
