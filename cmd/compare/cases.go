package compare

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/debug"
	"github.com/Korrnals/gotr/internal/models/data"
	outpututils "github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/Korrnals/gotr/pkg/reporter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	compareCasesMarshalIndent = json.MarshalIndent
	compareCasesWriteFile     = os.WriteFile
)

// casesCmd is the exported command.
var casesCmd = newCasesCmd()

// projectDataStats contains structural statistics about a single project.
type projectDataStats struct {
	Suites            int           // number of suites (from GetSuites API)
	Sections          int           // unique sections (from SectionID in cases)
	CasesRaw          int           // total raw cases before deduplication
	CasesUnique       int           // unique cases (after ID-dedup)
	CasesExpected     int           // expected cases per API (sum of totalSize across all suites; -1 if unknown)
	SuitesWithTotal   int           // suites that reported totalSize
	SuitesVerified    int           // suites with confirmed completeness (all pages loaded, clean exhaustion)
	SuiteDetailsSum   int           // sum of cases across all suites (for integrity check)
	SuiteDetailsEmpty int           // number of empty suites (0 cases)
	SuiteDetailsCount int           // number of tracked suites
	TotalPages        int           // total number of requested pages
	FailedPages       int           // pages with errors
	UniqueTitles      int           // unique titles (for verification)
	EmptyTitles       int           // cases without a title
	Elapsed           time.Duration // loading time for this project
}

type casesExecutionStats struct {
	Project1           projectDataStats
	Project2           projectDataStats
	LoadErrorsP1       int
	LoadErrorsP2       int
	FailedPagesBefore  int
	RetryStats         retryFailedPagesStats
	FailedPagesAfter   int
	FailedPagesReport  string
	RetryAttempted     bool
	RetryFailedWithErr bool
	Interrupted        bool // context was canceled before completion (Ctrl+C or deadline)
}

// newCasesCmd creates the 'compare cases' subcommand.
func newCasesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cases",
		Short: "Compare test cases between projects",
		Long: `Compares test cases between two projects.

By default, comparison uses the 'title' field.
You can specify another field using the --field flag.

Supported fields:
  title, priority_id, type_id, milestone_id, refs, 
	custom_preconds, custom_steps, custom_expected, and more.

Examples:
	# Compare cases by title
  gotr compare cases --pid1 30 --pid2 31

	# Compare by priority
  gotr compare cases --pid1 30 --pid2 31 --field priority_id

	# Save result to the default file
  gotr compare cases --pid1 30 --pid2 31 --save

	# Save result to a specific file
  gotr compare cases --pid1 30 --pid2 31 --save-to cases_diff.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClientSafe(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			// Parse flags
			pid1, pid2, format, savePath, err := parseCommonFlags(cmd, cli)
			if err != nil {
				return err
			}

			// Get field for comparison
			field, _ := cmd.Flags().GetString("field")
			if field == "" {
				field = "title"
			}

			// Get project names
			project1Name, project2Name, err := GetProjectNames(ctx, cli, pid1, pid2)
			if err != nil {
				return err
			}

			// Start timer
			startTime := time.Now()

			// Execute comparison
			result, execStats, err := compareCasesInternal(ctx, cmd, cli, pid1, pid2, field)
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
				PrintCasesStatsWithErrors(
					pid1,
					pid2,
					len(result.OnlyInFirst),
					len(result.OnlyInSecond),
					len(result.Common),
					elapsed,
					execStats,
				)
			}

			return nil
		},
	}

	// Add flags
	addCommonFlags(cmd)
	cmd.Flags().String("field", "title", "Field to compare by (title, priority_id, etc.)")

	return cmd
}

// getCaseKey extracts the comparison key from a case based on field name.
func getCaseKey(item ItemInfo, field string) string {
	if field == "title" {
		return item.Name
	}
	return item.Name
}

// compareCasesInternal compares cases between two projects and returns the result.
// Uses ui.Display for live progress — no mpb, no progress.Monitor.
func compareCasesInternal(ctx context.Context, cmd *cobra.Command, cli client.ClientInterface, pid1, pid2 int64, field string, preloadedSuites ...map[int64]data.GetSuitesResponse) (*CompareResult, casesExecutionStats, error) {
	execStats := casesExecutionStats{}
	quiet, _ := cmd.Flags().GetBool("quiet")
	operation := ui.NewOperation(ui.StatusConfig{
		Title:  "Loading data",
		Writer: os.Stderr,
		Quiet:  quiet,
	})
	defer operation.Finish()

	// Phase 1: Get suites for both projects (quick operation)
	debug.DebugPrint("[Compare] Phase 1: Fetching suites for projects %d and %d", pid1, pid2)
	operation.Info("Loading project structure for %d and %d...", pid1, pid2)

	var suitesMap map[int64]data.GetSuitesResponse
	var err error
	if len(preloadedSuites) > 0 && preloadedSuites[0] != nil {
		suitesMap = preloadedSuites[0]
	} else {
		suitesMap, err = cli.GetSuitesParallel(ctx, []int64{pid1, pid2}, 2, nil)
		if err != nil && len(suitesMap) == 0 {
			return nil, execStats, fmt.Errorf("failed to get suites: %w", err)
		}
	}

	suites1 := suitesMap[pid1]
	suites2 := suitesMap[pid2]

	debug.DebugPrint("[Compare] Found suites: P%d=%d, P%d=%d", pid1, len(suites1), pid2, len(suites2))

	// Phase 2: Parallel loading of both projects with live display
	debug.DebugPrint("[Compare] Phase 2: Parallel loading of projects %d and %d", pid1, pid2)

	runtimeConfig, err := resolveCompareCasesRuntimeConfig(collectCompareCasesFlagOverrides(cmd), viper.GetString("base_url"))
	if err != nil {
		return nil, execStats, err
	}

	debug.DebugPrint("[Compare] RuntimeConfig: parallelSuites=%d, parallelPages=%d, rateLimit=%d, pageRetries=%d, retryAttempts=%d, retryWorkers=%d, retryDelay=%s",
		runtimeConfig.ParallelSuites, runtimeConfig.ParallelPages, runtimeConfig.RateLimit,
		runtimeConfig.PageRetries, runtimeConfig.RetryAttempts, runtimeConfig.RetryWorkers, runtimeConfig.RetryDelay)

	// Create live operation tasks
	task1 := operation.AddTask(fmt.Sprintf("P%d (%d suites)", pid1, len(suites1)), len(suites1))
	task2 := operation.AddTask(fmt.Sprintf("P%d (%d suites)", pid2, len(suites2)), len(suites2))

	var cases1, cases2 []ItemInfo
	var failedPages1, failedPages2 []concurrency.FailedPage
	var stats1, stats2 projectDataStats
	var err1, err2 error

	done1 := make(chan struct{})
	done2 := make(chan struct{})

	// Load Project 1
	go func() {
		cases1, failedPages1, stats1, err1 = fetchCasesForProject(ctx, cli, pid1, suites1, task1, runtimeConfig.ParallelSuites, runtimeConfig.ParallelPages, runtimeConfig.Timeout, runtimeConfig.RateLimit, runtimeConfig.PageRetries)
		task1.Finish()
		close(done1)
	}()

	// Load Project 2
	go func() {
		cases2, failedPages2, stats2, err2 = fetchCasesForProject(ctx, cli, pid2, suites2, task2, runtimeConfig.ParallelSuites, runtimeConfig.ParallelPages, runtimeConfig.Timeout, runtimeConfig.RateLimit, runtimeConfig.PageRetries)
		task2.Finish()
		close(done2)
	}()

	// Wait for both
	<-done1
	<-done2

	execStats.Project1 = stats1
	execStats.Project2 = stats2

	// Detect Ctrl+C or deadline — even when partial data was returned without error.
	execStats.Interrupted = ctx.Err() != nil

	// Print summary
	if !quiet {
		ui.Section(os.Stderr, "Loading summary")
		ui.Stat(os.Stderr, "📦", fmt.Sprintf("Project %d", pid1),
			fmt.Sprintf("%d cases in %s", len(cases1), task1.Elapsed().Round(time.Second)))
		ui.Stat(os.Stderr, "📦", fmt.Sprintf("Project %d", pid2),
			fmt.Sprintf("%d cases in %s", len(cases2), task2.Elapsed().Round(time.Second)))
	}

	if task1.Errors() > 0 || task2.Errors() > 0 {
		ui.Warningf(os.Stderr, "Errors: P%d=%d, P%d=%d", pid1, task1.Errors(), pid2, task2.Errors())
	}
	execStats.LoadErrorsP1 = int(task1.Errors())
	execStats.LoadErrorsP2 = int(task2.Errors())

	// Merge failed pages from both projects and delegate reporting/retry
	// to handleFailedPages (extracted to reduce cyclomatic complexity).
	allFailedPages := append(append([]concurrency.FailedPage{}, failedPages1...), failedPages2...)
	execStats.FailedPagesBefore = len(allFailedPages)
	handleFailedPages(ctx, cli, allFailedPages, runtimeConfig, operation, &execStats)

	if err1 != nil {
		return nil, execStats, fmt.Errorf("failed to load project %d: %w", pid1, err1)
	}
	if err2 != nil {
		return nil, execStats, fmt.Errorf("failed to load project %d: %w", pid2, err2)
	}

	// Phase 3: Analysis
	debug.DebugPrint("[Compare] Phase 3: Analysis and comparison")
	operation.Phase("Analyzing and comparing data...")

	start := time.Now()
	result := analyzeCases(cases1, cases2, pid1, pid2, field)
	elapsed := time.Since(start)

	if !quiet {
		ui.Successf(os.Stderr, "Analysis completed (%s)", elapsed.Round(time.Millisecond))
	}
	debug.DebugPrint("[Compare] Analysis complete: P%d=%d unique, P%d=%d unique, common=%d",
		pid1, len(result.OnlyInFirst), pid2, len(result.OnlyInSecond), len(result.Common))

	// Tag result for JSON output and stats banner.
	switch {
	case execStats.Interrupted:
		result.Status = CompareStatusInterrupted
	case execStats.FailedPagesAfter > 0:
		result.Status = CompareStatusPartial
	default:
		result.Status = CompareStatusComplete
	}

	return result, execStats, nil
}

// handleFailedPages reports, saves, and optionally auto-retries failed pages.
// Extracted from compareCasesInternal to keep the main orchestration function lean.
// Three-phase logic:
//  1. Display up to 10 failed pages as a human-readable summary.
//  2. Persist the full list to a JSON report for later manual retry.
//  3. If auto-retry is enabled in runtimeConfig, run executeRetryFailedPages
//     and record the outcome in execStats.
func handleFailedPages(ctx context.Context, cli client.ClientInterface, allFailedPages []concurrency.FailedPage, runtimeConfig compareCasesRuntimeConfig, operation ui.Operation, execStats *casesExecutionStats) {
	if len(allFailedPages) == 0 {
		return
	}

	// Display a truncated summary (up to 10 pages) so the user sees
	// which pages were not fetched without flooding the terminal.
	ui.Warningf(os.Stderr, "Unfetched pages after retry/recovery: %d", len(allFailedPages))
	showLimit := 10
	if len(allFailedPages) < showLimit {
		showLimit = len(allFailedPages)
	}
	for i := 0; i < showLimit; i++ {
		fp := allFailedPages[i]
		ui.Infof(os.Stderr, "  - project=%d suite=%d page=%d offset=%d limit=%d", fp.ProjectID, fp.SuiteID, fp.PageNum, fp.Offset, fp.Limit)
	}
	if len(allFailedPages) > showLimit {
		ui.Infof(os.Stderr, "  ... and %d more pages (see JSON report)", len(allFailedPages)-showLimit)
	}

	// Save full failed-pages list to JSON for post-hoc `gotr compare retry-failed-pages`.
	reportPath, saveErr := saveFailedPagesReport(allFailedPages, "")
	if saveErr != nil {
		ui.Warningf(os.Stderr, "Failed to save failed-pages report: %v", saveErr)
	} else {
		ui.Infof(os.Stderr, "Failed-pages report saved: %s", reportPath)
		execStats.FailedPagesReport = reportPath
	}

	// Auto-retry: when enabled, immediately re-fetch failed pages using the
	// worker pool in executeRetryFailedPages. The result updates execStats
	// so the caller can set the correct CompareStatus (partial/complete).
	if runtimeConfig.AutoRetryFailedPages {
		operation.Phase("Running auto-retry for failed pages...")
		execStats.RetryAttempted = true
		remaining, retryStats, retryErr := executeRetryFailedPages(
			ctx,
			cli,
			allFailedPages,
			retryFailedPagesOptions{
				Attempts: runtimeConfig.RetryAttempts,
				Workers:  runtimeConfig.RetryWorkers,
				Delay:    runtimeConfig.RetryDelay,
			},
			"auto-retry after compare cases",
			"",
		)
		execStats.RetryStats = retryStats
		execStats.FailedPagesAfter = len(remaining)
		if retryErr != nil {
			execStats.RetryFailedWithErr = true
			ui.Warningf(os.Stderr, "Auto-retry finished with error: %v", retryErr)
		} else if len(remaining) == 0 {
			ui.Successf(os.Stderr, "Auto-retry: all failed pages were processed successfully")
		}
	} else {
		execStats.FailedPagesAfter = len(allFailedPages)
		ui.Warningf(os.Stderr, "Auto-retry is disabled via compare.cases.auto_retry_failed_pages")
	}
}

// fetchCasesForProject loads all cases for a single project.
// task is a *ui.Task implementing concurrency.PaginatedProgressReporter — gets live updates.
func fetchCasesForProject(ctx context.Context, cli client.ClientInterface, projectID int64, suites data.GetSuitesResponse, task ui.TaskHandle, parallelSuites, parallelPages int, timeout time.Duration, rateLimit int, pageRetries int) ([]ItemInfo, []concurrency.FailedPage, projectDataStats, error) {
	fetchStart := time.Now()
	pds := projectDataStats{Suites: len(suites)}

	if len(suites) == 0 {
		debug.DebugPrint("[Project %d] No suites, fetching all cases", projectID)
		cases, err := cli.GetCases(ctx, projectID, 0, 0)
		if err != nil {
			return nil, nil, pds, err
		}

		allCases := make([]ItemInfo, 0, len(cases))
		sectionIDs := make(map[int64]struct{})
		for _, c := range cases {
			allCases = append(allCases, ItemInfo{
				ID:   c.ID,
				Name: c.Title,
			})
			if c.SectionID != 0 {
				sectionIDs[c.SectionID] = struct{}{}
			}
		}
		pds.CasesRaw = len(cases)
		pds.CasesUnique = len(allCases)
		pds.Sections = len(sectionIDs)
		pds.Elapsed = time.Since(fetchStart)
		return allCases, nil, pds, nil
	}

	// Extract suite IDs
	suiteIDs := make([]int64, len(suites))
	for i, s := range suites {
		suiteIDs[i] = s.ID
	}

	// Create parallel controller config with Reporter = task (ui.Task implements PaginatedProgressReporter)
	config := &concurrency.ControllerConfig{
		MaxConcurrentSuites: parallelSuites,
		MaxConcurrentPages:  parallelPages,
		RequestsPerMinute:   rateLimit,
		MaxRetries:          pageRetries,
		Timeout:             timeout,
		Reporter:            task, // *ui.Task → concurrency.PaginatedProgressReporter
	}

	debug.DebugPrint("[Project %d] Starting GetCasesParallelCtx (streaming) with parallelSuites=%d, parallelPages=%d, timeout=%s",
		projectID, parallelSuites, parallelPages, timeout)
	cases, result, err := cli.GetCasesParallelCtx(ctx, projectID, suiteIDs, config)

	if err != nil && len(cases) == 0 {
		return nil, nil, pds, err
	}

	// Populate per-suite diagnostics from the parallel execution result.
	// Extracted into collectParallelFetchStats to reduce the CC of fetchCasesForProject.
	if result != nil {
		collectParallelFetchStats(result, projectID, &pds)
	}

	// Remove duplicate case IDs (suites may overlap) and gather title/section stats.
	// Extracted into deduplicateCases to keep fetchCasesForProject focused on orchestration.
	allCases, sectionCount, emptyTitles, uniqueTitles := deduplicateCases(cases)

	pds.CasesRaw = len(cases)
	pds.CasesUnique = len(allCases)
	pds.UniqueTitles = uniqueTitles
	pds.EmptyTitles = emptyTitles
	pds.Sections = sectionCount
	pds.Elapsed = time.Since(fetchStart)

	debug.DebugPrint("[Project %d] Total: %d raw → %d unique IDs → %d unique titles (empty titles: %d), %d sections",
		projectID, len(cases), len(allCases), uniqueTitles, emptyTitles, sectionCount)

	if result != nil {
		return allCases, result.FailedPages, pds, nil
	}

	return allCases, nil, pds, nil
}

// collectParallelFetchStats fills projectDataStats from a parallel execution result.
// It walks SuiteResults to compute per-suite aggregates (sum, empty count)
// and populates page/failure counters used in the loading summary banner.
func collectParallelFetchStats(result *concurrency.ExecutionResult, projectID int64, pds *projectDataStats) {
	stats := result.Stats
	debug.DebugPrint("[Project %d] Fetch stats: %d suites completed, %d pages, %d raw cases, expected=%d, partial=%v",
		projectID, stats.CompletedSuites, stats.TotalPages, stats.TotalCases, stats.ExpectedCases, result.Partial)
	pds.CasesExpected = int(stats.ExpectedCases)
	pds.SuitesWithTotal = stats.SuitesWithTotal
	pds.SuitesVerified = stats.SuitesVerified
	if len(stats.SuiteResults) > 0 {
		sum := 0
		emptySuites := 0
		for _, r := range stats.SuiteResults {
			sum += r.CasesFetched
			if r.CasesFetched == 0 {
				emptySuites++
			}
			verified := "✗"
			if r.Verified {
				verified = "✓"
			}
			debug.DebugPrint("[Project %d] Suite %d: %d cases [%s]",
				projectID, r.SuiteID, r.CasesFetched, verified)
		}
		debug.DebugPrint("[Project %d] Suite totals: Σ=%d, empty=%d, count=%d",
			projectID, sum, emptySuites, len(stats.SuiteResults))
		pds.SuiteDetailsSum = sum
		pds.SuiteDetailsEmpty = emptySuites
		pds.SuiteDetailsCount = len(stats.SuiteResults)
	}
	pds.TotalPages = stats.TotalPages
	pds.FailedPages = stats.FailedPages
}

// deduplicateCases removes duplicate cases by ID and counts sections/titles.
// When cases are fetched from multiple suites, IDs can repeat; this function
// produces a clean unique slice and computes three auxiliary metrics:
//   - sectionCount: distinct section IDs encountered (for integrity checks)
//   - emptyTitles: cases with blank titles (data quality indicator)
//   - uniqueTitles: case-insensitive distinct titles (for title-based comparison)
func deduplicateCases(cases data.GetCasesResponse) (items []ItemInfo, sectionCount, emptyTitles, uniqueTitles int) {
	caseIDs := make(map[int64]bool)
	sectionIDs := make(map[int64]struct{})

	for _, c := range cases {
		if !caseIDs[c.ID] {
			caseIDs[c.ID] = true
			if c.Title == "" {
				emptyTitles++
			}
			items = append(items, ItemInfo{
				ID:   c.ID,
				Name: c.Title,
			})
		}
		if c.SectionID != 0 {
			sectionIDs[c.SectionID] = struct{}{}
		}
	}

	titleSet := make(map[string]struct{})
	for _, item := range items {
		if item.Name != "" {
			titleSet[strings.ToLower(item.Name)] = struct{}{}
		}
	}

	return items, len(sectionIDs), emptyTitles, len(titleSet)
}

func collectCompareHeavyFlagOverrides(cmd *cobra.Command) map[string]any {
	overrides := map[string]any{}

	intFlags := []struct{ flag, key string }{
		{"rate-limit", "rate_limit"},
		{"parallel-suites", "parallel_suites"},
		{"parallel-pages", "parallel_pages"},
		{"page-retries", "page_retries"},
		{"retry-attempts", "retry_attempts"},
		{"retry-workers", "retry_workers"},
	}
	for _, f := range intFlags {
		if flag := cmd.Flags().Lookup(f.flag); flag != nil && flag.Changed {
			value, _ := cmd.Flags().GetInt(f.flag)
			overrides[f.key] = value
		}
	}

	durationFlags := []struct{ flag, key string }{
		{"timeout", "timeout"},
		{"retry-delay", "retry_delay"},
	}
	for _, f := range durationFlags {
		if flag := cmd.Flags().Lookup(f.flag); flag != nil && flag.Changed {
			value, _ := cmd.Flags().GetDuration(f.flag)
			overrides[f.key] = value
		}
	}

	return overrides
}

func collectCompareCasesFlagOverrides(cmd *cobra.Command) map[string]any {
	return collectCompareHeavyFlagOverrides(cmd)
}

func saveFailedPagesReport(failedPages []concurrency.FailedPage, requestedPath string) (string, error) {
	if len(failedPages) == 0 {
		return "", nil
	}

	path := strings.TrimSpace(requestedPath)
	if path == "" {
		exportsDir, _ := outpututils.GetExportsDir("compare")
		if err := os.MkdirAll(exportsDir, 0o755); err != nil {
			return "", fmt.Errorf("creating reports directory: %w", err)
		}
		path = filepath.Join(exportsDir, fmt.Sprintf("failed_pages_%s.json", time.Now().Format("2006-01-02_15-04-05")))
	} else {
		dir := filepath.Dir(path)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return "", fmt.Errorf("creating directory %s: %w", dir, err)
			}
		}
	}

	payload := struct {
		GeneratedAt string                   `json:"generated_at"`
		Total       int                      `json:"total"`
		FailedPages []concurrency.FailedPage `json:"failed_pages"`
	}{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Total:       len(failedPages),
		FailedPages: failedPages,
	}

	raw, err := compareCasesMarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal failed pages: %w", err)
	}

	if err := compareCasesWriteFile(path, raw, 0o644); err != nil {
		return "", fmt.Errorf("writing report %s: %w", path, err)
	}

	return path, nil
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
		Status:       CompareStatusComplete,
		OnlyInFirst:  onlyInFirst,
		OnlyInSecond: onlyInSecond,
		Common:       common,
	}
}

// printCasesFieldDiff prints differences by field for cases.
func printCasesFieldDiff(ctx context.Context, cli client.ClientInterface, pid1, pid2 int64, field string) {
	diff, err := cli.DiffCasesData(ctx, pid1, pid2, field)
	if err != nil {
		ui.Warningf(os.Stdout, "Error getting differences for field '%s': %v", field, err)
		return
	}

	if len(diff.DiffByField) == 0 {
		ui.Infof(os.Stdout, "No differences found for field '%s'", field)
		return
	}

	ui.Section(os.Stdout, fmt.Sprintf("Differences for field '%s'", field))
	for _, d := range diff.DiffByField {
		firstValue := getFieldValue(d.First, field)
		secondValue := getFieldValue(d.Second, field)

		ui.Infof(os.Stdout, "Case: %s (ID: %d)", d.First.Title, d.CaseID)
		ui.Infof(os.Stdout, "Project %d: %s", pid1, firstValue)
		ui.Infof(os.Stdout, "Project %d: %s", pid2, secondValue)
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

	r := reporter.New("cases").
		Section("General statistics").
		Stat("⏱️", "Execution time", elapsed.Round(time.Millisecond)).
		Stat("📦", "Total cases processed", totalCases).
		Section("Comparison results").
		Stat("🔹", fmt.Sprintf("Only in project %d", result.Project1ID), len(result.OnlyInFirst)).
		Stat("🔹", fmt.Sprintf("Only in project %d", result.Project2ID), len(result.OnlyInSecond)).
		Stat("🔗", "Common cases", len(result.Common))

	r.Print()
}
