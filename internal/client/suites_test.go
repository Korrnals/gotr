// internal/client/suites_test.go
// Tests for Suites API POST methods
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

			ctx := context.Background()
			suite, err := client.AddSuite(ctx, tt.projectID, tt.request)

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

			ctx := context.Background()
			suite, err := client.UpdateSuite(ctx, tt.suiteID, tt.request)

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

			ctx := context.Background()
			err := client.DeleteSuite(ctx, tt.suiteID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSuites(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/index.php", r.URL.Path)
		assert.Equal(t, "0", r.URL.Query().Get("offset"))
		assert.Equal(t, "250", r.URL.Query().Get("limit"))
		assert.Contains(t, r.URL.String(), "get_suites/30")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.GetSuitesResponse{{ID: 1, Name: "Smoke"}})
	}

	client, server := mockClient(t, handler)
	defer server.Close()

	resp, err := client.GetSuites(context.Background(), 30)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, int64(1), resp[0].ID)
}

func TestGetSuite(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_suite/100", r.URL.String())
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.Suite{ID: 100, Name: "Regression"})
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		suite, err := client.GetSuite(context.Background(), 100)
		assert.NoError(t, err)
		assert.NotNil(t, suite)
		assert.Equal(t, int64(100), suite.ID)
	})

	t.Run("api error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("missing"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetSuite(context.Background(), 100)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned")
	})

	t.Run("decode error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetSuite(context.Background(), 100)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error suite")
	})
}
