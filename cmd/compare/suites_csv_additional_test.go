package compare

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWave6F_CompareSuitesInternalWithSuites_SuccessPreloaded(t *testing.T) {
	preloaded := map[int64]data.GetSuitesResponse{
		1: {
			{ID: 1, Name: "Suite A"},
			{ID: 2, Name: "Suite B"},
		},
		2: {
			{ID: 1, Name: "Suite A"},
			{ID: 3, Name: "Suite C"},
		},
	}

	result, err := compareSuitesInternalWithSuites(context.Background(), &client.MockClient{}, 1, 2, true, preloaded)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, CompareStatusComplete, result.Status)
	assert.Len(t, result.OnlyInFirst, 1)
	assert.Len(t, result.OnlyInSecond, 1)
	assert.Len(t, result.Common, 1)
}

func TestWave6F_CompareSuitesInternalWithSuites_ErrorWhenNoData(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return nil, errors.New("upstream unavailable")
		},
	}

	result, err := compareSuitesInternalWithSuites(context.Background(), mock, 1, 2, true, nil)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get suites")
}

func TestWave6F_CompareSuitesInternalWithSuites_PartialStatusWhenNotQuiet(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Only first"}}, nil
			}
			return nil, errors.New("project 2 failed")
		},
	}

	result, err := compareSuitesInternalWithSuites(context.Background(), mock, 1, 2, false, nil)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, CompareStatusPartial, result.Status)
	assert.Len(t, result.OnlyInFirst, 1)
}

func TestWave6F_CompareSuitesInternalWithSuites_EmptyPreloaded(t *testing.T) {
	result, err := compareSuitesInternalWithSuites(context.Background(), &client.MockClient{}, 1, 2, true, map[int64]data.GetSuitesResponse{})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, CompareStatusComplete, result.Status)
	assert.Empty(t, result.OnlyInFirst)
	assert.Empty(t, result.OnlyInSecond)
	assert.Empty(t, result.Common)
}

func TestWave6F_SaveCSV_EmptyResult(t *testing.T) {
	path := t.TempDir() + "/empty.csv"

	err := saveCSV(CompareResult{}, path)
	require.NoError(t, err)

	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, "Type,Name,ID Project 1,ID Project 2\n", string(content))
}

func TestWave6F_SaveCSV_WriteErrorFromRows(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires /dev/full")
	}

	huge := strings.Repeat("x", 1<<20)

	tests := []struct {
		name   string
		result CompareResult
	}{
		{
			name: "only-in-first",
			result: CompareResult{
				OnlyInFirst: []ItemInfo{{ID: 1, Name: huge}},
			},
		},
		{
			name: "only-in-second",
			result: CompareResult{
				OnlyInSecond: []ItemInfo{{ID: 2, Name: huge}},
			},
		},
		{
			name: "common",
			result: CompareResult{
				Common: []CommonItemInfo{{Name: huge, ID1: 1, ID2: 2}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := saveCSV(tc.result, "/dev/full")
			assert.Error(t, err)
		})
	}
}
