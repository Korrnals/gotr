package compare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingCSVWriter struct {
	failsOn int
	calls   int
}

func (w *failingCSVWriter) Write(_ []string) error {
	w.calls++
	if w.calls == w.failsOn {
		return errors.New("forced csv write error")
	}
	return nil
}

func (w *failingCSVWriter) Flush() {}

func TestCoverageHooks_SaveAllSummaryToFile_PipeError(t *testing.T) {
	oldPipe := compareAllPipe
	compareAllPipe = func() (*os.File, *os.File, error) {
		return nil, nil, errors.New("forced pipe error")
	}
	t.Cleanup(func() { compareAllPipe = oldPipe })

	err := saveAllSummaryToFile(&cobra.Command{}, &allResult{}, "P1", 1, "P2", 2, nil, filepath.Join(t.TempDir(), "x.txt"), time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipe create error")
}

func TestCoverageHooks_SaveAllSummaryToFile_CopyError(t *testing.T) {
	oldCopy := compareAllCopy
	compareAllCopy = func(io.Writer, io.Reader) (int64, error) {
		return 0, errors.New("forced copy error")
	}
	t.Cleanup(func() { compareAllCopy = oldCopy })

	err := saveAllSummaryToFile(&cobra.Command{}, &allResult{}, "P1", 1, "P2", 2, nil, filepath.Join(t.TempDir(), "x.txt"), time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "output read error")
}

func TestCoverageHooks_SaveTableToFile_PipeError(t *testing.T) {
	oldPipe := compareTypesPipe
	compareTypesPipe = func() (*os.File, *os.File, error) {
		return nil, nil, errors.New("forced pipe error")
	}
	t.Cleanup(func() { compareTypesPipe = oldPipe })

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", true, "")
	err := saveTableToFile(cmd, CompareResult{}, "P1", "P2", filepath.Join(t.TempDir(), "x.txt"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipe create error")
}

func TestCoverageHooks_SaveTableToFile_CopyError(t *testing.T) {
	oldCopy := compareTypesCopy
	compareTypesCopy = func(io.Writer, io.Reader) (int64, error) {
		return 0, errors.New("forced copy error")
	}
	t.Cleanup(func() { compareTypesCopy = oldCopy })

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", true, "")
	err := saveTableToFile(cmd, CompareResult{}, "P1", "P2", filepath.Join(t.TempDir(), "x.txt"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "output read error")
}

func TestCoverageHooks_SaveTableToFile_PrintError(t *testing.T) {
	oldPrint := compareTypesPrint
	compareTypesPrint = func(result CompareResult, project1Name, project2Name string) error {
		return errors.New("forced table print error")
	}
	t.Cleanup(func() { compareTypesPrint = oldPrint })

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", true, "")
	err := saveTableToFile(cmd, CompareResult{}, "P1", "P2", filepath.Join(t.TempDir(), "x.txt"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forced table print error")
}

func TestCoverageHooks_PrintCSV_HeaderWriteError(t *testing.T) {
	oldFactory := newCompareCSVWriter
	newCompareCSVWriter = func(io.Writer) compareCSVWriter {
		return &failingCSVWriter{failsOn: 1}
	}
	t.Cleanup(func() { newCompareCSVWriter = oldFactory })

	err := printCSV(CompareResult{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forced csv write error")
}

func TestCoverageHooks_SaveCSV_HeaderWriteError(t *testing.T) {
	oldFactory := newCompareCSVWriter
	newCompareCSVWriter = func(io.Writer) compareCSVWriter {
		return &failingCSVWriter{failsOn: 1}
	}
	t.Cleanup(func() { newCompareCSVWriter = oldFactory })

	err := saveCSV(CompareResult{}, filepath.Join(t.TempDir(), "out.csv"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forced csv write error")
}

func TestCoverageHooks_SaveFailedPagesReport_MarshalError(t *testing.T) {
	oldMarshal := compareCasesMarshalIndent
	compareCasesMarshalIndent = func(v any, prefix, indent string) ([]byte, error) {
		return nil, errors.New("forced marshal error")
	}
	t.Cleanup(func() { compareCasesMarshalIndent = oldMarshal })

	_, err := saveFailedPagesReport([]concurrency.FailedPage{{ProjectID: 1, SuiteID: 1, Offset: 0, Limit: 250, PageNum: 1}}, filepath.Join(t.TempDir(), "failed.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal failed pages")
}

func TestCoverageHooks_ParseCommonFlags_SaveToHasPriorityOverSave(t *testing.T) {
	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	require.NoError(t, cmd.Flags().Set("pid1", "1"))
	require.NoError(t, cmd.Flags().Set("pid2", "2"))
	require.NoError(t, cmd.Flags().Set("save", "true"))
	require.NoError(t, cmd.Flags().Set("save-to", "out.json"))

	pid1, pid2, format, savePath, err := parseCommonFlags(cmd, &client.MockClient{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pid1)
	assert.Equal(t, int64(2), pid2)
	assert.Equal(t, "table", format)
	assert.Equal(t, "out.json", savePath)
}

func TestCoverageHooks_ParseCommonFlags_InteractiveSavePromptSuccess(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "A"}, {ID: 20, Name: "B"}}, nil
		},
	}

	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 1}).
		WithConfirmResponses(true, false).
		WithInputResponses("my-report.json")

	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	cmd.SetContext(interactive.WithPrompter(context.Background(), mp))

	pid1, pid2, format, savePath, err := parseCommonFlags(cmd, mock)
	require.NoError(t, err)
	assert.Equal(t, int64(10), pid1)
	assert.Equal(t, int64(20), pid2)
	assert.Equal(t, "table", format)
	assert.Equal(t, "my-report.json", savePath)
}

func TestCoverageHooks_ParseCommonFlags_InteractiveSavePromptError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "A"}, {ID: 20, Name: "B"}}, nil
		},
	}

	// No input response is provided intentionally to force prompt input error.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 1}).
		WithConfirmResponses(true, false)

	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	cmd.SetContext(interactive.WithPrompter(context.Background(), mp))

	_, _, _, _, err := parseCommonFlags(cmd, mock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestCoverageHooks_ParseCommonFlags_ResolveSavePathError(t *testing.T) {
	oldResolve := resolveSavePathFn
	resolveSavePathFn = func(cmd *cobra.Command) (string, bool, error) {
		return "", false, errors.New("forced resolve save error")
	}
	t.Cleanup(func() { resolveSavePathFn = oldResolve })

	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	require.NoError(t, cmd.Flags().Set("pid1", "1"))
	require.NoError(t, cmd.Flags().Set("pid2", "2"))

	_, _, _, _, err := parseCommonFlags(cmd, &client.MockClient{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forced resolve save error")
}

func TestCoverageHooks_AllCmd_InterruptedAtEachStage(t *testing.T) {
	stages := []string{
		"cases",
		"suites",
		"sections",
		"sharedsteps",
		"runs",
		"plans",
		"milestones",
		"datasets",
		"groups",
		"labels",
		"templates",
		"configurations",
	}

	baseMock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: fmt.Sprintf("P%d", projectID)}, nil
		},
		GetSuitesParallelFunc: func(ctx context.Context, projectIDs []int64, workers int) (map[int64]data.GetSuitesResponse, error) {
			return map[int64]data.GetSuitesResponse{1: {}, 2: {}}, nil
		},
	}

	for _, stopStage := range stages {
		stopStage := stopStage
		t.Run(stopStage, func(t *testing.T) {
			oldCases := compareAllCasesFn
			oldSuites := compareAllSuitesFn
			oldSections := compareAllSectionsFn
			oldSimple := compareAllSimpleFn
			t.Cleanup(func() {
				compareAllCasesFn = oldCases
				compareAllSuitesFn = oldSuites
				compareAllSectionsFn = oldSections
				compareAllSimpleFn = oldSimple
			})

			okResult := &CompareResult{Resource: "ok", Project1ID: 1, Project2ID: 2, Status: CompareStatusComplete}
			interruptedErr := fmt.Errorf("wrapped cancel: %w", context.Canceled)

			compareAllCasesFn = func(ctx context.Context, cmd *cobra.Command, cli client.ClientInterface, pid1, pid2 int64, field string, preloadedSuites ...map[int64]data.GetSuitesResponse) (*CompareResult, casesExecutionStats, error) {
				if stopStage == "cases" {
					return nil, casesExecutionStats{}, interruptedErr
				}
				return okResult, casesExecutionStats{}, nil
			}
			compareAllSuitesFn = func(ctx context.Context, cli client.ClientInterface, pid1, pid2 int64, quiet bool, preloaded map[int64]data.GetSuitesResponse) (*CompareResult, error) {
				if stopStage == "suites" {
					return nil, interruptedErr
				}
				return okResult, nil
			}
			compareAllSectionsFn = func(ctx context.Context, cmd *cobra.Command, cli client.ClientInterface, pid1, pid2 int64, quiet bool, preloaded map[int64]data.GetSuitesResponse) (*CompareResult, error) {
				if stopStage == "sections" {
					return nil, interruptedErr
				}
				return okResult, nil
			}
			compareAllSimpleFn = func(ctx context.Context, cli client.ClientInterface, pid1, pid2 int64, resource string, fetcher FetchFunc, quiet ...bool) (*CompareResult, error) {
				if stopStage == resource {
					return nil, interruptedErr
				}
				return okResult, nil
			}

			SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface { return baseMock })
			t.Cleanup(func() { SetGetClientForTests(nil) })

			cmd := newAllCmd()
			addPersistentFlagsForTests(cmd)
			cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--quiet"})

			err := cmd.Execute()
			require.NoError(t, err)
		})
	}
}

func TestCoverageHooks_ExecuteRetryFailedPages_NegativeDelayNormalized(t *testing.T) {
	mockCli := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: 1, Title: "ok"}}, nil
		},
	}

	failed := []concurrency.FailedPage{{ProjectID: 1, SuiteID: 2, Offset: 0, Limit: 250, PageNum: 1, Error: "x"}}
	remaining, stats, err := executeRetryFailedPages(context.Background(), mockCli, failed, retryFailedPagesOptions{Attempts: 1, Workers: 1, Delay: -time.Second}, "src", "")
	require.NoError(t, err)
	assert.Nil(t, remaining)
	assert.Equal(t, 1, stats.RecoveredPages)
}

func TestCoverageHooks_ResolveCompareCasesRuntimeConfig_AutoRetryExplicitTrue(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("compare.cases.auto_retry_failed_pages", true)

	cfg, err := resolveCompareCasesRuntimeConfig(nil, "https://team.testrail.io")
	require.NoError(t, err)
	assert.True(t, cfg.AutoRetryFailedPages)
}

func TestCoverageHooks_ExecuteRetryFailedPages_SortByLimitTiebreak(t *testing.T) {
	mockCli := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			return nil, errors.New("still failing")
		},
	}

	failed := []concurrency.FailedPage{
		{ProjectID: 1, SuiteID: 10, Offset: 0, Limit: 200, PageNum: 1, Error: "x"},
		{ProjectID: 1, SuiteID: 10, Offset: 0, Limit: 100, PageNum: 1, Error: "x"},
	}

	remaining, stats, err := executeRetryFailedPages(context.Background(), mockCli, failed, retryFailedPagesOptions{Attempts: 1, Workers: 1, Delay: 0}, "src", filepath.Join(t.TempDir(), "remaining.json"))
	require.NoError(t, err)
	require.Len(t, remaining, 2)
	assert.Equal(t, 100, remaining[0].Limit)
	assert.Equal(t, 200, remaining[1].Limit)
	assert.Equal(t, 2, stats.RemainingPages)
}

func TestCoverageHooks_CompareSectionsInternalWithSuites_Project1Canceled(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	preloaded := map[int64]data.GetSuitesResponse{1: {{ID: 101, Name: "S1"}}, 2: {{ID: 201, Name: "S2"}}}
	mock := &client.MockClient{
		GetSectionsParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return nil, context.Canceled
			}
			return data.GetSectionsResponse{}, nil
		},
	}

	result, err := compareSectionsInternalWithSuites(context.Background(), nil, mock, 1, 2, true, preloaded)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCoverageHooks_CompareSectionsInternalWithSuites_Project2Deadline(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	preloaded := map[int64]data.GetSuitesResponse{1: {{ID: 101, Name: "S1"}}, 2: {{ID: 201, Name: "S2"}}}
	mock := &client.MockClient{
		GetSectionsParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
			if projectID == 2 {
				return nil, context.DeadlineExceeded
			}
			return data.GetSectionsResponse{}, nil
		},
	}

	result, err := compareSectionsInternalWithSuites(context.Background(), nil, mock, 1, 2, true, preloaded)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestCoverageHooks_FetchCasesForProject_EmptySuiteResultStat(t *testing.T) {
	task := &pushTaskStub{}
	mock := &client.MockClient{
		GetCasesParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			return data.GetCasesResponse{{ID: 1, Title: "A"}}, &concurrency.ExecutionResult{
				Stats: concurrency.AggregationStats{
					SuiteResults: []concurrency.SuiteResultInfo{{SuiteID: 10, CasesFetched: 0, Verified: false}},
				},
			}, nil
		},
	}

	items, failed, pds, err := fetchCasesForProject(context.Background(), mock, 1, data.GetSuitesResponse{{ID: 10, Name: "S"}}, task, 1, 1, time.Second, 0, 0)
	require.NoError(t, err)
	assert.Empty(t, failed)
	assert.Len(t, items, 1)
	assert.Equal(t, 1, pds.SuiteDetailsEmpty)
}

func TestCoverageHooks_CompareCasesInternal_TaskErrorsAndRetryError(t *testing.T) {
	viper.Set("compare.cases.auto_retry_failed_pages", true)
	t.Cleanup(func() {
		viper.Set("compare.cases.auto_retry_failed_pages", true)
	})

	// Make exports path creation fail during auto-retry save.
	homeAsFile := filepath.Join(t.TempDir(), "home-file")
	require.NoError(t, os.WriteFile(homeAsFile, []byte("x"), 0o644))
	t.Setenv("HOME", homeAsFile)

	preloaded := map[int64]data.GetSuitesResponse{1: {{ID: 10, Name: "S1"}}, 2: {{ID: 20, Name: "S2"}}}
	mockClient := &client.MockClient{
		GetCasesParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			if cfg != nil && cfg.Reporter != nil {
				cfg.Reporter.OnError()
			}
			return data.GetCasesResponse{{ID: projectID, Title: "Case"}}, &concurrency.ExecutionResult{
				FailedPages: []concurrency.FailedPage{{ProjectID: projectID, SuiteID: suiteIDs[0], Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"}},
				Stats:       concurrency.AggregationStats{TotalPages: 1, FailedPages: 1},
			}, nil
		},
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			return nil, errors.New("still failing")
		},
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", true, "")

	result, stats, err := compareCasesInternal(context.Background(), cmd, mockClient, 1, 2, "title", preloaded)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, CompareStatusPartial, result.Status)
	assert.GreaterOrEqual(t, stats.LoadErrorsP1, 1)
	assert.GreaterOrEqual(t, stats.LoadErrorsP2, 1)
	assert.True(t, stats.RetryAttempted)
	assert.True(t, stats.RetryFailedWithErr)
}

func TestCoverageHooks_CompareSectionsInternalWithSuites_InvalidRuntimeConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("compare.cases.timeout", "bad-duration")

	preloaded := map[int64]data.GetSuitesResponse{1: {}, 2: {}}
	_, err := compareSectionsInternalWithSuites(context.Background(), &cobra.Command{}, &client.MockClient{}, 1, 2, true, preloaded)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid compare.cases.timeout")
}

func TestCoverageHooks_AllCmd_UnsupportedSaveFormat(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "P"}, nil
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
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) { return []data.Milestone{}, nil },
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
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface { return mock })
	t.Cleanup(func() { SetGetClientForTests(nil) })

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=xml", "--save-to=result.bin", "--quiet"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format 'xml'")
}

func TestCoverageHooks_AllCmd_ProjectNameError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return nil, fmt.Errorf("boom %d", projectID)
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface { return mock })
	t.Cleanup(func() { SetGetClientForTests(nil) })

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--quiet"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get project")
}
