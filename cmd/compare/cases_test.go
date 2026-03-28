// Package compare tests - tests for cases comparison
package compare

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintCasesStats(t *testing.T) {
	result := &CompareResult{
		Resource:     "cases",
		Project1ID:   30,
		Project2ID:   34,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Case A"}, {ID: 2, Name: "Case B"}},
		OnlyInSecond: []ItemInfo{{ID: 3, Name: "Case C"}},
		Common:       []CommonItemInfo{{Name: "Case D", ID1: 4, ID2: 5, IDsMatch: true}},
	}
	elapsed := 1500 * time.Millisecond

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printCasesStats(result, elapsed)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains key information
	assert.Contains(t, output, "STATS: cases")
	assert.Contains(t, output, "1.5s") // elapsed time
	assert.Contains(t, output, "4")    // total cases (2 + 1 + 1)
	assert.Contains(t, output, "30")   // project 1 ID
	assert.Contains(t, output, "34")   // project 2 ID
	assert.Contains(t, output, "2")    // only in first
	assert.Contains(t, output, "1")    // only in second / common
}

func TestPrintCasesStats_ZeroCases(t *testing.T) {
	result := &CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
	}
	elapsed := 500 * time.Millisecond

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printCasesStats(result, elapsed)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "0") // total cases
	assert.Contains(t, output, "500ms")
}

func TestSaveFailedPagesReport(t *testing.T) {
	pages := []concurrency.FailedPage{{ProjectID: 1, SuiteID: 2, Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"}}
	path := filepath.Join(t.TempDir(), "failed_pages.json")

	savedPath, err := saveFailedPagesReport(pages, path)
	assert.NoError(t, err)
	assert.Equal(t, path, savedPath)

	data, readErr := os.ReadFile(path)
	assert.NoError(t, readErr)
	assert.Contains(t, string(data), "failed_pages")
}

func TestFetchCasesForProject_NoSuites(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{
				{ID: 1, Title: "Case 1", SectionID: 10},
				{ID: 2, Title: "Case 2", SectionID: 20},
			}, nil
		},
	}

	items, failed, stats, err := fetchCasesForProject(context.Background(), mock, 1, data.GetSuitesResponse{}, nil, 1, 1, time.Second, 60, 2)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Nil(t, failed)
	assert.Equal(t, 2, stats.CasesRaw)
	assert.Equal(t, 2, stats.CasesUnique)
	assert.Equal(t, 2, stats.Sections)
}

func TestFetchCasesForProject_WithSuitesParallel(t *testing.T) {
	mock := &client.MockClient{
		GetCasesParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, []int64{10, 20}, suiteIDs)
			require.NotNil(t, cfg)
			return data.GetCasesResponse{
				{ID: 1, Title: "Case 1", SectionID: 10},
				{ID: 1, Title: "Case 1", SectionID: 10},
				{ID: 2, Title: "", SectionID: 20},
			}, &concurrency.ExecutionResult{
				FailedPages: []concurrency.FailedPage{{ProjectID: 1, SuiteID: 10, Offset: 0, Limit: 250, PageNum: 1, Error: "timeout"}},
				Stats: concurrency.AggregationStats{
					TotalPages:      3,
					FailedPages:     1,
					ExpectedCases:   3,
					SuitesWithTotal: 2,
					SuitesVerified:  1,
					SuiteResults: []concurrency.SuiteResultInfo{
						{SuiteID: 10, CasesFetched: 2, Verified: true},
						{SuiteID: 20, CasesFetched: 1, Verified: false},
					},
				},
			}, nil
		},
	}

	suites := data.GetSuitesResponse{{ID: 10, Name: "S1"}, {ID: 20, Name: "S2"}}
	items, failed, stats, err := fetchCasesForProject(context.Background(), mock, 1, suites, nil, 2, 3, time.Second, 120, 3)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Len(t, failed, 1)
	assert.Equal(t, 3, stats.CasesRaw)
	assert.Equal(t, 2, stats.CasesUnique)
	assert.Equal(t, 1, stats.EmptyTitles)
	assert.Equal(t, 2, stats.Sections)
	assert.Equal(t, 3, stats.TotalPages)
	assert.Equal(t, 1, stats.FailedPages)
}
