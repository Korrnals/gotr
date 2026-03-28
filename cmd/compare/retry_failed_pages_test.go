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

func TestExecuteRetryFailedPages(t *testing.T) {
	mock := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			if offset == 0 {
				return data.GetCasesResponse{{ID: 1, Title: "Recovered"}}, nil
			}
			return nil, fmt.Errorf("temporary error")
		},
	}

	failed := []concurrency.FailedPage{
		{ProjectID: 1, SuiteID: 2, Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"},
		{ProjectID: 1, SuiteID: 2, Offset: 250, Limit: 250, PageNum: 2, Error: "timeout"},
	}

	remaining, stats, err := executeRetryFailedPages(context.Background(), mock, failed, retryFailedPagesOptions{Attempts: 2, Workers: 1, Delay: 0}, "", "")
	require.NoError(t, err)
	assert.Len(t, remaining, 1)
	assert.Equal(t, 2, stats.InputPages)
	assert.Equal(t, 1, stats.RecoveredPages)
	assert.Equal(t, 1, stats.RecoveredCases)
}

func TestExecuteRetryFailedPages_EmptyInput(t *testing.T) {
	remaining, stats, err := executeRetryFailedPages(context.Background(), &client.MockClient{}, nil, retryFailedPagesOptions{}, "empty-source", "")
	require.NoError(t, err)
	assert.Nil(t, remaining)
	assert.Equal(t, 0, stats.InputPages)
	assert.Equal(t, 0, stats.UniquePages)
}

func TestExecuteRetryFailedPages_SaveRemainingError(t *testing.T) {
	mock := &client.MockClient{
		GetCasesPageFunc: func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			return nil, fmt.Errorf("always failing")
		},
	}

	failed := []concurrency.FailedPage{{ProjectID: 1, SuiteID: 2, Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"}}
	badSavePath := t.TempDir()

	remaining, stats, err := executeRetryFailedPages(
		context.Background(),
		mock,
		failed,
		retryFailedPagesOptions{Attempts: 0, Workers: 0, Delay: -time.Second},
		"source.json",
		badSavePath,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save remaining failed pages error")
	assert.Len(t, remaining, 1)
	assert.Equal(t, 1, stats.RemainingPages)
}
