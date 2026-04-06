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
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeepCoverage_AllCmd_SaveToYamlByExtensionWithTableFormat(t *testing.T) {
	mock := makeFullPassMock()
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	t.Cleanup(func() {
		SetGetClientForTests(nil)
	})

	savePath := filepath.Join(t.TempDir(), "result.yaml")

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=table", "--save-to=" + savePath, "--quiet"})

	err := cmd.Execute()
	require.NoError(t, err)

	content, readErr := os.ReadFile(savePath)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "meta:")
}

func TestDeepCoverage_ParseCommonFlags_SaveToEmptySkipsPrompt(t *testing.T) {
	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	require.NoError(t, cmd.Flags().Set("pid1", "1"))
	require.NoError(t, cmd.Flags().Set("pid2", "2"))
	require.NoError(t, cmd.Flags().Set("save-to", ""))

	pid1, pid2, format, savePath, err := parseCommonFlags(cmd, &client.MockClient{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pid1)
	assert.Equal(t, int64(2), pid2)
	assert.Equal(t, "table", format)
	assert.Equal(t, "", savePath)
}

func TestDeepCoverage_CompareSectionsInternalWithSuites_PartialSuitesMapOnError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesParallelFunc: func(ctx context.Context, projectIDs []int64, workers int) (map[int64]data.GetSuitesResponse, error) {
			return map[int64]data.GetSuitesResponse{
				1: {{ID: 101, Name: "S1"}},
				2: {{ID: 201, Name: "S2"}},
			}, errors.New("partial suites fetch")
		},
		GetSectionsParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 1, SuiteID: 101, Name: "A"}}, nil
			}
			return data.GetSectionsResponse{{ID: 2, SuiteID: 201, Name: "B"}}, nil
		},
	}

	result, err := compareSectionsInternalWithSuites(context.Background(), nil, mock, 1, 2, true, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, CompareStatusComplete, result.Status)
}

func TestDeepCoverage_CompareSectionsInternalWithSuites_ContextDeadlineExceeded(t *testing.T) {
	preloaded := map[int64]data.GetSuitesResponse{
		1: {{ID: 101, Name: "S1"}},
		2: {{ID: 201, Name: "S2"}},
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	mock := &client.MockClient{
		GetSectionsParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: projectID, SuiteID: suiteIDs[0], Name: "A"}}, nil
		},
	}

	result, err := compareSectionsInternalWithSuites(ctx, &cobra.Command{}, mock, 1, 2, true, preloaded)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestDeepCoverage_FetchSectionsForProject_DedupAndSuiteNameFallback(t *testing.T) {
	task := &pushTaskStub{}
	suites := data.GetSuitesResponse{{ID: 11, Name: "Suite X"}}

	mock := &client.MockClient{
		GetSectionsParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{
				{ID: 1, SuiteID: 11, Name: "Section A"},
				{ID: 1, SuiteID: 11, Name: "Section A duplicate"},
				{ID: 2, SuiteID: 999, Name: "Orphan Section"},
			}, nil
		},
	}

	items, err := fetchSectionsForProject(context.Background(), mock, 1, suites, task, compareHeavyRuntimeConfig{ParallelSuites: 1, ParallelPages: 1, PageRetries: 1, Timeout: time.Second})
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "Suite X / Section A", items[0].Name)
	assert.Equal(t, "Orphan Section", items[1].Name)
}

func TestDeepCoverage_FetchSectionsForProject_GenericErrorReturned(t *testing.T) {
	task := &pushTaskStub{}
	mock := &client.MockClient{
		GetSectionsParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
			return nil, errors.New("sections backend failed")
		},
	}

	_, err := fetchSectionsForProject(context.Background(), mock, 1, data.GetSuitesResponse{{ID: 11, Name: "S"}}, task, compareHeavyRuntimeConfig{ParallelSuites: 1, ParallelPages: 1, PageRetries: 1, Timeout: time.Second})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sections backend failed")
}

func TestDeepCoverage_CompareCasesInternal_ReportSaveErrorDoesNotBreakRun(t *testing.T) {
	viper.Set("compare.cases.auto_retry_failed_pages", false)
	t.Cleanup(func() {
		viper.Set("compare.cases.auto_retry_failed_pages", true)
	})

	homeAsFile := filepath.Join(t.TempDir(), "home-file")
	require.NoError(t, os.WriteFile(homeAsFile, []byte("x"), 0o644))
	t.Setenv("HOME", homeAsFile)

	preloaded := map[int64]data.GetSuitesResponse{
		1: {{ID: 10, Name: "S1"}},
		2: {{ID: 20, Name: "S2"}},
	}

	mockClient := &client.MockClient{
		GetCasesParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			return data.GetCasesResponse{{ID: projectID, Title: "Case"}}, &concurrency.ExecutionResult{
				FailedPages: []concurrency.FailedPage{{ProjectID: projectID, SuiteID: suiteIDs[0], Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"}},
				Stats:       concurrency.AggregationStats{TotalPages: 1, FailedPages: 1},
			}, nil
		},
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", true, "")

	result, stats, err := compareCasesInternal(context.Background(), cmd, mockClient, 1, 2, "title", preloaded)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, CompareStatusPartial, result.Status)
	assert.Equal(t, 2, stats.FailedPagesBefore)
	assert.Equal(t, 2, stats.FailedPagesAfter)
	assert.Equal(t, "", stats.FailedPagesReport)
}

func TestDeepCoverage_SaveFailedPagesReport_CurrentDirectoryPath(t *testing.T) {
	t.Chdir(t.TempDir())

	pages := []concurrency.FailedPage{{ProjectID: 1, SuiteID: 2, Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"}}
	savedPath, err := saveFailedPagesReport(pages, "failed_pages_local.json")
	require.NoError(t, err)
	assert.Equal(t, "failed_pages_local.json", savedPath)

	content, readErr := os.ReadFile(savedPath)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "failed_pages")
}
