// internal/client/sharedsteps_test.go
// Tests for SharedSteps API POST methods
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

			ctx := context.Background()
			step, err := client.AddSharedStep(ctx, tt.projectID, tt.request)

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

			ctx := context.Background()
			step, err := client.UpdateSharedStep(ctx, tt.sharedStepID, tt.request)

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

			ctx := context.Background()
			err := client.DeleteSharedStep(ctx, tt.sharedStepID, 0)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSharedSteps(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/index.php", r.URL.Path)
		assert.Equal(t, "0", r.URL.Query().Get("offset"))
		assert.Equal(t, "250", r.URL.Query().Get("limit"))
		assert.Contains(t, r.URL.String(), "get_shared_steps/30")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.GetSharedStepsResponse{{ID: 11, Title: "Login"}})
	}

	client, server := mockClient(t, handler)
	defer server.Close()

	steps, err := client.GetSharedSteps(context.Background(), 30)
	assert.NoError(t, err)
	assert.Len(t, steps, 1)
	assert.Equal(t, int64(11), steps[0].ID)
}

func TestGetSharedStep(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_shared_step/200", r.URL.String())
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.SharedStep{ID: 200, Title: "Step"})
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		step, err := client.GetSharedStep(context.Background(), 200)
		assert.NoError(t, err)
		assert.NotNil(t, step)
		assert.Equal(t, int64(200), step.ID)
	})

	t.Run("api error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("missing"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetSharedStep(context.Background(), 200)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned")
	})
}

func TestGetSharedStepHistory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_shared_step_history/200", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"offset":0,"limit":250,"size":1,"step_history":[{"id":1,"timestamp":10,"user_id":5,"title":"v1"}]}`))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		history, err := client.GetSharedStepHistory(context.Background(), 200)
		assert.NoError(t, err)
		assert.NotNil(t, history)
		assert.Len(t, history.History, 1)
	})

	t.Run("decode error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetSharedStepHistory(context.Background(), 200)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error shared step history")
	})
}

func TestSharedSteps_ErrorBranches(t *testing.T) {
	t.Run("GetSharedSteps request error from API status", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.String(), "get_shared_steps/30")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		})
		defer server.Close()

		_, err := client.GetSharedSteps(context.Background(), 30)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request error GetSharedSteps")
	})

	t.Run("GetSharedStep decode error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_shared_step/200", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"broken":`))
		})
		defer server.Close()

		_, err := client.GetSharedStep(context.Background(), 200)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error shared step")
	})

	t.Run("GetSharedStepHistory non-OK", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_shared_step_history/201", r.URL.String())
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("missing"))
		})
		defer server.Close()

		_, err := client.GetSharedStepHistory(context.Background(), 201)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned 404 Not Found")
	})

	t.Run("AddSharedStep decode error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/add_shared_step/30", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"broken":`))
		})
		defer server.Close()

		_, err := client.AddSharedStep(context.Background(), 30, &data.AddSharedStepRequest{Title: "t"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error created shared step")
	})

	t.Run("UpdateSharedStep decode error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/update_shared_step/31", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"broken":`))
		})
		defer server.Close()

		_, err := client.UpdateSharedStep(context.Background(), 31, &data.UpdateSharedStepRequest{Title: "u"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error updated shared step")
	})

	t.Run("DeleteSharedStep query parameter keep_in_cases", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/delete_shared_step/32&keep_in_cases=1", r.URL.String())
			w.WriteHeader(http.StatusOK)
		})
		defer server.Close()

		err := client.DeleteSharedStep(context.Background(), 32, 1)
		assert.NoError(t, err)
	})
}

func TestSharedSteps_RequestErrors(t *testing.T) {
	t.Run("GetSharedStep request error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server.Close()

		_, err := client.GetSharedStep(context.Background(), 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request error GetSharedStep")
	})

	t.Run("GetSharedStepHistory request error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server.Close()

		_, err := client.GetSharedStepHistory(context.Background(), 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request error GetSharedStepHistory")
	})

	t.Run("AddSharedStep request error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server.Close()

		_, err := client.AddSharedStep(context.Background(), 3, &data.AddSharedStepRequest{Title: "x"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request error AddSharedStep")
	})

	t.Run("UpdateSharedStep request error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server.Close()

		_, err := client.UpdateSharedStep(context.Background(), 4, &data.UpdateSharedStepRequest{Title: "x"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request error UpdateSharedStep")
	})

	t.Run("DeleteSharedStep request error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server.Close()

		err := client.DeleteSharedStep(context.Background(), 5, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request error DeleteSharedStep")
	})
}
