// internal/service/result_test.go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResultService_Constructors(t *testing.T) {
	httpClient := &client.HTTPClient{}
	svc := NewResultService(httpClient)
	assert.NotNil(t, svc)

	mock := &client.MockClient{}
	svc2 := NewResultServiceFromInterface(mock)
	assert.NotNil(t, svc2)
}

func TestResultService_Getters(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{}
	svc := NewResultServiceFromInterface(mock)

	t.Run("GetForTest invalid id", func(t *testing.T) {
		_, err := svc.GetForTest(ctx, 0)
		assert.Error(t, err)
	})

	t.Run("GetForTest success", func(t *testing.T) {
		mock.GetResultsFunc = func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{{ID: 1, TestID: testID}}, nil
		}

		res, err := svc.GetForTest(ctx, 10)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, int64(10), res[0].TestID)
	})

	t.Run("GetForCase invalid ids", func(t *testing.T) {
		_, err := svc.GetForCase(ctx, 0, 1)
		assert.Error(t, err)
		_, err = svc.GetForCase(ctx, 1, 0)
		assert.Error(t, err)
	})

	t.Run("GetForCase success", func(t *testing.T) {
		mock.GetResultsForCaseFunc = func(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{{ID: 2, TestID: 200}}, nil
		}

		res, err := svc.GetForCase(ctx, 11, 22)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("GetForRun success", func(t *testing.T) {
		mock.GetResultsForRunFunc = func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{{ID: 3}}, nil
		}

		res, err := svc.GetForRun(ctx, 33)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("GetRunsForProject success", func(t *testing.T) {
		mock.GetRunsFunc = func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 99, Name: "run"}}, nil
		}

		runs, err := svc.GetRunsForProject(ctx, 44)
		assert.NoError(t, err)
		assert.Len(t, runs, 1)
		assert.Equal(t, int64(99), runs[0].ID)
	})
}

func TestResultService_AddMethods(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{}
	svc := NewResultServiceFromInterface(mock)

	t.Run("AddForTest validation error", func(t *testing.T) {
		_, err := svc.AddForTest(ctx, 0, &data.AddResultRequest{StatusID: 1})
		assert.Error(t, err)
	})

	t.Run("AddForTest client error", func(t *testing.T) {
		mock.AddResultFunc = func(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
			return nil, errors.New("add failed")
		}

		_, err := svc.AddForTest(ctx, 1, &data.AddResultRequest{StatusID: 1})
		assert.Error(t, err)
	})

	t.Run("AddForTest success", func(t *testing.T) {
		mock.AddResultFunc = func(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
			return &data.Result{ID: 10, TestID: testID, StatusID: req.StatusID}, nil
		}

		res, err := svc.AddForTest(ctx, 1, &data.AddResultRequest{StatusID: 5})
		assert.NoError(t, err)
		assert.Equal(t, int64(10), res.ID)
	})

	t.Run("AddForCase success", func(t *testing.T) {
		mock.AddResultForCaseFunc = func(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
			return &data.Result{ID: 11, StatusID: req.StatusID}, nil
		}

		res, err := svc.AddForCase(ctx, 2, 3, &data.AddResultRequest{StatusID: 1})
		assert.NoError(t, err)
		assert.Equal(t, int64(11), res.ID)
	})

	t.Run("AddResults success", func(t *testing.T) {
		mock.AddResultsFunc = func(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{{ID: 12, StatusID: req.Results[0].StatusID}}, nil
		}

		res, err := svc.AddResults(ctx, 4, &data.AddResultsRequest{Results: []data.ResultEntry{{TestID: 1, StatusID: 2}}})
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("AddResultsForCases success", func(t *testing.T) {
		mock.AddResultsForCasesFunc = func(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{{ID: 13, StatusID: req.Results[0].StatusID}}, nil
		}

		res, err := svc.AddResultsForCases(ctx, 5, &data.AddResultsForCasesRequest{Results: []data.ResultForCaseEntry{{CaseID: 6, StatusID: 1}}})
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})
}

func TestResultService_ParseID(t *testing.T) {
	svc := &ResultService{}

	t.Run("missing arg", func(t *testing.T) {
		_, err := svc.ParseID(context.Background(), []string{}, 0)
		assert.Error(t, err)
	})

	t.Run("invalid number", func(t *testing.T) {
		_, err := svc.ParseID(context.Background(), []string{"abc"}, 0)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		id, err := svc.ParseID(context.Background(), []string{"123"}, 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(123), id)
	})
}

func TestResultService_validateID(t *testing.T) {
	svc := &ResultService{}

	tests := []struct {
		name      string
		id        int64
		fieldName string
		wantErr   bool
	}{
		{"valid positive ID", 123, "test_id", false},
		{"zero ID", 0, "run_id", true},
		{"negative ID", -10, "case_id", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateID(tt.id, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResultService_validateAddResultRequest(t *testing.T) {
	svc := &ResultService{}

	tests := []struct {
		name    string
		req     *data.AddResultRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "nil",
		},
		{
			name: "zero status_id",
			req: &data.AddResultRequest{
				StatusID: 0,
				Comment:  "Test",
			},
			wantErr: true,
			errMsg:  "status_id",
		},
		{
			name: "negative status_id",
			req: &data.AddResultRequest{
				StatusID: -1,
				Comment:  "Test",
			},
			wantErr: true,
			errMsg:  "status_id",
		},
		{
			name: "valid status_id",
			req: &data.AddResultRequest{
				StatusID: 1,
				Comment:  "Test passed",
			},
			wantErr: false,
		},
		{
			name: "valid with all fields",
			req: &data.AddResultRequest{
				StatusID:   5,
				Comment:    "Failed",
				Version:    "v1.0",
				Elapsed:    "1m",
				Defects:    "BUG-123",
				AssignedTo: 10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateAddResultRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResultService_validateAddResultsRequest(t *testing.T) {
	svc := &ResultService{}

	tests := []struct {
		name    string
		req     *data.AddResultsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "nil",
		},
		{
			name:    "empty results",
			req:     &data.AddResultsRequest{Results: []data.ResultEntry{}},
			wantErr: true,
			errMsg:  "empty",
		},
		{
			name: "valid single result",
			req: &data.AddResultsRequest{
				Results: []data.ResultEntry{
					{TestID: 1, StatusID: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid status in result",
			req: &data.AddResultsRequest{
				Results: []data.ResultEntry{
					{TestID: 1, StatusID: 0},
				},
			},
			wantErr: true,
			errMsg:  "status_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateAddResultsRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResultService_validateAddResultsForCasesRequest(t *testing.T) {
	svc := &ResultService{}

	tests := []struct {
		name    string
		req     *data.AddResultsForCasesRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "nil",
		},
		{
			name:    "empty results",
			req:     &data.AddResultsForCasesRequest{Results: []data.ResultForCaseEntry{}},
			wantErr: true,
			errMsg:  "empty",
		},
		{
			name: "valid result",
			req: &data.AddResultsForCasesRequest{
				Results: []data.ResultForCaseEntry{
					{CaseID: 100, StatusID: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "missing case_id",
			req: &data.AddResultsForCasesRequest{
				Results: []data.ResultForCaseEntry{
					{CaseID: 0, StatusID: 1},
				},
			},
			wantErr: true,
			errMsg:  "case_id",
		},
		{
			name: "missing status_id",
			req: &data.AddResultsForCasesRequest{
				Results: []data.ResultForCaseEntry{
					{CaseID: 100, StatusID: 0},
				},
			},
			wantErr: true,
			errMsg:  "status_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateAddResultsForCasesRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
