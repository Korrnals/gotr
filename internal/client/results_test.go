// internal/client/results_test.go
// Tests for Results API POST methods
package client

import (
	"context"
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
				// Verify method and path
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/index.php?/api/v2/add_result/12345", r.URL.String())
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Verify request body
				var req data.AddResultRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, tt.request.StatusID, req.StatusID)

				// Send response
				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			// Act
			ctx := context.Background()
			result, err := client.AddResult(ctx, tt.testID, tt.request)

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

			ctx := context.Background()
			result, err := client.AddResultForCase(ctx, tt.runID, tt.caseID, tt.request)

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
			name:  "successful bulk add",
			runID: 100,
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
			name:  "empty results array",
			runID: 100,
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

				// Verify request body
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

			ctx := context.Background()
			results, err := client.AddResults(ctx, tt.runID, tt.request)

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
			name:  "successful bulk for cases",
			runID: 100,
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

			ctx := context.Background()
			results, err := client.AddResultsForCases(ctx, tt.runID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, results)
			}
		})
	}
}

func TestGetResultsEndpoints(t *testing.T) {
	t.Run("GetResults success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php", r.URL.Path)
			assert.Contains(t, r.URL.String(), "get_results/123")
			assert.Equal(t, "0", r.URL.Query().Get("offset"))
			assert.Equal(t, "250", r.URL.Query().Get("limit"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetResultsResponse{{ID: 1, TestID: 123}})
		})
		defer server.Close()

		results, err := client.GetResults(context.Background(), 123)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("GetResultsForRun success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php", r.URL.Path)
			assert.Contains(t, r.URL.String(), "get_results_for_run/55")
			assert.Equal(t, "0", r.URL.Query().Get("offset"))
			assert.Equal(t, "250", r.URL.Query().Get("limit"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetResultsResponse{{ID: 2, TestID: 777}})
		})
		defer server.Close()

		results, err := client.GetResultsForRun(context.Background(), 55)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("GetResultsForCase success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_results_for_case/55/66", r.URL.String())
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetResultsResponse{{ID: 3, TestID: 888}})
		})
		defer server.Close()

		results, err := client.GetResultsForCase(context.Background(), 55, 66)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestResults_ErrorBranches(t *testing.T) {
	t.Run("AddResult decode error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/add_result/321", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"broken":`))
		})
		defer server.Close()

		_, err := client.AddResult(context.Background(), 321, &data.AddResultRequest{StatusID: 1})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error result")
	})

	t.Run("AddResultForCase decode error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/add_result_for_case/12/34", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"broken":`))
		})
		defer server.Close()

		_, err := client.AddResultForCase(context.Background(), 12, 34, &data.AddResultRequest{StatusID: 1})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error result")
	})

	t.Run("AddResults decode error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/add_results/55", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"broken":`))
		})
		defer server.Close()

		_, err := client.AddResults(context.Background(), 55, &data.AddResultsRequest{Results: []data.ResultEntry{{TestID: 1, StatusID: 1}}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error results")
	})

	t.Run("AddResultsForCases non-OK", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/add_results_for_cases/55", r.URL.String())
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`bad`))
		})
		defer server.Close()

		_, err := client.AddResultsForCases(context.Background(), 55, &data.AddResultsForCasesRequest{Results: []data.ResultForCaseEntry{{CaseID: 10, StatusID: 5}}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned 400 Bad Request")
	})

	t.Run("AddResultsForCases decode error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/add_results_for_cases/56", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"broken":`))
		})
		defer server.Close()

		_, err := client.AddResultsForCases(context.Background(), 56, &data.AddResultsForCasesRequest{Results: []data.ResultForCaseEntry{{CaseID: 11, StatusID: 1}}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error results")
	})

	t.Run("GetResults request error from API status", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.String(), "get_results/321")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`boom`))
		})
		defer server.Close()

		_, err := client.GetResults(context.Background(), 321)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request error GetResults")
	})

	t.Run("GetResultsForRun request error from API status", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.String(), "get_results_for_run/654")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`boom`))
		})
		defer server.Close()

		_, err := client.GetResultsForRun(context.Background(), 654)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request error GetResultsForRun")
	})

	t.Run("GetResultsForCase non-OK", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_results_for_case/7/8", r.URL.String())
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
		})
		defer server.Close()

		_, err := client.GetResultsForCase(context.Background(), 7, 8)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned 404 Not Found")
	})

	t.Run("GetResultsForCase decode error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_results_for_case/9/10", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"broken":`))
		})
		defer server.Close()

		_, err := client.GetResultsForCase(context.Background(), 9, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error results")
	})
}
