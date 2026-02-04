// internal/client/runs_test.go
// Тесты для Runs API POST-методов
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddRun(t *testing.T) {
	tests := []struct {
		name         string
		projectID    int64
		request      *data.AddRunRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
		expectedErr  string
	}{
		{
			name:      "successful run creation",
			projectID: 30,
			request: &data.AddRunRequest{
				Name:        "Smoke Tests",
				Description: "Daily smoke tests",
				SuiteID:     100,
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Run{
				ID:          12345,
				ProjectID:   30,
				SuiteID:     100,
				Name:        "Smoke Tests",
				Description: "Daily smoke tests",
			},
			wantErr: false,
		},
		{
			name:      "run with all fields",
			projectID: 30,
			request: &data.AddRunRequest{
				Name:        "Regression v2.0",
				Description: "Full regression",
				SuiteID:     100,
				MilestoneID: 50,
				AssignedTo:  10,
				CaseIDs:     []int64{1, 2, 3, 4, 5},
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Run{
				ID:          12345,
				ProjectID:   30,
				Name:        "Regression v2.0",
				Description: "Full regression",
				SuiteID:     100,
				MilestoneID: 50,
				AssignedTo:  10,
			},
			wantErr: false,
		},
		{
			name:         "API error - project not found",
			projectID:    99999,
			request:      &data.AddRunRequest{Name: "Test", SuiteID: 100},
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
			expectedErr:  "404",
		},
		{
			name:         "API error - invalid suite",
			projectID:    30,
			request:      &data.AddRunRequest{Name: "Test", SuiteID: 99999},
			mockStatus:   http.StatusBadRequest,
			mockResponse: nil,
			wantErr:      true,
			expectedErr:  "400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := "/index.php?/api/v2/add_run/30"
				if tt.projectID == 99999 {
					expectedPath = "/index.php?/api/v2/add_run/99999"
				}
				assert.Equal(t, expectedPath, r.URL.String())
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var req data.AddRunRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, tt.request.Name, req.Name)
				assert.Equal(t, tt.request.SuiteID, req.SuiteID)

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			run, err := client.AddRun(tt.projectID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, run)
				assert.Equal(t, int64(12345), run.ID)
				assert.Equal(t, tt.request.Name, run.Name)
			}
		})
	}
}

func TestUpdateRun(t *testing.T) {
	tests := []struct {
		name         string
		runID        int64
		request      *data.UpdateRunRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:  "successful update",
			runID: 12345,
			request: &data.UpdateRunRequest{
				Name:        ptr("Updated Name"),
				Description: ptr("Updated description"),
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Run{
				ID:          12345,
				Name:        "Updated Name",
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name:  "update assigned user",
			runID: 12345,
			request: &data.UpdateRunRequest{
				AssignedTo: ptr(int64(20)),
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Run{
				ID:         12345,
				AssignedTo: 20,
			},
			wantErr: false,
		},
		{
			name:         "run not found",
			runID:        99999,
			request:      &data.UpdateRunRequest{Name: ptr("Test")},
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/update_run/%d", tt.runID)
				assert.Equal(t, expectedPath, r.URL.String())

				var req data.UpdateRunRequest
				json.NewDecoder(r.Body).Decode(&req)

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			run, err := client.UpdateRun(tt.runID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, run)
			}
		})
	}
}

func TestCloseRun(t *testing.T) {
	tests := []struct {
		name         string
		runID        int64
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:   "successful close",
			runID:  12345,
			mockStatus: http.StatusOK,
			mockResponse: data.Run{
				ID:          12345,
				IsCompleted: true,
				CompletedOn: 1707000000,
			},
			wantErr: false,
		},
		{
			name:         "already closed",
			runID:        12345,
			mockStatus:   http.StatusBadRequest,
			mockResponse: nil,
			wantErr:      true,
		},
		{
			name:         "run not found",
			runID:        99999,
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/close_run/%d", tt.runID)
				assert.Equal(t, expectedPath, r.URL.String())
				// CloseRun не имеет тела запроса
				assert.Equal(t, int64(0), r.ContentLength)

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			run, err := client.CloseRun(tt.runID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, run)
				assert.True(t, run.IsCompleted)
			}
		})
	}
}

func TestDeleteRun(t *testing.T) {
	tests := []struct {
		name       string
		runID      int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			runID:      12345,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "run not found",
			runID:      99999,
			mockStatus: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "already deleted",
			runID:      12345,
			mockStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/delete_run/%d", tt.runID)
				assert.Equal(t, expectedPath, r.URL.String())

				w.WriteHeader(tt.mockStatus)
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			err := client.DeleteRun(tt.runID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
