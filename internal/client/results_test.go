// internal/client/results_test.go
// Тесты для Results API POST-методов
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddResult(t *testing.T) {
	tests := []struct {
		name         string
		testID       int64
		request      *data.AddResultRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
		expectedErr  string
	}{
		{
			name:   "successful result creation",
			testID: 12345,
			request: &data.AddResultRequest{
				StatusID: 1,
				Comment:  "Test passed",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Result{
				ID:       999,
				TestID:   12345,
				StatusID: 1,
				Comment:  "Test passed",
			},
			wantErr: false,
		},
		{
			name:   "result with all fields",
			testID: 12345,
			request: &data.AddResultRequest{
				StatusID:   5,
				Comment:    "Found bug",
				Version:    "v1.0.0",
				Elapsed:    "1m 30s",
				Defects:    "BUG-123",
				AssignedTo: 10,
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Result{
				ID:         999,
				TestID:     12345,
				StatusID:   5,
				Comment:    "Found bug",
				Version:    "v1.0.0",
				Elapsed:    "1m 30s",
				Defects:    "BUG-123",
				AssignedTo: 10,
			},
			wantErr: false,
		},
		{
			name:         "API error",
			testID:       12345,
			request:      &data.AddResultRequest{StatusID: 1},
			mockStatus:   http.StatusBadRequest,
			mockResponse: nil,
			wantErr:      true,
			expectedErr:  "400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := func(w http.ResponseWriter, r *http.Request) {
				// Проверяем метод и путь
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/index.php?/api/v2/add_result/12345", r.URL.String())
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Проверяем тело запроса
				var req data.AddResultRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, tt.request.StatusID, req.StatusID)

				// Отправляем ответ
				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			// Act
			result, err := client.AddResult(tt.testID, tt.request)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, int64(999), result.ID)
				assert.Equal(t, tt.request.StatusID, result.StatusID)
			}
		})
	}
}

func TestAddResultForCase(t *testing.T) {
	tests := []struct {
		name         string
		runID        int64
		caseID       int64
		request      *data.AddResultRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:   "successful result for case",
			runID:  100,
			caseID: 200,
			request: &data.AddResultRequest{
				StatusID: 1,
				Comment:  "Passed",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Result{
				ID:     999,
				TestID: 12345,
			},
			wantErr: false,
		},
		{
			name:         "invalid case id",
			runID:        100,
			caseID:       0,
			request:      &data.AddResultRequest{StatusID: 1},
			mockStatus:   http.StatusBadRequest,
			mockResponse: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/add_result_for_case/%d/%d", tt.runID, tt.caseID)
				assert.Equal(t, expectedPath, r.URL.String())

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			result, err := client.AddResultForCase(tt.runID, tt.caseID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAddResults(t *testing.T) {
	tests := []struct {
		name         string
		runID        int64
		request      *data.AddResultsRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:   "successful bulk add",
			runID:  100,
			request: &data.AddResultsRequest{
				Results: []data.ResultEntry{
					{TestID: 1, StatusID: 1, Comment: "Passed"},
					{TestID: 2, StatusID: 5, Comment: "Failed"},
				},
			},
			mockStatus: http.StatusOK,
			mockResponse: []data.Result{
				{ID: 901, TestID: 1, StatusID: 1},
				{ID: 902, TestID: 2, StatusID: 5},
			},
			wantErr: false,
		},
		{
			name:   "empty results array",
			runID:  100,
			request: &data.AddResultsRequest{
				Results: []data.ResultEntry{},
			},
			mockStatus:   http.StatusBadRequest,
			mockResponse: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/index.php?/api/v2/add_results/100", r.URL.String())

				// Проверяем тело запроса
				var req data.AddResultsRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Len(t, req.Results, len(tt.request.Results))

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			results, err := client.AddResults(tt.runID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, len(tt.request.Results))
			}
		})
	}
}

func TestAddResultsForCases(t *testing.T) {
	tests := []struct {
		name         string
		runID        int64
		request      *data.AddResultsForCasesRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:   "successful bulk for cases",
			runID:  100,
			request: &data.AddResultsForCasesRequest{
				Results: []data.ResultForCaseEntry{
					{CaseID: 10, StatusID: 1},
					{CaseID: 20, StatusID: 5},
				},
			},
			mockStatus: http.StatusOK,
			mockResponse: []data.Result{
				{ID: 901, TestID: 1001, StatusID: 1},
				{ID: 902, TestID: 1002, StatusID: 5},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/index.php?/api/v2/add_results_for_cases/100", r.URL.String())

				var req data.AddResultsForCasesRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Len(t, req.Results, len(tt.request.Results))

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			results, err := client.AddResultsForCases(tt.runID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, results)
			}
		})
	}
}
