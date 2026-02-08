// internal/client/sections_test.go
// Тесты для Sections API POST-методов
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddSection(t *testing.T) {
	tests := []struct {
		name         string
		projectID    int64
		request      *data.AddSectionRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
		expectedErr  string
	}{
		{
			name:      "successful section creation",
			projectID: 30,
			request: &data.AddSectionRequest{
				SuiteID:     100,
				ParentID:    0,
				Name:        "Login Tests",
				Description: "All login-related test cases",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Section{
				ID:          500,
				SuiteID:     100,
				Name:        "Login Tests",
				Description: "All login-related test cases",
				DisplayOrder: 1,
				Depth:       0,
			},
			wantErr: false,
		},
		{
			name:      "nested section",
			projectID: 30,
			request: &data.AddSectionRequest{
				SuiteID:  100,
				ParentID: 500,
				Name:     "OAuth Login",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Section{
				ID:       501,
				SuiteID:  100,
				ParentID: 500,
				Name:     "OAuth Login",
				Depth:    1,
			},
			wantErr: false,
		},
		{
			name:      "minimal section",
			projectID: 30,
			request: &data.AddSectionRequest{
				SuiteID: 100,
				Name:    "Quick Section",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Section{
				ID:      502,
				SuiteID: 100,
				Name:    "Quick Section",
			},
			wantErr: false,
		},
		{
			name:         "project not found",
			projectID:    99999,
			request:      &data.AddSectionRequest{SuiteID: 100, Name: "Test"},
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
			expectedErr:  "404",
		},
		{
			name:         "invalid suite",
			projectID:    30,
			request:      &data.AddSectionRequest{SuiteID: 99999, Name: "Test"},
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
				expectedPath := fmt.Sprintf("/index.php?/api/v2/add_section/%d", tt.projectID)
				assert.Equal(t, expectedPath, r.URL.String())
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var req data.AddSectionRequest
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

			section, err := client.AddSection(tt.projectID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, section)
				assert.Equal(t, tt.request.Name, section.Name)
			}
		})
	}
}

func TestUpdateSection(t *testing.T) {
	tests := []struct {
		name         string
		sectionID    int64
		request      *data.UpdateSectionRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:      "successful update",
			sectionID: 500,
			request: &data.UpdateSectionRequest{
				Name:        "Updated Section",
				Description: "Updated description",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Section{
				ID:          500,
				Name:        "Updated Section",
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name:         "section not found",
			sectionID:    99999,
			request:      &data.UpdateSectionRequest{Name: "Test"},
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/update_section/%d", tt.sectionID)
					assert.Equal(t, expectedPath, r.URL.String())

				var req data.UpdateSectionRequest
				json.NewDecoder(r.Body).Decode(&req)

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			section, err := client.UpdateSection(tt.sectionID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, section)
			}
		})
	}
}

func TestDeleteSection(t *testing.T) {
	tests := []struct {
		name       string
		sectionID  int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			sectionID:  500,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "section not found",
			sectionID:  99999,
			mockStatus: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "section has cases",
			sectionID:  500,
			mockStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/delete_section/%d", tt.sectionID)
				assert.Equal(t, expectedPath, r.URL.String())

				w.WriteHeader(tt.mockStatus)
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			err := client.DeleteSection(tt.sectionID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
