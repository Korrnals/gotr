// internal/client/suites_test.go
// Тесты для Suites API POST-методов
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddSuite(t *testing.T) {
	tests := []struct {
		name         string
		projectID    int64
		request      *data.AddSuiteRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
		expectedErr  string
	}{
		{
			name:      "successful suite creation",
			projectID: 30,
			request: &data.AddSuiteRequest{
				Name:        "Regression Suite",
				Description: "Full regression test suite",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Suite{
				ID:          100,
				ProjectID:   30,
				Name:        "Regression Suite",
				Description: "Full regression test suite",
				IsBaseline:  false,
				IsMaster:    false,
				IsCompleted: false,
			},
			wantErr: false,
		},
		{
			name:      "minimal suite (name only)",
			projectID: 30,
			request: &data.AddSuiteRequest{
				Name: "Quick Suite",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Suite{
				ID:        101,
				ProjectID: 30,
				Name:      "Quick Suite",
			},
			wantErr: false,
		},
		{
			name:         "project not found",
			projectID:    99999,
			request:      &data.AddSuiteRequest{Name: "Test Suite"},
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
			expectedErr:  "404",
		},
		{
			name:         "duplicate name",
			projectID:    30,
			request:      &data.AddSuiteRequest{Name: "Existing Suite"},
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
				expectedPath := fmt.Sprintf("/index.php?/api/v2/add_suite/%d", tt.projectID)
				assert.Equal(t, expectedPath, r.URL.String())
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var req data.AddSuiteRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, tt.request.Name, req.Name)

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			suite, err := client.AddSuite(tt.projectID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, suite)
				assert.Equal(t, tt.request.Name, suite.Name)
			}
		})
	}
}

func TestUpdateSuite(t *testing.T) {
	tests := []struct {
		name         string
		suiteID      int64
		request      *data.UpdateSuiteRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:    "successful update",
			suiteID: 100,
			request: &data.UpdateSuiteRequest{
				Name:        "Updated Suite Name",
				Description: "Updated description",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Suite{
				ID:          100,
				Name:        "Updated Suite Name",
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name:         "suite not found",
			suiteID:      99999,
			request:      &data.UpdateSuiteRequest{Name: "Test"},
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/update_suite/%d", tt.suiteID)
				assert.Equal(t, expectedPath, r.URL.String())

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			suite, err := client.UpdateSuite(tt.suiteID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, suite)
			}
		})
	}
}

func TestDeleteSuite(t *testing.T) {
	tests := []struct {
		name       string
		suiteID    int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			suiteID:    100,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "suite not found",
			suiteID:    99999,
			mockStatus: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "suite has cases",
			suiteID:    100,
			mockStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/delete_suite/%d", tt.suiteID)
				assert.Equal(t, expectedPath, r.URL.String())

				w.WriteHeader(tt.mockStatus)
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			err := client.DeleteSuite(tt.suiteID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
