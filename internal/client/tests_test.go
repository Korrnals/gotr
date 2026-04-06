// internal/client/tests_test.go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestGetTest(t *testing.T) {
	tests := []struct {
		name       string
		testID     int64
		mockResp   *data.Test
		mockStatus int
		wantErr    bool
	}{
		{
			name:   "successful test retrieval",
			testID: 12345,
			mockResp: &data.Test{
				ID:         12345,
				CaseID:     100,
				RunID:      50,
				Title:      "Test Login",
				StatusID:   1,
				AssignedTo: 5,
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "test not found",
			testID:     99999,
			mockResp:   nil,
			mockStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}

				w.WriteHeader(tt.mockStatus)
				if tt.mockResp != nil {
					json.NewEncoder(w).Encode(tt.mockResp)
				}
			}))
			defer server.Close()

			client, _ := NewClient(server.URL, "test@test.com", "testpass", false)
			ctx := context.Background()
			test, err := client.GetTest(ctx, tt.testID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if test.ID != tt.mockResp.ID {
				t.Errorf("expected test ID %d, got %d", tt.mockResp.ID, test.ID)
			}

			if test.Title != tt.mockResp.Title {
				t.Errorf("expected title %s, got %s", tt.mockResp.Title, test.Title)
			}
		})
	}
}

func TestGetTests(t *testing.T) {
	tests := []struct {
		name       string
		runID      int64
		filters    map[string]string
		mockResp   []data.Test
		mockStatus int
		wantErr    bool
	}{
		{
			name:    "successful tests retrieval without filters",
			runID:   100,
			filters: nil,
			mockResp: []data.Test{
				{ID: 1, CaseID: 10, Title: "Test 1", StatusID: 1},
				{ID: 2, CaseID: 20, Title: "Test 2", StatusID: 5},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:    "successful tests retrieval with status filter",
			runID:   100,
			filters: map[string]string{"status_id": "1"},
			mockResp: []data.Test{
				{ID: 1, CaseID: 10, Title: "Test 1", StatusID: 1},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "run not found",
			runID:      99999,
			filters:    nil,
			mockResp:   nil,
			mockStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}

				if tt.filters != nil {
					for key, value := range tt.filters {
						if r.URL.Query().Get(key) != value {
							t.Errorf("expected query param %s=%s", key, value)
						}
					}
				}

				w.WriteHeader(tt.mockStatus)
				if tt.mockResp != nil {
					json.NewEncoder(w).Encode(tt.mockResp)
				}
			}))
			defer server.Close()

			client, _ := NewClient(server.URL, "test@test.com", "testpass", false)
			ctx := context.Background()
			tests, err := client.GetTests(ctx, tt.runID, tt.filters)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(tests) != len(tt.mockResp) {
				t.Errorf("expected %d tests, got %d", len(tt.mockResp), len(tests))
			}
		})
	}
}

func TestUpdateTest(t *testing.T) {
	tests := []struct {
		name       string
		testID     int64
		request    *data.UpdateTestRequest
		mockResp   *data.Test
		mockStatus int
		wantErr    bool
	}{
		{
			name:   "successful test update - status",
			testID: 12345,
			request: &data.UpdateTestRequest{
				StatusID: 5, // Failed
			},
			mockResp: &data.Test{
				ID:       12345,
				CaseID:   100,
				Title:    "Test Login",
				StatusID: 5,
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "successful test update - assignee",
			testID: 12345,
			request: &data.UpdateTestRequest{
				AssignedTo: 10,
			},
			mockResp: &data.Test{
				ID:         12345,
				CaseID:     100,
				Title:      "Test Login",
				AssignedTo: 10,
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "nil request",
			testID:     12345,
			request:    nil,
			mockResp:   nil,
			mockStatus: http.StatusOK,
			wantErr:    true,
		},
		{
			name:   "test not found",
			testID: 99999,
			request: &data.UpdateTestRequest{
				StatusID: 1,
			},
			mockResp:   nil,
			mockStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}

				w.WriteHeader(tt.mockStatus)
				if tt.mockResp != nil {
					json.NewEncoder(w).Encode(tt.mockResp)
				}
			}))
			defer server.Close()

			client, _ := NewClient(server.URL, "test@test.com", "testpass", false)
			ctx := context.Background()
			test, err := client.UpdateTest(ctx, tt.testID, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if test.ID != tt.mockResp.ID {
				t.Errorf("expected test ID %d, got %d", tt.mockResp.ID, test.ID)
			}

			if tt.request.StatusID > 0 && test.StatusID != tt.request.StatusID {
				t.Errorf("expected status %d, got %d", tt.request.StatusID, test.StatusID)
			}

			if tt.request.AssignedTo > 0 && test.AssignedTo != tt.request.AssignedTo {
				t.Errorf("expected assignedto %d, got %d", tt.request.AssignedTo, test.AssignedTo)
			}
		})
	}
}

func TestGetTestsHelpers(t *testing.T) {
	t.Run("GetTestsByStatus", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php", r.URL.Path)
			assert.Contains(t, r.URL.String(), "get_tests/100")
			assert.Equal(t, "1", r.URL.Query().Get("status_id"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]data.Test{{ID: 1, StatusID: 1}})
		})
		defer server.Close()

		tests, err := client.GetTestsByStatus(context.Background(), 100, 1)
		assert.NoError(t, err)
		assert.Len(t, tests, 1)
	})

	t.Run("GetTestsAssignedTo", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php", r.URL.Path)
			assert.Contains(t, r.URL.String(), "get_tests/100")
			assert.Equal(t, "7", r.URL.Query().Get("assignedto_id"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]data.Test{{ID: 2, AssignedTo: 7}})
		})
		defer server.Close()

		tests, err := client.GetTestsAssignedTo(context.Background(), 100, 7)
		assert.NoError(t, err)
		assert.Len(t, tests, 1)
	})
}

func TestUpdateTest_HTTPErrorAndDecodeBranches(t *testing.T) {
	t.Run("decode error on 200", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.URL.String(), "update_test/42")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer server.Close()

		_, err := client.UpdateTest(context.Background(), 42, &data.UpdateTestRequest{StatusID: 5})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error test")
	})

	t.Run("request error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		server.Close()

		_, err := client.UpdateTest(context.Background(), 42, &data.UpdateTestRequest{AssignedTo: 7})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request error UpdateTest")
	})
}
