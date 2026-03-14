package compare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newSectionsCmd creates the 'compare sections' subcommand.
func newSectionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sections",
		Short: "Сравнить секции между проектами",
		Long: `Выполняет сравнение секций (разделов) между двумя проектами.

Примеры:
  # Сравнить секции
  gotr compare sections --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare sections --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare sections --pid1 30 --pid2 31 --save-to sections_diff.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClientSafe(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			// Parse flags
			pid1, pid2, format, savePath, err := parseCommonFlags(cmd)
			if err != nil {
				return err
			}

			// Get project names
			project1Name, project2Name, err := GetProjectNames(ctx, cli, pid1, pid2)
			if err != nil {
				return err
			}

			// Start timer
			startTime := time.Now()

			// Compare sections
			quiet, _ := cmd.Flags().GetBool("quiet")
			result, err := compareSectionsInternal(ctx, cli, pid1, pid2, quiet)
			if err != nil {
				return err
			}

			elapsed := time.Since(startTime)

			// Print or save result
			if err := PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath); err != nil {
				return err
			}

			// Print statistics
			if !quiet {
				PrintCompareStats("sections", pid1, pid2,
					len(result.OnlyInFirst), len(result.OnlyInSecond), len(result.Common), elapsed)
			}

			return nil
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// sectionsCmd — экспортированная команда
var sectionsCmd = newSectionsCmd()

// compareSectionsInternal compares sections between two projects using FetchParallel.
func compareSectionsInternal(ctx context.Context, cli client.ClientInterface, pid1, pid2 int64, quiet bool) (*CompareResult, error) {
	ui.Infof(os.Stderr, "Получение структуры проектов %d и %d...", pid1, pid2)

	suitesMap, err := cli.GetSuitesParallel(ctx, []int64{pid1, pid2}, 2, nil)
	if err != nil && len(suitesMap) == 0 {
		return nil, fmt.Errorf("failed to get suites: %w", err)
	}

	suites1 := suitesMap[pid1]
	suites2 := suitesMap[pid2]

	display := ui.New(ui.WithQuiet(quiet))
	display.SetHeader("Загрузка sections")
	task1 := display.AddTask(fmt.Sprintf("П%d (%d сьютов)", pid1, len(suites1)), taskTotal(len(suites1)))
	task2 := display.AddTask(fmt.Sprintf("П%d (%d сьютов)", pid2, len(suites2)), taskTotal(len(suites2)))

	var items1, items2 []ItemInfo
	var err1, err2 error

	done1 := make(chan struct{})
	done2 := make(chan struct{})

	go func() {
		items1, err1 = fetchSectionsForProject(ctx, cli, pid1, suites1, task1)
		task1.Finish()
		close(done1)
	}()

	go func() {
		items2, err2 = fetchSectionsForProject(ctx, cli, pid2, suites2, task2)
		task2.Finish()
		close(done2)
	}()

	<-done1
	<-done2
	display.Finish()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if err1 != nil {
		if errors.Is(err1, context.Canceled) || errors.Is(err1, context.DeadlineExceeded) {
			return nil, err1
		}
		return nil, fmt.Errorf("failed to load project %d sections: %w", pid1, err1)
	}
	if err2 != nil {
		if errors.Is(err2, context.Canceled) || errors.Is(err2, context.DeadlineExceeded) {
			return nil, err2
		}
		return nil, fmt.Errorf("failed to load project %d sections: %w", pid2, err2)
	}

	if !quiet {
		ui.Section(os.Stderr, "Результаты загрузки")
		ui.Stat(os.Stderr, "📦", fmt.Sprintf("Project %d", pid1),
			fmt.Sprintf("%d sections за %s", len(items1), task1.Elapsed().Round(time.Second)))
		ui.Stat(os.Stderr, "📦", fmt.Sprintf("Project %d", pid2),
			fmt.Sprintf("%d sections за %s", len(items2), task2.Elapsed().Round(time.Second)))
	}

	return compareItemInfos("sections", pid1, pid2, items1, items2), nil
}

// fetchSectionItems fetches all sections for a project, parallelizing across suites
// using FetchParallelBySuite.
func fetchSectionItems(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	// Get all suites for the project
	suites, err := cli.GetSuites(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// If no suites, try without suite filter
	if len(suites) == 0 {
		sections, err := cli.GetSections(ctx, projectID, 0)
		if err != nil {
			return nil, err
		}
		items := make([]ItemInfo, 0, len(sections))
		for _, s := range sections {
			items = append(items, ItemInfo{ID: s.ID, Name: s.Name})
		}
		return items, nil
	}

	// Build suite ID list and name map
	suiteIDs := make([]int64, len(suites))
	suiteNames := make(map[int64]string)
	for i, s := range suites {
		suiteIDs[i] = s.ID
		suiteNames[s.ID] = s.Name
	}

	// Fetch sections from all suites in parallel
	allSections, err := concurrency.FetchParallelBySuite(ctx, suiteIDs,
		func(suiteID int64) ([]ItemInfo, error) {
			sections, sErr := cli.GetSections(ctx, projectID, suiteID)
			if sErr != nil {
				return nil, sErr
			}
			items := make([]ItemInfo, 0, len(sections))
			suiteName := suiteNames[suiteID]
			for _, s := range sections {
				name := s.Name
				if suiteName != "" {
					name = fmt.Sprintf("%s / %s", suiteName, s.Name)
				}
				items = append(items, ItemInfo{ID: s.ID, Name: name})
			}
			return items, nil
		},
		concurrency.WithContinueOnError(),
	)
	if err != nil {
		return nil, err
	}

	// Deduplicate by ID
	seen := make(map[int64]bool)
	result := make([]ItemInfo, 0, len(allSections))
	for _, item := range allSections {
		if !seen[item.ID] {
			seen[item.ID] = true
			result = append(result, item)
		}
	}

	return result, nil
}

func fetchSectionsForProject(ctx context.Context, cli client.ClientInterface, projectID int64, suites data.GetSuitesResponse, task *ui.Task) ([]ItemInfo, error) {
	if len(suites) == 0 {
		sections, err := cli.GetSections(ctx, projectID, 0)
		if err != nil {
			return nil, err
		}

		items := make([]ItemInfo, 0, len(sections))
		for _, section := range sections {
			items = append(items, ItemInfo{ID: section.ID, Name: section.Name})
		}

		task.OnItemComplete()
		task.OnBatchReceived(len(items))
		return items, nil
	}

	suiteIDs := make([]int64, len(suites))
	suiteNames := make(map[int64]string, len(suites))
	for i, suite := range suites {
		suiteIDs[i] = suite.ID
		suiteNames[suite.ID] = suite.Name
	}

	if httpClient, ok := cli.(*client.HTTPClient); ok {
		allSections, err := concurrency.FetchParallelBySuite(ctx, suiteIDs,
			func(suiteID int64) ([]ItemInfo, error) {
				return fetchSectionsForSuitePaged(ctx, httpClient, projectID, suiteID, suiteNames[suiteID], task)
			},
			concurrency.WithContinueOnError(),
			concurrency.WithReporter(task),
			concurrency.WithMaxConcurrency(10),
		)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, err
		}

		seen := make(map[int64]bool, len(allSections))
		result := make([]ItemInfo, 0, len(allSections))
		for _, item := range allSections {
			if seen[item.ID] {
				continue
			}
			seen[item.ID] = true
			result = append(result, item)
		}

		return result, nil
	}

	allSections, err := concurrency.FetchParallelBySuite(ctx, suiteIDs,
		func(suiteID int64) ([]ItemInfo, error) {
			sections, fetchErr := cli.GetSections(ctx, projectID, suiteID)
			if fetchErr != nil {
				return nil, fetchErr
			}

			items := make([]ItemInfo, 0, len(sections))
			suiteName := suiteNames[suiteID]
			for _, section := range sections {
				name := section.Name
				if suiteName != "" {
					name = fmt.Sprintf("%s / %s", suiteName, section.Name)
				}
				items = append(items, ItemInfo{ID: section.ID, Name: name})
			}
			return items, nil
		},
		concurrency.WithContinueOnError(),
		concurrency.WithReporter(task),
		concurrency.WithMaxConcurrency(10),
	)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	seen := make(map[int64]bool, len(allSections))
	result := make([]ItemInfo, 0, len(allSections))
	for _, item := range allSections {
		if seen[item.ID] {
			continue
		}
		seen[item.ID] = true
		result = append(result, item)
	}

	return result, nil
}

func fetchSectionsForSuitePaged(ctx context.Context, cli *client.HTTPClient, projectID, suiteID int64, suiteName string, task *ui.Task) ([]ItemInfo, error) {
	const pageLimit = 250

	offset := 0
	allItems := make([]ItemInfo, 0)

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		query := map[string]string{
			"suite_id": strconv.FormatInt(suiteID, 10),
			"offset":   strconv.Itoa(offset),
			"limit":    strconv.Itoa(pageLimit),
		}

		resp, err := cli.Get(ctx, fmt.Sprintf("get_sections/%d", projectID), query)
		if err != nil {
			return nil, err
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read sections page body (project=%d suite=%d offset=%d): %w", projectID, suiteID, offset, readErr)
		}

		sections, pageLen, decodeErr := decodeSectionsPage(body)
		if decodeErr != nil {
			return nil, fmt.Errorf("decode sections page (project=%d suite=%d offset=%d): %w", projectID, suiteID, offset, decodeErr)
		}

		items := make([]ItemInfo, 0, len(sections))
		for _, section := range sections {
			name := section.Name
			if suiteName != "" {
				name = fmt.Sprintf("%s / %s", suiteName, section.Name)
			}
			items = append(items, ItemInfo{ID: section.ID, Name: name})
		}

		if len(items) > 0 {
			allItems = append(allItems, items...)
			task.OnBatchReceived(len(items))
		}
		task.OnPageFetched()

		if pageLen < pageLimit {
			break
		}
		offset += pageLimit
	}

	return allItems, nil
}

func decodeSectionsPage(body []byte) ([]data.Section, int, error) {
	if len(body) == 0 {
		return nil, 0, nil
	}

	for _, b := range body {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			var wrapper struct {
				Sections []data.Section `json:"sections"`
			}
			if err := json.Unmarshal(body, &wrapper); err != nil {
				return nil, 0, err
			}
			return wrapper.Sections, len(wrapper.Sections), nil
		case '[':
			var sections []data.Section
			if err := json.Unmarshal(body, &sections); err != nil {
				return nil, 0, err
			}
			return sections, len(sections), nil
		default:
			return nil, 0, fmt.Errorf("unexpected response format starting with %q", strings.TrimSpace(string([]byte{b})))
		}
	}

	return nil, 0, nil
}

func taskTotal(n int) int {
	if n == 0 {
		return 1
	}
	return n
}
