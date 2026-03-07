package compare

import (
	"context"
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
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
				return fmt.Errorf("HTTP клиент не инициализирован")
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
			result, err := compareSectionsInternal(cli, pid1, pid2)
			if err != nil {
				return fmt.Errorf("ошибка сравнения секций: %w", err)
			}

			elapsed := time.Since(startTime)

			// Print or save result
			if err := PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath); err != nil {
				return err
			}

			// Print statistics
			quiet, _ := cmd.Flags().GetBool("quiet")
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
func compareSectionsInternal(cli client.ClientInterface, pid1, pid2 int64) (*CompareResult, error) {
	return compareSimpleInternal(cli, pid1, pid2, "sections", fetchSectionItems)
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
	ctx = context.Background()
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
