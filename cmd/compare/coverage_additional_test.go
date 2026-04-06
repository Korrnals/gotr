package compare

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCompareCasesRuntimeConfig_FallbackForNonPositiveValues(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("compare.deployment", "")
	viper.Set("compare.cases.parallel_suites", 0)
	viper.Set("compare.cases.parallel_pages", -1)
	viper.Set("compare.cases.page_retries", 0)
	viper.Set("compare.cases.retry.attempts", 0)
	viper.Set("compare.cases.retry.workers", -2)
	viper.Set("compare.cases.timeout", "")
	viper.Set("compare.cases.retry.delay", "")
	viper.Set("compare.cloud_rate_limit", 245)

	cfg, err := resolveCompareCasesRuntimeConfig(nil, "https://example.testrail.io")
	require.NoError(t, err)

	assert.Equal(t, 245, cfg.RateLimit)
	assert.Equal(t, 12, cfg.ParallelSuites)
	assert.Equal(t, 8, cfg.ParallelPages)
	assert.Equal(t, 5, cfg.PageRetries)
	assert.Equal(t, 30*time.Minute, cfg.Timeout)
	assert.Equal(t, 5, cfg.RetryAttempts)
	assert.Equal(t, 12, cfg.RetryWorkers)
	assert.Equal(t, 500*time.Millisecond, cfg.RetryDelay)
	assert.True(t, cfg.AutoRetryFailedPages)
}

func TestFetchConfigurationItems_FlattensGroupAndConfigs(t *testing.T) {
	mock := &client.MockClient{
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			assert.Equal(t, int64(42), projectID)
			return data.GetConfigsResponse{{
				ID:   10,
				Name: "Browsers",
				Configs: []data.Config{
					{ID: 100, Name: "Chrome"},
					{ID: 101, Name: "Firefox"},
				},
			}}, nil
		},
	}

	items, err := fetchConfigurationItems(context.Background(), mock, 42)
	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Equal(t, ItemInfo{ID: 10, Name: "Browsers"}, items[0])
	assert.Equal(t, ItemInfo{ID: 100, Name: "Browsers / Chrome"}, items[1])
	assert.Equal(t, ItemInfo{ID: 101, Name: "Browsers / Firefox"}, items[2])
}

func TestExecuteRetryFailedPages_ManyRemaining_AreSorted(t *testing.T) {
	ctx := context.Background()
	mockCli := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			return nil, fmt.Errorf("forced retry failure")
		},
	}

	failedPages := []concurrency.FailedPage{
		{ProjectID: 2, SuiteID: 2, Offset: 200, Limit: 50, PageNum: 5},
		{ProjectID: 1, SuiteID: 4, Offset: 0, Limit: 50, PageNum: 1},
		{ProjectID: 1, SuiteID: 3, Offset: 100, Limit: 50, PageNum: 3},
		{ProjectID: 1, SuiteID: 3, Offset: 0, Limit: 50, PageNum: 1},
		{ProjectID: 2, SuiteID: 1, Offset: 0, Limit: 50, PageNum: 1},
		{ProjectID: 2, SuiteID: 1, Offset: 50, Limit: 50, PageNum: 2},
		{ProjectID: 2, SuiteID: 1, Offset: 100, Limit: 50, PageNum: 3},
		{ProjectID: 2, SuiteID: 1, Offset: 150, Limit: 50, PageNum: 4},
		{ProjectID: 2, SuiteID: 1, Offset: 200, Limit: 50, PageNum: 5},
		{ProjectID: 2, SuiteID: 1, Offset: 250, Limit: 50, PageNum: 6},
		{ProjectID: 2, SuiteID: 1, Offset: 300, Limit: 50, PageNum: 7},
		{ProjectID: 2, SuiteID: 1, Offset: 350, Limit: 50, PageNum: 8},
	}

	remaining, stats, err := executeRetryFailedPages(
		ctx,
		mockCli,
		failedPages,
		retryFailedPagesOptions{Attempts: 1, Workers: 1, Delay: 0},
		"test.json",
		"",
	)
	require.NoError(t, err)
	require.Len(t, remaining, len(failedPages))
	assert.Equal(t, len(failedPages), stats.RemainingPages)

	for i := 1; i < len(remaining); i++ {
		prev := remaining[i-1]
		cur := remaining[i]
		isSorted := prev.ProjectID < cur.ProjectID ||
			(prev.ProjectID == cur.ProjectID && prev.SuiteID < cur.SuiteID) ||
			(prev.ProjectID == cur.ProjectID && prev.SuiteID == cur.SuiteID && prev.Offset <= cur.Offset)
		assert.True(t, isSorted)
	}
}
