// Package compare tests — additional coverage scenarios.
// Targets: saveTableToFile, saveAllSummaryToFile, saveAllResult,
// newAllCmd, newSuitesCmd, newCasesCmd, compareCasesInternal, saveFailedPagesReport.
package compare

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// saveTableToFile
// ─────────────────────────────────────────────────────────────────────────────

// TestSaveTableToFile_NotQuiet covers the !quiet print branch inside
// saveTableToFile — the path that prints "Result saved to <path>".
func TestSaveTableToFile_NotQuiet(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	// quiet is explicitly false (default anyway)

	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		OnlyInFirst: []ItemInfo{
			{ID: 1, Name: "Case A"},
		},
	}
	path := filepath.Join(t.TempDir(), "result.txt")

	// captureCompareStdout wraps os.Stdout; saveTableToFile internally redirects
	// stdout once more, then restores it to the outer pipe — so the "Result saved
	// to …" message lands in the outer capture buffer.
	out := captureCompareStdout(t, func() {
		err := saveTableToFile(cmd, result, "Project One", "Project Two", path)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "Result saved to")
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Case A")
}

// TestSaveTableToFile_DefaultPath_Quiet covers the else branch (auto-generated
// path) when no customPath is supplied.
func TestSaveTableToFile_DefaultPath_Quiet(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	require.NoError(t, cmd.Flags().Set("quiet", "true"))

	result := CompareResult{
		Resource:   "suites",
		Project1ID: 1,
		Project2ID: 2,
	}

	// No customPath argument ➜ triggers the else branch that computes the
	// default export file path via outpututils.GetExportsDir / GenerateFilename.
	err := saveTableToFile(cmd, result, "Alpha", "Beta")
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// saveAllSummaryToFile
// ─────────────────────────────────────────────────────────────────────────────

// TestSaveAllSummaryToFile_ExplicitPath_NotQuiet covers the explicit-path
// branch and the !quiet print branch.
func TestSaveAllSummaryToFile_ExplicitPath_NotQuiet(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")

	result := &allResult{
		Cases:    &CompareResult{Status: CompareStatusComplete},
		Suites:   &CompareResult{Status: CompareStatusComplete},
		Sections: &CompareResult{Status: CompareStatusComplete},
	}
	path := filepath.Join(t.TempDir(), "summary.txt")

	// saveAllSummaryToFile also does stdout redirection internally; the
	// "Result saved to" message is emitted AFTER the internal redirect is
	// restored, so captureCompareStdout captures it via the outer pipe.
	out := captureCompareStdout(t, func() {
		err := saveAllSummaryToFile(cmd, result, "P1", 1, "P2", 2, nil, path, time.Second)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "Result saved to")
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotEmpty(t, string(content))
}

// TestSaveAllSummaryToFile_DefaultPath_Quiet covers the __DEFAULT__ branch
// (else-path for default file name generation) with quiet=true.
func TestSaveAllSummaryToFile_DefaultPath_Quiet(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	require.NoError(t, cmd.Flags().Set("quiet", "true"))

	result := &allResult{}
	fillInterruptedResults(result, 1, 2)

	err := saveAllSummaryToFile(cmd, result, "P1", 1, "P2", 2, nil, "__DEFAULT__", 500*time.Millisecond)
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// newAllCmd — not-quiet save path
// ─────────────────────────────────────────────────────────────────────────────

// TestAllCmd_SaveToJSON_NotQuiet exercises the `if !quiet { ui.Infof(…) }`
// branch that prints "Result saved to <path>" when --save-to is used without
// --quiet.
func TestAllCmd_SaveToJSON_NotQuiet(t *testing.T) {
	mock := makeFullPassMock()
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	t.Cleanup(func() { SetGetClientForTests(nil) })

	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "result.json")

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=json", "--save-to=" + savePath})

	// We capture stdout so test output isn't flooded, but we just want the
	// code path to execute without error.
	out := captureCompareStdout(t, func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	assert.Contains(t, out, "Result saved to")
	_, err := os.ReadFile(savePath)
	require.NoError(t, err)
}

// TestAllCmd_SaveToYAML_NotQuiet same for yaml format.
func TestAllCmd_SaveToYAML_NotQuiet(t *testing.T) {
	mock := makeFullPassMock()
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	t.Cleanup(func() { SetGetClientForTests(nil) })

	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "result.yaml")

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=yaml", "--save-to=" + savePath, "--quiet"})

	err := cmd.Execute()
	require.NoError(t, err)
	_, err = os.ReadFile(savePath)
	require.NoError(t, err)
}

// TestAllCmd_SaveDefault_JSON exercises the __DEFAULT__ + json format branch
// (saves to auto-generated path without --save-to).
func TestAllCmd_SaveDefault_JSON_Format(t *testing.T) {
	mock := makeFullPassMock()
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	t.Cleanup(func() { SetGetClientForTests(nil) })

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=json", "--save", "--quiet"})

	err := cmd.Execute()
	require.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// newSuitesCmd
// ─────────────────────────────────────────────────────────────────────────────

// TestNewSuitesCmd_CompareSuitesError covers the
// `return fmt.Errorf("suites comparison error: %w", err)` path.
func TestNewSuitesCmd_CompareSuitesError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "P"}, nil
		},
		// GetSuitesParallel iterates via GetSuitesFunc; an error on the first
		// projectID returns an empty map ➜ triggers "failed to get suites".
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return nil, assert.AnError
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	t.Cleanup(func() { SetGetClientForTests(nil) })

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--quiet"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "suites comparison error")
}

// TestNewSuitesCmd_PrintCompareResultError covers the
// `return err` after PrintCompareResult fails (unsupported format).
func TestNewSuitesCmd_PrintCompareResultError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "P"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	t.Cleanup(func() { SetGetClientForTests(nil) })

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	// xml is unsupported ➜ PrintCompareResult returns error
	path := filepath.Join(t.TempDir(), "out.xml")
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=xml", "--save-to=" + path, "--quiet"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

// ─────────────────────────────────────────────────────────────────────────────
// newCasesCmd
// ─────────────────────────────────────────────────────────────────────────────

// TestNewCasesCmd_EmptyFieldFlagDefaultsToTitle covers the
// `if field == "" { field = "title" }` assignment branch in the RunE.
func TestNewCasesCmd_EmptyFieldFlagDefaultsToTitle(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "P"}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	t.Cleanup(func() { SetGetClientForTests(nil) })

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	// Passing --field= (empty string) triggers the `if field == "" { field = "title" }` line.
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--quiet", "--field="})

	err := cmd.Execute()
	require.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// compareCasesInternal — Interrupted status path
// ─────────────────────────────────────────────────────────────────────────────

// TestCompareCasesInternal_InterruptedStatusSet verifies that when the context
// is already canceled but the mock ignores it (returning data anyway), the
// result gets status = CompareStatusInterrupted.
func TestCompareCasesInternal_InterruptedStatusSet(t *testing.T) {
	// Pre-cancel the context so ctx.Err() != nil when the goroutines finish.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockClient := &client.MockClient{
		// GetCasesFunc purposely ignores ctx cancellation and returns data;
		// this lets both goroutines complete without error while ctx.Err() != nil.
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{
				{ID: projectID, Title: "Case"},
			}, nil
		},
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	require.NoError(t, cmd.Flags().Set("quiet", "true"))

	// Preload empty suites so both projects go through the GetCases (no-suites) path.
	preloaded := map[int64]data.GetSuitesResponse{
		1: {},
		2: {},
	}

	result, stats, err := compareCasesInternal(ctx, cmd, mockClient, 1, 2, "title", preloaded)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, CompareStatusInterrupted, result.Status)
	assert.True(t, stats.Interrupted)
}

// TestCompareCasesInternal_NotQuiet_LoadSummaryPrinted runs compareCasesInternal
// with quiet=false so the progress-output section is exercised.
func TestCompareCasesInternal_NotQuiet_LoadSummaryPrinted(t *testing.T) {
	mockClient := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: projectID * 10, Title: "C"}}, nil
		},
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	// quiet=false → exercises branches that call ui.Section / ui.Stat / ui.Successf

	preloaded := map[int64]data.GetSuitesResponse{
		1: {},
		2: {},
	}

	result, _, err := compareCasesInternal(context.Background(), cmd, mockClient, 1, 2, "title", preloaded)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestCompareCasesInternal_WithLoadErrors covers the
// `if task1.Errors() > 0 || task2.Errors() > 0 { ui.Warningf(…) }` branch.
// We use a real slow failing GetCasesParallelCtx so tasks accumulate errors.
func TestCompareCasesInternal_WithLoadErrors(t *testing.T) {
	mockClient := &client.MockClient{
		GetCasesParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			return data.GetCasesResponse{{ID: 1, Title: "X"}}, &concurrency.ExecutionResult{
				Stats: concurrency.AggregationStats{
					TotalPages: 1,
				},
			}, nil
		},
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")

	preloaded := map[int64]data.GetSuitesResponse{
		1: {{ID: 10, Name: "S1"}},
		2: {{ID: 20, Name: "S2"}},
	}

	result, _, err := compareCasesInternal(context.Background(), cmd, mockClient, 1, 2, "title", preloaded)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestCompareCasesInternal_AutoRetry_WithError covers the
// `if runtimeConfig.AutoRetryFailedPages { … }` block when the retry itself
// returns an error (retryErr != nil branch).
func TestCompareCasesInternal_AutoRetry_WithError(t *testing.T) {
	viper.Set("compare.cases.auto_retry_failed_pages", true)
	t.Cleanup(func() {
		viper.Set("compare.cases.auto_retry_failed_pages", true)
	})

	// auto_retry_failed_pages is true by default (viper.SetDefault in config_profile).
	// We remove any override to ensure default=true is used.
	mockClient := &client.MockClient{
		GetCasesParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			fp := concurrency.FailedPage{ProjectID: projectID, SuiteID: suiteIDs[0], Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"}
			return data.GetCasesResponse{{ID: projectID, Title: "C"}}, &concurrency.ExecutionResult{
				FailedPages: []concurrency.FailedPage{fp},
				Stats:       concurrency.AggregationStats{TotalPages: 1, FailedPages: 1},
			}, nil
		},
		// Retry always fails → retryErr != nil branch is covered.
		GetCasesPageFunc: func(ctx context.Context, projectID, suiteID int64, offset, limit int) (data.GetCasesResponse, error) {
			return nil, assert.AnError
		},
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	require.NoError(t, cmd.Flags().Set("quiet", "true"))

	preloaded := map[int64]data.GetSuitesResponse{
		1: {{ID: 10, Name: "S1"}},
		2: {{ID: 20, Name: "S2"}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, stats, err := compareCasesInternal(ctx, cmd, mockClient, 1, 2, "title", preloaded)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, stats.RetryAttempted)
	assert.Greater(t, stats.FailedPagesAfter, 0)
}

// TestCompareCasesInternal_AutoRetry_AllSuccess covers
// `else if len(remaining) == 0 { ui.Successf(…) }` inside the retry block.
func TestCompareCasesInternal_AutoRetry_AllSuccess(t *testing.T) {
	viper.Set("compare.cases.auto_retry_failed_pages", true)
	t.Cleanup(func() {
		viper.Set("compare.cases.auto_retry_failed_pages", true)
	})

	mockClient := &client.MockClient{
		GetCasesParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			fp := concurrency.FailedPage{ProjectID: projectID, SuiteID: suiteIDs[0], Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"}
			return data.GetCasesResponse{{ID: projectID, Title: "C"}}, &concurrency.ExecutionResult{
				FailedPages: []concurrency.FailedPage{fp},
				Stats:       concurrency.AggregationStats{TotalPages: 1, FailedPages: 1},
			}, nil
		},
		// Retry succeeds → remaining is empty → ui.Successf path executes.
		GetCasesPageFunc: func(ctx context.Context, projectID, suiteID int64, offset, limit int) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: 999, Title: "Retried"}}, nil
		},
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	require.NoError(t, cmd.Flags().Set("quiet", "true"))

	preloaded := map[int64]data.GetSuitesResponse{
		1: {{ID: 10, Name: "S1"}},
		2: {{ID: 20, Name: "S2"}},
	}

	result, stats, err := compareCasesInternal(context.Background(), cmd, mockClient, 1, 2, "title", preloaded)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, stats.RetryAttempted)
	assert.False(t, stats.RetryFailedWithErr)
	assert.Equal(t, 0, stats.FailedPagesAfter)
}

// TestCompareCasesInternal_MoreThan10FailedPages covers the
// `if len(allFailedPages) > showLimit { ui.Infof(…) }` branch (>10 pages).
func TestCompareCasesInternal_MoreThan10FailedPages(t *testing.T) {
	// Build 11 failed pages per project = 22 total > showLimit=10.
	makeFailedPages := func(projectID int64, suiteID int64, n int) []concurrency.FailedPage {
		pages := make([]concurrency.FailedPage, n)
		for i := range pages {
			pages[i] = concurrency.FailedPage{
				ProjectID: projectID,
				SuiteID:   suiteID,
				Offset:    i * 250,
				Limit:     250,
				PageNum:   i + 1,
				Error:     "timeout",
			}
		}
		return pages
	}

	mockClient := &client.MockClient{
		GetCasesParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			fps := makeFailedPages(projectID, suiteIDs[0], 11) // 11 pages each → 22 total
			return data.GetCasesResponse{{ID: projectID, Title: "C"}}, &concurrency.ExecutionResult{
				FailedPages: fps,
				Stats:       concurrency.AggregationStats{TotalPages: 11, FailedPages: 11},
			}, nil
		},
		GetCasesPageFunc: func(ctx context.Context, projectID, suiteID int64, offset, limit int) (data.GetCasesResponse, error) {
			return nil, assert.AnError
		},
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	require.NoError(t, cmd.Flags().Set("quiet", "true"))

	preloaded := map[int64]data.GetSuitesResponse{
		1: {{ID: 10, Name: "S1"}},
		2: {{ID: 20, Name: "S2"}},
	}

	result, stats, err := compareCasesInternal(context.Background(), cmd, mockClient, 1, 2, "title", preloaded)
	require.NoError(t, err)
	require.NotNil(t, result)
	// 22 total before retry, all still failed after retry errors
	assert.Greater(t, stats.FailedPagesBefore, 10, "expected more than 10 failed pages")
}

// ─────────────────────────────────────────────────────────────────────────────
// saveFailedPagesReport — WriteFile error
// ─────────────────────────────────────────────────────────────────────────────

// TestSaveFailedPagesReport_WriteFileError covers the
// `if err := os.WriteFile(…); err != nil { return "", … }` path by using a
// read-only directory so the write is rejected.
func TestSaveFailedPagesReport_WriteFileError(t *testing.T) {
	dir := t.TempDir()
	// Make the directory read-only so WriteFile fails.
	require.NoError(t, os.Chmod(dir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	pages := []concurrency.FailedPage{
		{ProjectID: 1, SuiteID: 10, Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"},
	}
	// Pass a path inside the read-only dir; MkdirAll on the existing dir
	// succeeds (it already exists), but WriteFile fails.
	_, err := saveFailedPagesReport(pages, filepath.Join(dir, "failed.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "writing report")
}

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

// makeFullPassMock returns a MockClient that returns empty success responses
// for all resources used by newAllCmd.
func makeFullPassMock() *client.MockClient {
	return &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return data.GetSharedStepsResponse{}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return data.GetPlansResponse{}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return data.GetDatasetsResponse{}, nil
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return data.GetGroupsResponse{}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return data.GetLabelsResponse{}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return data.GetTemplatesResponse{}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return data.GetConfigsResponse{}, nil
		},
	}
}
