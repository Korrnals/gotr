// internal/service/result_test.go
package service

import (
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

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
			errMsg:  "пустым",
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
			errMsg:  "пустым",
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
