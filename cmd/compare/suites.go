package compare

import (
	"fmt"
	"strings"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/spf13/cobra"
)

// newSuitesCmd creates the 'compare suites' subcommand.
func newSuitesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suites",
		Short: "Сравнить тест-сюиты между проектами",
		Long: `Выполняет сравнение тест-сюитов между двумя проектами.

Примеры:
  # Сравнить сюиты
  gotr compare suites --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare suites --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare suites --pid1 30 --pid2 31 --save-to suites_diff.json
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

			// Compare suites
			result, err := compareSuitesInternal(cli, pid1, pid2, pm)
			if err != nil {
				return fmt.Errorf("ошибка сравнения сюитов: %w", err)
			}

			// Print or save result
			return PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath)
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// suitesCmd — экспортированная команда
var suitesCmd = newSuitesCmd()

// compareSuitesInternal compares suites between two projects and returns the result.
// Uses parallel API to fetch both projects simultaneously.
func compareSuitesInternal(cli client.ClientInterface, pid1, pid2 int64, pm *progress.Manager) (*CompareResult, error) {
	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Параллельная загрузка сьютов из проектов %d и %d...", pid1, pid2))

	// Fetch suites from both projects in parallel
	suitesByProject, err := cli.GetSuitesParallel([]int64{pid1, pid2}, 2)
	if err != nil && len(suitesByProject) == 0 {
		return nil, fmt.Errorf("ошибка получения сюитов: %w", err)
	}

	// Convert to ItemInfo slices
	suites1 := suitesToItems(suitesByProject[pid1])
	suites2 := suitesToItems(suitesByProject[pid2])

	if err != nil {
		// Partial failure - log warning but continue with what we have
		fmt.Printf("⚠ Предупреждение: не все проекты загружены: %v\n", err)
	}

	return compareItemInfos("suites", pid1, pid2, suites1, suites2), nil
}

// suitesToItems converts GetSuitesResponse to []ItemInfo
func suitesToItems(suites data.GetSuitesResponse) []ItemInfo {
	items := make([]ItemInfo, 0, len(suites))
	for _, s := range suites {
		items = append(items, ItemInfo{
			ID:   s.ID,
			Name: s.Name,
		})
	}
	return items
}

// fetchSuiteItems fetches all suites for a project and returns them as ItemInfo slice.
func fetchSuiteItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	suites, err := cli.GetSuites(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(suites))
	for _, s := range suites {
		items = append(items, ItemInfo{
			ID:   s.ID,
			Name: s.Name,
		})
	}

	return items, nil
}

// compareItemInfos compares two slices of ItemInfo and returns a CompareResult.
func compareItemInfos(resource string, pid1, pid2 int64, items1, items2 []ItemInfo) *CompareResult {
	// Build name maps
	map1 := make(map[string]ItemInfo)
	map2 := make(map[string]ItemInfo)

	for _, item := range items1 {
		key := strings.ToLower(strings.TrimSpace(item.Name))
		if key != "" {
			map1[key] = item
		}
	}

	for _, item := range items2 {
		key := strings.ToLower(strings.TrimSpace(item.Name))
		if key != "" {
			map2[key] = item
		}
	}

	// Compare
	var onlyInFirst, onlyInSecond []ItemInfo
	var common []CommonItemInfo

	// Items only in first
	for key, item := range map1 {
		if _, found := map2[key]; !found {
			onlyInFirst = append(onlyInFirst, item)
		}
	}

	// Items only in second
	for key, item := range map2 {
		if _, found := map1[key]; !found {
			onlyInSecond = append(onlyInSecond, item)
		}
	}

	// Common items
	for key, item1 := range map1 {
		if item2, found := map2[key]; found {
			common = append(common, CommonItemInfo{
				Name:     item1.Name,
				ID1:      item1.ID,
				ID2:      item2.ID,
				IDsMatch: item1.ID == item2.ID,
			})
		}
	}

	return &CompareResult{
		Resource:     resource,
		Project1ID:   pid1,
		Project2ID:   pid2,
		OnlyInFirst:  onlyInFirst,
		OnlyInSecond: onlyInSecond,
		Common:       common,
	}
}
