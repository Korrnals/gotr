package compare

import (
	"fmt"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// newCasesCmd creates the 'compare cases' subcommand.
func newCasesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cases",
		Short: "Сравнить тест-кейсы между проектами",
		Long: `Выполняет сравнение тест-кейсов между двумя проектами.

По умолчанию сравнение выполняется по полю 'title'.
Можно указать другое поле для сравнения с помощью флага --field.

Поддерживаемые поля:
  title, priority_id, type_id, milestone_id, refs, 
  custom_preconds, custom_steps, custom_expected и др.

Примеры:
  # Сравнить кейсы по названию
  gotr compare cases --pid1 30 --pid2 31

  # Сравнить по приоритету
  gotr compare cases --pid1 30 --pid2 31 --field priority_id

  # Сохранить результат в файл по умолчанию
  gotr compare cases --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare cases --pid1 30 --pid2 31 --save-to cases_diff.json
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

			field, _ := cmd.Flags().GetString("field")
			if field == "" {
				field = "title"
			}

			quiet, _ := cmd.Flags().GetBool("quiet")

			// Get project names
			project1Name, project2Name, err := GetProjectNames(cli, pid1, pid2)
			if err != nil {
				return err
			}

			// Create progress manager (quiet mode disables progress bars)
			var pm *progress.Manager
			if !quiet {
				pm = progress.NewManager()
			}

			// Start timer
			startTime := time.Now()

			// Compare cases
			result, err := compareCasesInternal(cli, pid1, pid2, field, pm)
			if err != nil {
				return fmt.Errorf("ошибка сравнения кейсов: %w", err)
			}

			// Calculate elapsed time
			elapsed := time.Since(startTime)

			// Print or save result (this also finishes any progress bars)
			if err := PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath); err != nil {
				return err
			}

			// Print additional field diff information for cases
			if field != "title" && !quiet {
				printCasesFieldDiff(cli, pid1, pid2, field)
			}

			// Print statistics (unless quiet mode)
			if !quiet {
				printCasesStats(result, elapsed)
			}

			return nil
		},
	}

	// Add flags
	addCommonFlags(cmd)
	cmd.Flags().String("field", "title", "Поле для сравнения (title, priority_id, и т.д.)")

	return cmd
}

// casesCmd — экспортированная команда
var casesCmd = newCasesCmd()

// compareCasesInternal compares cases between two projects and returns the result.
// Shows single informative progress bar for parallel loading from both projects.
func compareCasesInternal(cli client.ClientInterface, pid1, pid2 int64, field string, pm *progress.Manager) (*CompareResult, error) {
	// Get suites count for both projects first (quick operation)
	suitesMap, err := cli.GetSuitesParallel([]int64{pid1, pid2}, 2)
	if err != nil && len(suitesMap) == 0 {
		return nil, fmt.Errorf("ошибка получения сьютов: %w", err)
	}

	suites1 := suitesMap[pid1]
	suites2 := suitesMap[pid2]
	totalSuites := len(suites1) + len(suites2)

	// Create single progress bar for all suites (if there are any)
	var bar *progressbar.ProgressBar
	if pm != nil && totalSuites > 0 {
		bar = pm.NewBar(int64(totalSuites), 
			fmt.Sprintf("Параллельная загрузка кейсов (%d сьютов)...", totalSuites))
	}

	// Get cases for both projects with shared progress bar
	cases1, err := fetchCaseItemsWithProgress(cli, pid1, suites1, bar)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения кейсов проекта %d: %w", pid1, err)
	}

	cases2, err := fetchCaseItemsWithProgress(cli, pid2, suites2, bar)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения кейсов проекта %d: %w", pid2, err)
	}

	progress.Finish(bar)

	// Build name maps for comparison
	cases1Map := make(map[string]ItemInfo)
	cases2Map := make(map[string]ItemInfo)

	for _, c := range cases1 {
		key := getCaseKey(c, field)
		if key != "" {
			cases1Map[strings.ToLower(key)] = c
		}
	}

	for _, c := range cases2 {
		key := getCaseKey(c, field)
		if key != "" {
			cases2Map[strings.ToLower(key)] = c
		}
	}

	// Compare
	var onlyInFirst, onlyInSecond []ItemInfo
	var common []CommonItemInfo

	// Items only in first
	for key, item := range cases1Map {
		if _, found := cases2Map[key]; !found {
			onlyInFirst = append(onlyInFirst, item)
		}
	}

	// Items only in second
	for key, item := range cases2Map {
		if _, found := cases1Map[key]; !found {
			onlyInSecond = append(onlyInSecond, item)
		}
	}

	// Common items
	for key, item1 := range cases1Map {
		if item2, found := cases2Map[key]; found {
			common = append(common, CommonItemInfo{
				Name:     item1.Name,
				ID1:      item1.ID,
				ID2:      item2.ID,
				IDsMatch: item1.ID == item2.ID,
			})
		}
	}

	return &CompareResult{
		Resource:     "cases",
		Project1ID:   pid1,
		Project2ID:   pid2,
		OnlyInFirst:  onlyInFirst,
		OnlyInSecond: onlyInSecond,
		Common:       common,
	}, nil
}

// fetchCaseItemsWithProgress fetches all cases for a project with progress updates.
func fetchCaseItemsWithProgress(cli client.ClientInterface, projectID int64, suites data.GetSuitesResponse, bar *progressbar.ProgressBar) ([]ItemInfo, error) {
	// If no suites, fetch cases without suite filter
	if len(suites) == 0 {
		cases, err := cli.GetCases(projectID, 0, 0)
		if err != nil {
			return nil, err
		}

		allCases := make([]ItemInfo, 0, len(cases))
		for _, c := range cases {
			allCases = append(allCases, ItemInfo{
				ID:   c.ID,
				Name: c.Title,
			})
		}
		return allCases, nil
	}

	// Extract suite IDs
	suiteIDs := make([]int64, len(suites))
	for i, s := range suites {
		suiteIDs[i] = s.ID
	}

	// Fetch cases in parallel using concurrent API (5 workers, rate limited)
	casesBySuite, err := cli.GetCasesParallel(projectID, suiteIDs, 5)
	if err != nil && len(casesBySuite) == 0 {
		return nil, err
	}

	// Collect unique cases and update progress
	var allCases []ItemInfo
	caseIDs := make(map[int64]bool)

	for _, cases := range casesBySuite {
		for _, c := range cases {
			if !caseIDs[c.ID] {
				caseIDs[c.ID] = true
				allCases = append(allCases, ItemInfo{
					ID:   c.ID,
					Name: c.Title,
				})
			}
		}
		// Update progress for this suite
		if bar != nil {
			progress.Add(bar, 1)
		}
	}

	return allCases, nil
}

// fetchCaseItems fetches all cases for a project and returns them as ItemInfo slice.
// Uses parallel API for significant performance improvement (4-5x faster).
// DEPRECATED: Use fetchCaseItemsWithProgress for better UX.
func fetchCaseItems(cli client.ClientInterface, projectID int64, pm *progress.Manager) ([]ItemInfo, error) {
	// Get all suites for the project
	suites, err := cli.GetSuites(projectID)
	if err != nil {
		return nil, err
	}

	// If no suites, fetch cases without suite filter
	if len(suites) == 0 {
		cases, err := cli.GetCases(projectID, 0, 0)
		if err != nil {
			return nil, err
		}

		allCases := make([]ItemInfo, 0, len(cases))
		for _, c := range cases {
			allCases = append(allCases, ItemInfo{
				ID:   c.ID,
				Name: c.Title,
			})
		}
		return allCases, nil
	}

	// Create progress bar
	var bar *progressbar.ProgressBar
	if pm != nil {
		bar = pm.NewBar(int64(len(suites)), fmt.Sprintf("Параллельная загрузка из %d сьютов...", len(suites)))
	}

	// Extract suite IDs
	suiteIDs := make([]int64, len(suites))
	for i, s := range suites {
		suiteIDs[i] = s.ID
	}

	// Fetch cases in parallel using concurrent API (5 workers, rate limited)
	casesBySuite, err := cli.GetCasesParallel(projectID, suiteIDs, 5)
	if err != nil && len(casesBySuite) == 0 {
		return nil, err
	}

	// Collect unique cases
	var allCases []ItemInfo
	caseIDs := make(map[int64]bool)

	for _, cases := range casesBySuite {
		for _, c := range cases {
			if !caseIDs[c.ID] {
				caseIDs[c.ID] = true
				allCases = append(allCases, ItemInfo{
					ID:   c.ID,
					Name: c.Title,
				})
			}
		}
		// Update progress for this suite
		if bar != nil {
			progress.Add(bar, 1)
		}
	}

	progress.Finish(bar)
	return allCases, nil
}

// getCaseKey returns the comparison key for a case based on the field.
func getCaseKey(item ItemInfo, field string) string {
	// For title field, use the name directly
	if field == "title" {
		return item.Name
	}
	// For other fields, we'd need the full case data
	// This is simplified - in real implementation, we'd store the field value
	return item.Name
}

// printCasesFieldDiff prints differences by field for cases.
func printCasesFieldDiff(cli client.ClientInterface, pid1, pid2 int64, field string) {
	diff, err := cli.DiffCasesData(pid1, pid2, field)
	if err != nil {
		fmt.Printf("\nОшибка получения различий по полю '%s': %v\n", field, err)
		return
	}

	if len(diff.DiffByField) == 0 {
		fmt.Printf("\nОтличий по полю '%s' не найдено.\n", field)
		return
	}

	fmt.Printf("\n=== Отличия по полю '%s' ===\n", field)
	for _, d := range diff.DiffByField {
		firstValue := getFieldValue(d.First, field)
		secondValue := getFieldValue(d.Second, field)

		fmt.Printf("\nКейс: %s (ID: %d)\n", d.First.Title, d.CaseID)
		fmt.Printf("  Проект %d: %s\n", pid1, firstValue)
		fmt.Printf("  Проект %d: %s\n", pid2, secondValue)
	}
}

// getFieldValue extracts a field value from a Case using reflection-like access.
func getFieldValue(c data.Case, field string) string {
	switch field {
	case "title":
		return c.Title
	case "priority_id":
		return fmt.Sprintf("%d", c.PriorityID)
	case "type_id":
		return fmt.Sprintf("%d", c.TypeID)
	case "milestone_id":
		return fmt.Sprintf("%d", c.MilestoneID)
	case "refs":
		return c.Refs
	case "custom_preconds":
		return c.CustomPreconds
	case "custom_steps":
		return c.CustomSteps
	case "custom_expected":
		return c.CustomExpected
	default:
		return "<unknown field>"
	}
}

// printCasesStats prints execution statistics for compare cases
func printCasesStats(result *CompareResult, elapsed time.Duration) {
	totalCases := len(result.OnlyInFirst) + len(result.OnlyInSecond) + len(result.Common)
	
	fmt.Println()
	fmt.Println("┌──────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    СТАТИСТИКА ВЫПОЛНЕНИЯ                     │")
	fmt.Println("├──────────────────────────────────────────────────────────────┤")
	fmt.Printf("│  Время выполнения: %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("│  Всего кейсов обработано: %d\n", totalCases)
	fmt.Println("├──────────────────────────────────────────────────────────────┤")
	fmt.Printf("│  Только в проекте %d: %d кейсов\n", result.Project1ID, len(result.OnlyInFirst))
	fmt.Printf("│  Только в проекте %d: %d кейсов\n", result.Project2ID, len(result.OnlyInSecond))
	fmt.Printf("│  Общих кейсов: %d\n", len(result.Common))
	fmt.Println("└──────────────────────────────────────────────────────────────┘")
}
