package compare

import (
	"context"
	"errors"
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

type pushTaskStub struct {
	incremented int
	added       int
}

func (s *pushTaskStub) OnItemComplete()            {}
func (s *pushTaskStub) OnBatchReceived(n int)      {}
func (s *pushTaskStub) OnError()                   {}
func (s *pushTaskStub) OnPageFetched()             {}
func (s *pushTaskStub) Increment()                 { s.incremented++ }
func (s *pushTaskStub) Add(n int)                  { s.added += n }
func (s *pushTaskStub) Page()                      {}
func (s *pushTaskStub) Error(err error)            {}
func (s *pushTaskStub) Errors() int32              { return 0 }
func (s *pushTaskStub) Finish()                    {}
func (s *pushTaskStub) Elapsed() time.Duration     { return 0 }

func TestCoveragePush_ParseCommonFlags_InteractiveSavePromptError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "P10"}, {ID: 20, Name: "P20"}}, nil
		},
	}

	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 1}).
		WithConfirmResponses(true)

	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	_, _, _, _, err := parseCommonFlags(cmd, mock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "interactive save path selection failed")
}

func TestCoveragePush_SaveAllResult_UnsupportedFormat(t *testing.T) {
	err := saveAllResult(&allResult{}, "table", filepath.Join(t.TempDir(), "out.txt"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported for saving all resources")
}

func TestCoveragePush_FetchCasesForProject_ResultNil(t *testing.T) {
	mock := &client.MockClient{
		GetCasesParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			return data.GetCasesResponse{{ID: 1, Title: "A", SectionID: 7}}, nil, nil
		},
	}

	suites := data.GetSuitesResponse{{ID: 11, Name: "S1"}}
	items, failedPages, stats, err := fetchCasesForProject(context.Background(), mock, 1, suites, &pushTaskStub{}, 2, 2, time.Second, 0, 1)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Nil(t, failedPages)
	assert.Equal(t, 1, stats.CasesRaw)
	assert.Equal(t, 1, stats.CasesUnique)
	assert.Equal(t, 1, stats.Sections)
}

func TestCoveragePush_CompareSectionsInternalWithSuites_NilCmd(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	preloaded := map[int64]data.GetSuitesResponse{
		1: {{ID: 101, Name: "A"}},
		2: {{ID: 201, Name: "B"}},
	}

	mock := &client.MockClient{
		GetSectionsParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 1, SuiteID: 101, Name: "Sec"}}, nil
			}
			return data.GetSectionsResponse{{ID: 2, SuiteID: 201, Name: "Sec"}}, nil
		},
	}

	result, err := compareSectionsInternalWithSuites(context.Background(), nil, mock, 1, 2, true, preloaded)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, CompareStatusComplete, result.Status)
}

func TestCoveragePush_FetchSectionsForProject_EmptySuitesProgress(t *testing.T) {
	task := &pushTaskStub{}
	mock := &client.MockClient{
		GetSectionsParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 1, SuiteID: 0, Name: "Solo"}}, nil
		},
	}

	items, err := fetchSectionsForProject(context.Background(), mock, 1, data.GetSuitesResponse{}, task, compareHeavyRuntimeConfig{ParallelSuites: 1, ParallelPages: 1, PageRetries: 1, Timeout: time.Second})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 1, task.incremented)
	assert.Equal(t, 1, task.added)
}

func TestCoveragePush_FetchSectionsForProject_CtxErrPrecedence(t *testing.T) {
	task := &pushTaskStub{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock := &client.MockClient{
		GetSectionsParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
			return nil, errors.New("generic failure")
		},
	}

	_, err := fetchSectionsForProject(ctx, mock, 1, data.GetSuitesResponse{{ID: 1, Name: "S"}}, task, compareHeavyRuntimeConfig{ParallelSuites: 1, ParallelPages: 1, PageRetries: 1, Timeout: time.Second})
	require.ErrorIs(t, err, context.Canceled)
}

func TestCoveragePush_CompareSuitesInternalWithSuites_ContextInterrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	preloaded := map[int64]data.GetSuitesResponse{
		1: {{ID: 1, Name: "Suite"}},
		2: {{ID: 2, Name: "Suite"}},
	}

	result, err := compareSuitesInternalWithSuites(ctx, &client.MockClient{}, 1, 2, true, preloaded)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, CompareStatusInterrupted, result.Status)
}

func TestCoveragePush_ExecuteRetryFailedPages_SaveRemainingError(t *testing.T) {
	mockCli := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			return nil, errors.New("still failing")
		},
	}

	failed := []concurrency.FailedPage{{ProjectID: 1, SuiteID: 2, Offset: 0, Limit: 50, PageNum: 1, Error: "x"}}

	base := t.TempDir()
	fileAsDir := filepath.Join(base, "not-a-dir")
	require.NoError(t, os.WriteFile(fileAsDir, []byte("x"), 0o644))

	_, _, err := executeRetryFailedPages(context.Background(), mockCli, failed, retryFailedPagesOptions{Attempts: 1, Workers: 1, Delay: 0}, "src.json", filepath.Join(fileAsDir, "remaining.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save remaining failed pages error")
}

func TestCoveragePush_ResolveCompareHeavyRuntimeConfig_CloudTierFallback(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("compare.deployment", "cloud")
	viper.Set("compare.cloud_tier", "")
	viper.Set("compare.cloud_rate_limit", 0)
	viper.Set("compare.rate_limit", -1)

	cfg, err := resolveCompareHeavyRuntimeConfig(nil, "https://team.testrail.io")
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.RateLimit)
}
