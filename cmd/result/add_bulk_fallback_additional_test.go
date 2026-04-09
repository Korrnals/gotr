package result

import (
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWave6F_AddBulkResults_TestEntries_Success(t *testing.T) {
	var addResultsCalled bool
	var addCasesCalled bool

	mock := &client.MockClient{
		AddResultsFunc: func(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			addResultsCalled = true
			require.Equal(t, int64(42), runID)
			require.Len(t, req.Results, 1)
			assert.Equal(t, int64(101), req.Results[0].TestID)
			return data.GetResultsResponse{{ID: 1, TestID: 101}}, nil
		},
		AddResultsForCasesFunc: func(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
			addCasesCalled = true
			return nil, nil
		},
	}

	svc := newResultServiceFromInterface(mock)
	payload := []byte(`[{"test_id":101,"status_id":1,"comment":"ok"}]`)

	result, err := svc.AddBulkResults(context.Background(), 42, payload)

	require.NoError(t, err)
	assert.True(t, addResultsCalled)
	assert.False(t, addCasesCalled)
	assert.NotNil(t, result)
}

func TestWave6F_AddBulkResults_TestEntries_ServiceError(t *testing.T) {
	mock := &client.MockClient{
		AddResultsFunc: func(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			return nil, errors.New("add results failed")
		},
	}

	svc := newResultServiceFromInterface(mock)
	payload := []byte(`[{"test_id":101,"status_id":1}]`)

	result, err := svc.AddBulkResults(context.Background(), 42, payload)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "add results failed")
}

func TestWave6F_AddBulkResults_CaseEntries_FallbackSuccess(t *testing.T) {
	var addResultsCalled bool
	var addCasesCalled bool

	mock := &client.MockClient{
		AddResultsFunc: func(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			addResultsCalled = true
			return nil, nil
		},
		AddResultsForCasesFunc: func(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
			addCasesCalled = true
			require.Len(t, req.Results, 1)
			assert.Equal(t, int64(555), req.Results[0].CaseID)
			return data.GetResultsResponse{{ID: 2, TestID: 202}}, nil
		},
	}

	svc := newResultServiceFromInterface(mock)
	// test_id as an object breaks parsing of []ResultEntry, but remains valid for []ResultForCaseEntry.
	payload := []byte(`[{"test_id":{},"case_id":555,"status_id":1}]`)

	result, err := svc.AddBulkResults(context.Background(), 77, payload)

	require.NoError(t, err)
	assert.False(t, addResultsCalled)
	assert.True(t, addCasesCalled)
	assert.NotNil(t, result)
}

func TestWave6F_AddBulkResults_CaseEntries_FallbackError(t *testing.T) {
	mock := &client.MockClient{
		AddResultsForCasesFunc: func(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
			return nil, errors.New("add results for cases failed")
		},
	}

	svc := newResultServiceFromInterface(mock)
	payload := []byte(`[{"test_id":{},"case_id":777,"status_id":5}]`)

	result, err := svc.AddBulkResults(context.Background(), 77, payload)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "add results for cases failed")
}

func TestWave6F_AddBulkResults_EmptyArray_Error(t *testing.T) {
	svc := newResultServiceFromInterface(&client.MockClient{})

	result, err := svc.AddBulkResults(context.Background(), 77, []byte(`[]`))

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse JSON file")
}
