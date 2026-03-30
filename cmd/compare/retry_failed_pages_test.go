package compare

import (
	"context"
	"fmt"
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

func TestLoadFailedPages_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "failed.json")

	jsonData := `{
  "generated_at": "2026-03-03T10:15:00Z",
  "total": 2,
  "failed_pages": [
    {"project_id":30,"suite_id":1001,"offset":0,"limit":250,"page_num":1,"error":"timeout"},
    {"project_id":34,"suite_id":2002,"offset":250,"limit":250,"page_num":2,"error":"503"}
  ]
}`
	require.NoError(t, os.WriteFile(path, []byte(jsonData), 0644))

	pages, err := loadFailedPages(path)
	require.NoError(t, err)
	require.Len(t, pages, 2)
	assert.Equal(t, int64(30), pages[0].ProjectID)
	assert.Equal(t, int64(2002), pages[1].SuiteID)
}

func TestLoadFailedPages_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "failed_invalid.json")
	require.NoError(t, os.WriteFile(path, []byte("{invalid"), 0644))

	_, err := loadFailedPages(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON report")
}

func TestLoadFailedPages_NilFailedPagesField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "failed_nil.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"generated_at":"2026-03-29T00:00:00Z","total":0}`), 0644))

	pages, err := loadFailedPages(path)
	require.NoError(t, err)
	assert.NotNil(t, pages)
	assert.Len(t, pages, 0)
}

func TestDedupeFailedPages_RemovesDuplicates(t *testing.T) {
	in := []concurrency.FailedPage{
		{ProjectID: 30, SuiteID: 1001, Offset: 0, Limit: 250, PageNum: 1},
		{ProjectID: 30, SuiteID: 1001, Offset: 0, Limit: 250, PageNum: 1},
		{ProjectID: 30, SuiteID: 1001, Offset: 250, Limit: 250, PageNum: 2},
	}

	out := dedupeFailedPages(in)
	require.Len(t, out, 2)
	assert.Equal(t, 0, out[0].Offset)
	assert.Equal(t, 250, out[1].Offset)
}

func TestDedupeFailedPages_EmptyInput(t *testing.T) {
	out := dedupeFailedPages(nil)
	assert.Nil(t, out)
}

func TestDedupeFailedPages_DerivesPageNumFromOffsetAndLimit(t *testing.T) {
	in := []concurrency.FailedPage{
		{ProjectID: 30, SuiteID: 1001, Offset: 500, Limit: 250, PageNum: 0},
	}

	out := dedupeFailedPages(in)
	require.Len(t, out, 1)
	assert.Equal(t, 3, out[0].PageNum)
}

func TestResolveRetryFailedPagesOptionsFromConfig_UsesConfigDefaults(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("compare.cases.retry.attempts", 9)
	viper.Set("compare.cases.retry.workers", 12)
	viper.Set("compare.cases.retry.delay", "2s")

	cmd := newRetryFailedPagesCmd()
	require.NoError(t, cmd.ParseFlags([]string{}))

	opts, err := resolveRetryFailedPagesOptionsFromConfig(cmd)
	require.NoError(t, err)
	assert.Equal(t, 9, opts.Attempts)
	assert.Equal(t, 12, opts.Workers)
	assert.Equal(t, 2*time.Second, opts.Delay)
}

func TestResolveRetryFailedPagesOptionsFromConfig_FlagsOverrideConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("compare.cases.retry.attempts", 9)
	viper.Set("compare.cases.retry.workers", 12)
	viper.Set("compare.cases.retry.delay", "2s")

	cmd := newRetryFailedPagesCmd()
	require.NoError(t, cmd.ParseFlags([]string{"--attempts=2", "--workers=3", "--retry-delay=250ms"}))

	opts, err := resolveRetryFailedPagesOptionsFromConfig(cmd)
	require.NoError(t, err)
	assert.Equal(t, 2, opts.Attempts)
	assert.Equal(t, 3, opts.Workers)
	assert.Equal(t, 250*time.Millisecond, opts.Delay)
}

func TestRunRetryFailedPages_FromRequired(t *testing.T) {
	originalGetClient := getClient
	defer func() { getClient = originalGetClient }()

	getClient = func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	}

	cmd := newRetryFailedPagesCmd()
	err := runRetryFailedPages(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--from flag is required")
}

func TestRunRetryFailedPages_NoClient(t *testing.T) {
	originalGetClient := getClient
	defer func() { getClient = originalGetClient }()

	getClient = nil
	cmd := newRetryFailedPagesCmd()
	err := runRetryFailedPages(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestRunRetryFailedPages_LoadFileError(t *testing.T) {
	originalGetClient := getClient
	defer func() { getClient = originalGetClient }()

	getClient = func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	}

	cmd := newRetryFailedPagesCmd()
	require.NoError(t, cmd.Flags().Set("from", filepath.Join(t.TempDir(), "missing.json")))
	err := runRetryFailedPages(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestRunRetryFailedPages_ResolveOptionsError(t *testing.T) {
	originalGetClient := getClient
	defer func() { getClient = originalGetClient }()

	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("compare.cases.retry.delay", "invalid-duration")

	getClient = func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "failed.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"generated_at":"2026-03-29T00:00:00Z","total":0,"failed_pages":[]}`), 0644))

	cmd := newRetryFailedPagesCmd()
	require.NoError(t, cmd.Flags().Set("from", path))

	err := runRetryFailedPages(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid compare.cases.retry.delay")
}

func TestRunRetryFailedPages_Success(t *testing.T) {
	originalGetClient := getClient
	defer func() { getClient = originalGetClient }()

	viper.Reset()
	t.Cleanup(viper.Reset)

	getClient = func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "failed.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"generated_at":"2026-03-29T00:00:00Z","total":0,"failed_pages":[]}`), 0644))

	cmd := newRetryFailedPagesCmd()
	require.NoError(t, cmd.Flags().Set("from", path))

	err := runRetryFailedPages(cmd, nil)
	require.NoError(t, err)
}

func TestResolveRetryFailedPagesOptionsFromConfig_InvalidDelay(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("compare.cases.retry.delay", "not-a-duration")

	cmd := newRetryFailedPagesCmd()
	require.NoError(t, cmd.ParseFlags([]string{}))

	_, err := resolveRetryFailedPagesOptionsFromConfig(cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid compare.cases.retry.delay")
}

func TestExecuteRetryFailedPages_EmptyRetryList(t *testing.T) {
	ctx := context.Background()
	mockCli := &client.MockClient{}
	opts := retryFailedPagesOptions{Attempts: 1, Workers: 1, Delay: 0}

	remaining, stats, err := executeRetryFailedPages(
		ctx, mockCli, []concurrency.FailedPage{},
		opts, "empty.json", "",
	)

	require.NoError(t, err)
	assert.Nil(t, remaining)
	assert.Equal(t, 0, stats.InputPages)
	assert.Equal(t, 0, stats.RecoveredPages)
}

func TestExecuteRetryFailedPages_EmptyRetryListWithoutSourceLabel(t *testing.T) {
	ctx := context.Background()
	mockCli := &client.MockClient{}
	opts := retryFailedPagesOptions{Attempts: 1, Workers: 1, Delay: 0}

	remaining, stats, err := executeRetryFailedPages(
		ctx, mockCli, []concurrency.FailedPage{},
		opts, "", "",
	)

	require.NoError(t, err)
	assert.Nil(t, remaining)
	assert.Equal(t, 0, stats.InputPages)
	assert.Equal(t, 0, stats.UniquePages)
	assert.Equal(t, 0, stats.RemainingPages)
}

func TestExecuteRetryFailedPages_FullRetryFlow_AllSuccess(t *testing.T) {
	ctx := context.Background()

	mockCli := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{
				{ID: 100, Title: "Test case 1"},
				{ID: 101, Title: "Test case 2"},
			}, nil
		},
	}

	failedPages := []concurrency.FailedPage{
		{ProjectID: 30, SuiteID: 1001, Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"},
		{ProjectID: 30, SuiteID: 1001, Offset: 250, Limit: 250, PageNum: 2, Error: "503"},
	}

	opts := retryFailedPagesOptions{Attempts: 2, Workers: 2, Delay: 10 * time.Millisecond}

	remaining, stats, err := executeRetryFailedPages(
		ctx, mockCli, failedPages, opts, "test.json", "",
	)

	require.NoError(t, err)
	assert.Equal(t, 2, stats.InputPages)
	assert.Equal(t, 2, stats.UniquePages)
	assert.Equal(t, 2, stats.RecoveredPages)
	assert.Equal(t, 4, stats.RecoveredCases) // 2 pages × 2 cases each
	assert.Nil(t, remaining)
}

func TestExecuteRetryFailedPages_PartialRetrySuccess(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	mockCli := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			callCount++
			// First page succeeds, second page fails
			if offset == 0 {
				return data.GetCasesResponse{
					{ID: 100, Title: "Test"},
				}, nil
			}
			return data.GetCasesResponse{}, fmt.Errorf("still failing")
		},
	}

	failedPages := []concurrency.FailedPage{
		{ProjectID: 30, SuiteID: 1001, Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"},
		{ProjectID: 30, SuiteID: 1001, Offset: 250, Limit: 250, PageNum: 2, Error: "503"},
	}

	opts := retryFailedPagesOptions{Attempts: 2, Workers: 1, Delay: 5 * time.Millisecond}

	remaining, stats, err := executeRetryFailedPages(
		ctx, mockCli, failedPages, opts, "test.json", "",
	)

	require.NoError(t, err)
	assert.Equal(t, 1, stats.RecoveredPages)
	assert.Equal(t, 1, stats.RemainingPages)
	require.Len(t, remaining, 1)
	assert.Equal(t, 250, remaining[0].Offset)
}

func TestExecuteRetryFailedPages_AllRetryFailed(t *testing.T) {
	ctx := context.Background()

	mockCli := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, fmt.Errorf("persistent network error")
		},
	}

	failedPages := []concurrency.FailedPage{
		{ProjectID: 30, SuiteID: 1001, Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"},
	}

	opts := retryFailedPagesOptions{Attempts: 1, Workers: 1, Delay: 0}

	remaining, stats, err := executeRetryFailedPages(
		ctx, mockCli, failedPages, opts, "test.json", "",
	)

	require.NoError(t, err)
	assert.Equal(t, 0, stats.RecoveredPages)
	assert.Equal(t, 1, stats.RemainingPages)
	require.Len(t, remaining, 1)
	assert.Contains(t, remaining[0].Error, "persistent network error")
}

func TestExecuteRetryFailedPages_UserAbortContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockCli := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, ctx.Err()
		},
	}

	failedPages := []concurrency.FailedPage{
		{ProjectID: 30, SuiteID: 1001, Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"},
	}

	opts := retryFailedPagesOptions{Attempts: 3, Workers: 2, Delay: 100 * time.Millisecond}

	remaining, stats, err := executeRetryFailedPages(
		ctx, mockCli, failedPages, opts, "test.json", "",
	)

	require.NoError(t, err)
	assert.Equal(t, 0, stats.RecoveredPages)
	assert.Equal(t, 1, stats.RemainingPages)
	require.Len(t, remaining, 1)
}

func TestExecuteRetryFailedPages_NormalizesOptionsDefaults(t *testing.T) {
	ctx := context.Background()

	mockCli := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: 1}}, nil
		},
	}

	failedPages := []concurrency.FailedPage{
		{ProjectID: 30, SuiteID: 1001, Offset: 0, Limit: 250, PageNum: 1},
	}

	// Pass invalid options
	opts := retryFailedPagesOptions{
		Attempts: -1,    // Invalid
		Workers:  0,     // Invalid
		Delay:    -100,  // Invalid
	}

	remaining, _, err := executeRetryFailedPages(
		ctx, mockCli, failedPages, opts, "test.json", "",
	)

	require.NoError(t, err)
	assert.Nil(t, remaining) // Should succeed with corrected values
}
