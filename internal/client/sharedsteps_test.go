// internal/client/sharedsteps_test.go
// Тесты для SharedSteps API POST-методов
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddSharedStep(t *testing.T) {
	tests := []struct {
		name         string
		projectID    int64
		request      *data.AddSharedStepRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
		expectedErr  string
	}{
		{
			name:      "successful shared step creation",
			projectID: 30,
			request: &data.AddSharedStepRequest{
				Title: "Login Steps",
				CustomStepsSeparated: []data.Step{
					{
						Content:  "Enter username",
						Expected: "Username field is filled",
					},
					{
						Content:  "Enter password",
						Expected: "Password field is filled",
					},
					{
						Content:  "Click Login button",
						Expected: "User is logged in",
					},
				},
			},
			mockStatus: http.StatusOK,
			mockResponse: data.SharedStep{
				ID:        200,
				ProjectID: 30,
				Title:     "Login Steps",
				CustomStepsSeparated: []data.Step{
					{Content: "Enter username", Expected: "Username field is filled"},
					{Content: "Enter password", Expected: "Password field is filled"},
					{Content: "Click Login button", Expected: "User is logged in"},
				},
				CreatedBy: 10,
				CreatedOn: 1707000000,
			},
			wantErr: false,
		},
		{
			name:      "minimal shared step",
			projectID: 30,
			request: &data.AddSharedStepRequest{
				Title: "Simple Step",
				CustomStepsSeparated: []data.Step{
					{Content: "Do something", Expected: "Something happens"},
				},
			},
			mockStatus: http.StatusOK,
			mockResponse: data.SharedStep{
				ID:        201,
				ProjectID: 30,
				Title:     "Simple Step",
				CustomStepsSeparated: []data.Step{
					{Content: "Do something", Expected: "Something happens"},
				},
			},
			wantErr: false,
		},
		{
			name:         "project not found",
			projectID:    99999,
			request:      &data.AddSharedStepRequest{Title: "Test", CustomStepsSeparated: []data.Step{{Content: "Test"}}},
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
			expectedErr:  "404",
		},
		{
			name:         "empty title",
			projectID:    30,
			request:      &data.AddSharedStepRequest{Title: "", CustomStepsSeparated: []data.Step{{Content: "Test"}}},
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
				expectedPath := fmt.Sprintf("/index.php?/api/v2/add_shared_step/%d", tt.projectID)
				assert.Equal(t, expectedPath, r.URL.String())
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var req data.AddSharedStepRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, tt.request.Title, req.Title)
				assert.Len(t, req.CustomStepsSeparated, len(tt.request.CustomStepsSeparated))

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			step, err := client.AddSharedStep(tt.projectID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, step)
				assert.Equal(t, tt.request.Title, step.Title)
				assert.Len(t, step.CustomStepsSeparated, len(tt.request.CustomStepsSeparated))
			}
		})
	}
}

func TestUpdateSharedStep(t *testing.T) {
	tests := []struct {
		name         string
		sharedStepID int64
		request      *data.UpdateSharedStepRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:         "successful update",
			sharedStepID: 200,
			request: &data.UpdateSharedStepRequest{
				Title: "Updated Shared Step",
				CustomStepsSeparated: []data.Step{
					{Content: "Updated step", Expected: "Updated result"},
				},
			},
			mockStatus: http.StatusOK,
			mockResponse: data.SharedStep{
				ID:    200,
				Title: "Updated Shared Step",
				CustomStepsSeparated: []data.Step{
					{Content: "Updated step", Expected: "Updated result"},
				},
			},
			wantErr: false,
		},
		{
			name:         "shared step not found",
			sharedStepID: 99999,
			request:      &data.UpdateSharedStepRequest{Title: "Test"},
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/update_shared_step/%d", tt.sharedStepID)
				assert.Equal(t, expectedPath, r.URL.String())

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			step, err := client.UpdateSharedStep(tt.sharedStepID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, step)
			}
		})
	}
}

func TestDeleteSharedStep(t *testing.T) {
	tests := []struct {
		name         string
		sharedStepID int64
		mockStatus   int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			sharedStepID: 200,
			mockStatus:   http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "shared step not found",
			sharedStepID: 99999,
			mockStatus:   http.StatusNotFound,
			wantErr:      true,
		},
		{
			name:         "shared step in use",
			sharedStepID: 200,
			mockStatus:   http.StatusBadRequest,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/delete_shared_step/%d&keep_in_cases=0", tt.sharedStepID)
				assert.Equal(t, expectedPath, r.URL.String())

				w.WriteHeader(tt.mockStatus)
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			err := client.DeleteSharedStep(tt.sharedStepID, 0)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
