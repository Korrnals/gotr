package compare

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newSectionsCmd creates the 'compare sections' subcommand.
func newSectionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sections",
		Short: "Compare sections between projects",
		Long: `Compares sections between two projects.

Examples:
	# Compare sections
  gotr compare sections --pid1 30 --pid2 31

	# Save result to the default file
  gotr compare sections --pid1 30 --pid2 31 --save

	# Save result to a specific file
  gotr compare sections --pid1 30 --pid2 31 --save-to sections_diff.json
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

			// Get project names
			project1Name, project2Name, err := GetProjectNames(ctx, cli, pid1, pid2)
			if err != nil {
				return err
			}

			// Start timer
			startTime := time.Now()

			// Compare sections
			quiet, _ := cmd.Flags().GetBool("quiet")
			result, err := compareSectionsInternal(ctx, cmd, cli, pid1, pid2, quiet)
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
					len(result.OnlyInFirst), len(result.OnlyInSecond), len(result.Common), elapsed, result.Status)
			}

			return nil
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// sectionsCmd is the exported command.
var sectionsCmd = newSectionsCmd()

// compareSectionsInternal compares sections between two projects using client adapter path.
func compareSectionsInternal(ctx context.Context, cmd *cobra.Command, cli client.ClientInterface, pid1, pid2 int64, quiet bool) (*CompareResult, error) {
	return compareSectionsInternalWithSuites(ctx, cmd, cli, pid1, pid2, quiet, nil)
}

func compareSectionsInternalWithSuites(ctx context.Context, cmd *cobra.Command, cli client.ClientInterface, pid1, pid2 int64, quiet bool, preloaded map[int64]data.GetSuitesResponse) (*CompareResult, error) {
	operation := ui.NewOperation(ui.StatusConfig{
		Title:  "Loading sections",
		Writer: os.Stderr,
		Quiet:  quiet,
	})
	defer operation.Finish()

	operation.Info("Loading project structure for %d and %d...", pid1, pid2)

	suitesMap := preloaded
	var err error
	if suitesMap == nil {
		suitesMap, err = cli.GetSuitesParallel(ctx, []int64{pid1, pid2}, 2, nil)
		if err != nil && len(suitesMap) == 0 {
			return nil, fmt.Errorf("failed to get suites: %w", err)
		}
	}

	suites1 := suitesMap[pid1]
	suites2 := suitesMap[pid2]

	var runtimeConfig compareHeavyRuntimeConfig
	if cmd != nil {
		runtimeConfig, err = resolveCompareHeavyRuntimeConfig(collectCompareHeavyFlagOverrides(cmd), viper.GetString("base_url"))
	} else {
		runtimeConfig, err = resolveCompareHeavyRuntimeConfig(nil, viper.GetString("base_url"))
	}
	if err != nil {
		return nil, err
	}

	task1 := operation.AddTask(fmt.Sprintf("P%d (%d suites)", pid1, len(suites1)), taskTotal(len(suites1)))
	task2 := operation.AddTask(fmt.Sprintf("P%d (%d suites)", pid2, len(suites2)), taskTotal(len(suites2)))

	var items1, items2 []ItemInfo
	var err1, err2 error

	done1 := make(chan struct{})
	done2 := make(chan struct{})

	go func() {
		items1, err1 = fetchSectionsForProject(ctx, cli, pid1, suites1, task1, runtimeConfig)
		task1.Finish()
		close(done1)
	}()

	go func() {
		items2, err2 = fetchSectionsForProject(ctx, cli, pid2, suites2, task2, runtimeConfig)
		task2.Finish()
		close(done2)
	}()

	<-done1
	<-done2

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
		ui.Section(os.Stderr, "Loading summary")
		ui.Stat(os.Stderr, "📦", fmt.Sprintf("Project %d", pid1),
			fmt.Sprintf("%d sections in %s", len(items1), task1.Elapsed().Round(time.Second)))
		ui.Stat(os.Stderr, "📦", fmt.Sprintf("Project %d", pid2),
			fmt.Sprintf("%d sections in %s", len(items2), task2.Elapsed().Round(time.Second)))
	}

	return compareItemInfos("sections", pid1, pid2, items1, items2), nil
}

func fetchSectionsForProject(ctx context.Context, cli client.ClientInterface, projectID int64, suites data.GetSuitesResponse, task ui.TaskHandle, runtimeConfig compareHeavyRuntimeConfig) ([]ItemInfo, error) {
	suiteIDs := make([]int64, 0, len(suites))
	suiteNames := make(map[int64]string, len(suites))
	for _, suite := range suites {
		suiteIDs = append(suiteIDs, suite.ID)
		suiteNames[suite.ID] = suite.Name
	}

	controllerConfig := &concurrency.ControllerConfig{
		MaxConcurrentSuites: runtimeConfig.ParallelSuites,
		MaxConcurrentPages:  runtimeConfig.ParallelPages,
		RequestsPerMinute:   runtimeConfig.RateLimit,
		MaxRetries:          runtimeConfig.PageRetries,
		Timeout:             runtimeConfig.Timeout,
		Reporter:            task,
	}

	sections, err := cli.GetSectionsParallelCtx(ctx, projectID, suiteIDs, controllerConfig)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			return nil, err
		}
		return nil, err
	}

	if len(suites) == 0 {
		task.Increment()
		task.Add(len(sections))
	}

	seen := make(map[int64]bool, len(sections))
	result := make([]ItemInfo, 0, len(sections))
	for _, section := range sections {
		if seen[section.ID] {
			continue
		}
		seen[section.ID] = true

		name := section.Name
		if suiteName := suiteNames[section.SuiteID]; suiteName != "" {
			name = fmt.Sprintf("%s / %s", suiteName, section.Name)
		}

		result = append(result, ItemInfo{ID: section.ID, Name: name})
	}

	return result, nil
}

func taskTotal(n int) int {
	if n == 0 {
		return 1
	}
	return n
}
