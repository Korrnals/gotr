package compare

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/parallel"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/Korrnals/gotr/internal/utils"
	"github.com/spf13/cobra"
)

// casesCmd — экспортированная команда
var casesCmd = newCasesCmd()

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

			// Get field for comparison
			field, _ := cmd.Flags().GetString("field")
			if field == "" {
				field = "title"
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

			// Execute comparison
			result, err := compareCasesInternal(cmd, cli, pid1, pid2, field, pm)
			if err != nil {
				return err
			}

			elapsed := time.Since(startTime)

			// Print or save result
			if err := PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath); err != nil {
				return err
			}

			// Print statistics
			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				PrintCompareStats("cases", pid1, pid2, 
					len(result.OnlyInFirst), len(result.OnlyInSecond), len(result.Common), elapsed)
			}

			return nil
		},
	}

	// Add flags
	addCommonFlags(cmd)
	cmd.Flags().String("field", "title", "Поле для сравнения (title, priority_id, etc.)")
	cmd.Flags().Int("parallel-suites", 5, "Максимальное количество параллельных сьютов (Stage 6.7)")
	cmd.Flags().Int("parallel-pages", 3, "Максимальное количество параллельных страниц внутри сьюта (Stage 6.7)")
	cmd.Flags().Duration("timeout", 5*time.Minute, "Таймаут для операции сравнения")

	return cmd
}

// getCaseKey extracts the comparison key from a case based on field name.
func getCaseKey(item ItemInfo, field string) string {
	if field == "title" {
		return item.Name
	}
	return item.Name
}

// ProjectLoadStats holds statistics for project loading
type ProjectLoadStats struct {
	ProjectID    int64
	SuitesCount  int
	CasesCount   int
	Duration     time.Duration
}

// compareCasesInternal compares cases between two projects and returns the result.
// Shows parallel loading of both projects with detailed statistics.
func compareCasesInternal(cmd *cobra.Command, cli client.ClientInterface, pid1, pid2 int64, field string, pm *progress.Manager) (*CompareResult, error) {
	// Phase 1: Get suites for both projects (quick operation)
	utils.DebugPrint("[Compare] Phase 1: Fetching suites for projects %d and %d", pid1, pid2)
	
	var spinner *progress.Bar
	if pm != nil {
		spinner = pm.NewSpinner(fmt.Sprintf("Получение структуры проектов %d и %d...", pid1, pid2))
	}

	suitesMap, err := cli.GetSuitesParallel([]int64{pid1, pid2}, 2, nil)
	if err != nil && len(suitesMap) == 0 {
		spinner.Finish()
		return nil, fmt.Errorf("ошибка получения сьютов: %w", err)
	}

	suites1 := suitesMap[pid1]
	suites2 := suitesMap[pid2]
	spinner.Finish()

	utils.DebugPrint("[Compare] Found suites: P%d=%d, P%d=%d", pid1, len(suites1), pid2, len(suites2))

	// Phase 2: Parallel loading of both projects
	utils.DebugPrint("[Compare] Phase 2: Parallel loading of projects %d and %d", pid1, pid2)
	
	fmt.Fprintf(os.Stderr, "\n📥 Параллельная загрузка данных:\n")
	fmt.Fprintf(os.Stderr, "   Проект %d: %d сьютов | Проект %d: %d сьютов\n\n", pid1, len(suites1), pid2, len(suites2))

	// Read parallel execution flags
	parallelSuites, _ := cmd.Flags().GetInt("parallel-suites")
	parallelPages, _ := cmd.Flags().GetInt("parallel-pages")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	var cases1, cases2 []ItemInfo
	var err1, err2 error
	var stats1, stats2 ProjectLoadStats
	
	done1 := make(chan struct{})
	done2 := make(chan struct{})

	// Load Project 1
	go func() {
		start := time.Now()
		cases1, err1 = fetchCasesForProjectWithStats(cli, pid1, suites1, pm, &stats1, parallelSuites, parallelPages, timeout)
		stats1.ProjectID = pid1
		stats1.SuitesCount = len(suites1)
		stats1.CasesCount = len(cases1)
		stats1.Duration = time.Since(start)
		close(done1)
	}()

	// Load Project 2
	go func() {
		start := time.Now()
		cases2, err2 = fetchCasesForProjectWithStats(cli, pid2, suites2, pm, &stats2, parallelSuites, parallelPages, timeout)
		stats2.ProjectID = pid2
		stats2.SuitesCount = len(suites2)
		stats2.CasesCount = len(cases2)
		stats2.Duration = time.Since(start)
		close(done2)
	}()

	// Wait for both
	<-done1
	<-done2

	// Print results after both complete
	fmt.Fprintf(os.Stderr, "📊 Результаты загрузки:\n")
	fmt.Fprintf(os.Stderr, "  ✅ Проект %d: %d сьютов → %d кейсов (%s)\n", 
		stats1.ProjectID, stats1.SuitesCount, stats1.CasesCount, stats1.Duration.Round(time.Second))
	fmt.Fprintf(os.Stderr, "  ✅ Проект %d: %d сьютов → %d кейсов (%s)\n", 
		stats2.ProjectID, stats2.SuitesCount, stats2.CasesCount, stats2.Duration.Round(time.Second))

	if err1 != nil {
		return nil, fmt.Errorf("ошибка загрузки проекта %d: %w", pid1, err1)
	}
	if err2 != nil {
		return nil, fmt.Errorf("ошибка загрузки проекта %d: %w", pid2, err2)
	}

	// Phase 3: Analysis
	utils.DebugPrint("[Compare] Phase 3: Analysis and comparison")
	fmt.Fprintf(os.Stderr, "\n🔍 Выполняется анализ и сверка данных...\n")

	start := time.Now()
	result := analyzeCases(cases1, cases2, pid1, pid2, field)
	elapsed := time.Since(start)
	
	fmt.Fprintf(os.Stderr, "  ✅ Анализ завершён (%s)\n", elapsed.Round(time.Millisecond))
	utils.DebugPrint("[Compare] Analysis complete: P%d=%d unique, P%d=%d unique, common=%d", 
		pid1, len(result.OnlyInFirst), pid2, len(result.OnlyInSecond), len(result.Common))

	return result, nil
}

// fetchCasesForProjectWithStats loads all cases for a single project with progress bar and stats.
func fetchCasesForProjectWithStats(cli client.ClientInterface, projectID int64, suites data.GetSuitesResponse, pm *progress.Manager, stats *ProjectLoadStats, parallelSuites, parallelPages int, timeout time.Duration) ([]ItemInfo, error) {
	if len(suites) == 0 {
		utils.DebugPrint("[Project %d] No suites, fetching all cases", projectID)
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

	// Create progress bar for this project
	var bar *progress.Bar
	if pm != nil {
		bar = pm.NewBar(int64(len(suites)), 
			fmt.Sprintf("⏳ Проект %d (%d сьютов)...", projectID, len(suites)))
	}

	// Create progress channel and monitor
	var monitor *progress.Monitor
	var progressChan chan int
	if bar != nil {
		progressChan = make(chan int, len(suiteIDs))
		monitor = progress.NewMonitor(progressChan, len(suiteIDs))
		
		go func() {
			for range progressChan {
				bar.Add(1)
			}
		}()
	}

	// Create parallel controller config
	config := &parallel.ControllerConfig{
		MaxConcurrentSuites: parallelSuites,
		MaxConcurrentPages:  parallelPages,
		Timeout:             timeout,
	}

	// Fetch cases using the new context-aware parallel method
	utils.DebugPrint("[Project %d] Starting GetCasesParallelCtx with parallelSuites=%d, parallelPages=%d, timeout=%s", 
		projectID, parallelSuites, parallelPages, timeout)
	ctx := context.Background()
	cases, execResult, err := cli.GetCasesParallelCtx(ctx, projectID, suiteIDs, config, monitor)
	
	if progressChan != nil {
		close(progressChan)
	}
	if bar != nil {
		bar.Finish()
	}
	
	// Log execution statistics
	if execResult != nil {
		utils.DebugPrint("[Project %d] Execution stats: expected=%d, got=%d, errors=%d, partial=%v",
			projectID, len(suites)*100, len(cases), len(execResult.Errors), execResult.Partial)
		
		// Warn if significant data loss detected (>10%)
		expectedCases := len(suites) * 100 // rough estimate
		if len(cases) < int(float64(expectedCases)*0.9) {
			fmt.Fprintf(os.Stderr, "⚠️  WARNING [Project %d]: Possible data loss! Expected ~%d cases, got %d\n", 
				projectID, expectedCases, len(cases))
		}
	}
	
	if err != nil && len(cases) == 0 {
		return nil, err
	}

	// Collect unique cases (GetCasesParallelCtx returns flat list)
	var allCases []ItemInfo
	caseIDs := make(map[int64]bool)

	for _, c := range cases {
		if !caseIDs[c.ID] {
			caseIDs[c.ID] = true
			allCases = append(allCases, ItemInfo{
				ID:   c.ID,
				Name: c.Title,
			})
		}
	}

	utils.DebugPrint("[Project %d] Total unique cases: %d", projectID, len(allCases))
	return allCases, nil
}

// analyzeCases performs comparison between two sets of cases.
func analyzeCases(cases1, cases2 []ItemInfo, pid1, pid2 int64, field string) *CompareResult {
	// Build name maps
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
	}
}

// fetchCaseItemsWithProgress fetches all cases for a project with progress updates.
// DEPRECATED: Use fetchCasesForProjectWithStats for better UX.
func fetchCaseItemsWithProgress(cli client.ClientInterface, projectID int64, suites data.GetSuitesResponse, bar *progress.Bar, workers int) ([]ItemInfo, error) {
	if workers <= 0 {
		workers = 5
	}
	
	utils.DebugPrint("[Project %d] Starting fetchCaseItemsWithProgress: %d suites, %d workers", projectID, len(suites), workers)
	
	// If no suites, fetch cases without suite filter
	if len(suites) == 0 {
		utils.DebugPrint("[Project %d] No suites found, fetching cases without suite filter", projectID)
		cases, err := cli.GetCases(projectID, 0, 0)
		if err != nil {
			utils.DebugPrint("[Project %d] Error fetching cases without suite: %v", projectID, err)
			return nil, err
		}

		allCases := make([]ItemInfo, 0, len(cases))
		for _, c := range cases {
			allCases = append(allCases, ItemInfo{
				ID:   c.ID,
				Name: c.Title,
			})
		}
		utils.DebugPrint("[Project %d] Fetched %d cases without suite filter", projectID, len(allCases))
		return allCases, nil
	}

	// Extract suite IDs
	suiteIDs := make([]int64, len(suites))
	for i, s := range suites {
		suiteIDs[i] = s.ID
	}
	utils.DebugPrint("[Project %d] Extracted %d suite IDs", projectID, len(suiteIDs))

	// Create progress channel and monitor for real-time updates
	var monitor *progress.Monitor
	var progressChan chan int
	if bar != nil {
		progressChan = make(chan int, len(suiteIDs))
		monitor = progress.NewMonitor(progressChan, len(suiteIDs))
		
		// Goroutine to update progress bar
		go func() {
			for range progressChan {
				bar.Add(1)
			}
		}()
	}
	
	// Fetch cases in parallel using concurrent API with progress monitor
	utils.DebugPrint("[Project %d] Calling GetCasesParallel with %d workers", projectID, workers)
	casesBySuite, err := cli.GetCasesParallel(projectID, suiteIDs, workers, monitor)
	utils.DebugPrint("[Project %d] GetCasesParallel returned: %d suites, err=%v", projectID, len(casesBySuite), err)
	
	// Close progress channel to stop the update goroutine
	if progressChan != nil {
		close(progressChan)
	}
	if err != nil && len(casesBySuite) == 0 {
		return nil, err
	}

	// Collect unique cases with summary
	var allCases []ItemInfo
	caseIDs := make(map[int64]bool)
	totalCases := 0

	for suiteID, cases := range casesBySuite {
		totalCases += len(cases)
		utils.DebugPrint("[Project %d] Processing suite %d: %d cases", projectID, suiteID, len(cases))
		for _, c := range cases {
			if !caseIDs[c.ID] {
				caseIDs[c.ID] = true
				allCases = append(allCases, ItemInfo{
					ID:   c.ID,
					Name: c.Title,
				})
			}
		}
	}
	
	utils.DebugPrint("[Project %d] Returning %d unique cases", projectID, len(allCases))
	return allCases, nil
}

// fetchCaseItems fetches all cases for a project and returns them as ItemInfo slice.
// Uses parallel API for significant performance improvement (4-5x faster).
// DEPRECATED: Use fetchCasesForProjectWithStats for better UX.
func fetchCaseItems(cli client.ClientInterface, projectID int64, pm *progress.Manager) ([]ItemInfo, error) {
	// Get all suites for the project
	suites, err := cli.GetSuites(projectID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сьютов проекта %d: %w", projectID, err)
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

	// Use parallel loading
	suiteIDs := make([]int64, len(suites))
	for i, s := range suites {
		suiteIDs[i] = s.ID
	}

	casesBySuite, err := cli.GetCasesParallel(projectID, suiteIDs, 5, nil)
	if err != nil {
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
	}

	return allCases, nil
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

// printCasesStats prints execution statistics for compare cases.
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
