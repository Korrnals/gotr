// internal/client/plans_test.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func TestGetPlan(t *testing.T) {
	testCases := []struct {
		name         string
		planID       int64
		mockResponse *data.Plan
		wantErr      bool
	}{
		{
			name:   "Success",
			planID: 1,
			mockResponse: &data.Plan{
				ID:   1,
				Name: "Sprint 1 Plan",
			},
		},
		{
			name:    "PlanNotFound",
			planID:  999,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			if tc.mockResponse != nil {
				mockClient.GetPlanFunc = func(ctx context.Context, id int64) (*data.Plan, error) {
					return tc.mockResponse, nil
				}
			} else {
				mockClient.GetPlanFunc = func(ctx context.Context, id int64) (*data.Plan, error) {
					return nil, fmt.Errorf("plan %d not found", id)
				}
			}

			ctx := context.Background()
			result, err := mockClient.GetPlan(ctx, tc.planID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetPlan() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetPlan() unexpected error: %v", err)
				return
			}
			if result.ID != tc.mockResponse.ID {
				t.Errorf("GetPlan() ID = %d, want %d", result.ID, tc.mockResponse.ID)
			}
		})
	}
}

func TestGetPlans(t *testing.T) {
	testCases := []struct {
		name         string
		projectID    int64
		mockResponse data.GetPlansResponse
		wantErr      bool
		wantCount    int
	}{
		{
			name:      "Success with plans",
			projectID: 1,
			mockResponse: []data.Plan{
				{ID: 1, Name: "Plan 1", ProjectID: 1},
				{ID: 2, Name: "Plan 2", ProjectID: 1},
			},
			wantCount: 2,
		},
		{
			name:         "Success empty",
			projectID:    2,
			mockResponse: []data.Plan{},
			wantCount:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.GetPlansFunc = func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.GetPlans(ctx, tc.projectID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetPlans() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetPlans() unexpected error: %v", err)
				return
			}
			if len(result) != tc.wantCount {
				t.Errorf("GetPlans() returned %d plans, want %d", len(result), tc.wantCount)
			}
		})
	}
}

func TestAddPlan(t *testing.T) {
	testCases := []struct {
		name         string
		projectID    int64
		request      *data.AddPlanRequest
		mockResponse *data.Plan
		wantErr      bool
	}{
		{
			name:      "Success",
			projectID: 1,
			request: &data.AddPlanRequest{
				Name: "New Plan",
			},
			mockResponse: &data.Plan{
				ID:   3,
				Name: "New Plan",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.AddPlanFunc = func(ctx context.Context, projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.AddPlan(ctx, tc.projectID, tc.request)

			if tc.wantErr {
				if err == nil {
					t.Errorf("AddPlan() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("AddPlan() unexpected error: %v", err)
				return
			}
			if result.ID != tc.mockResponse.ID {
				t.Errorf("AddPlan() ID = %d, want %d", result.ID, tc.mockResponse.ID)
			}
		})
	}
}

func TestUpdatePlan(t *testing.T) {
	testCases := []struct {
		name         string
		planID       int64
		request      *data.UpdatePlanRequest
		mockResponse *data.Plan
		wantErr      bool
	}{
		{
			name:   "Success",
			planID: 1,
			request: &data.UpdatePlanRequest{
				Name: "Updated Plan",
			},
			mockResponse: &data.Plan{
				ID:   1,
				Name: "Updated Plan",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.UpdatePlanFunc = func(ctx context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.UpdatePlan(ctx, tc.planID, tc.request)

			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdatePlan() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdatePlan() unexpected error: %v", err)
				return
			}
			if result.Name != tc.mockResponse.Name {
				t.Errorf("UpdatePlan() Name = %s, want %s", result.Name, tc.mockResponse.Name)
			}
		})
	}
}

func TestClosePlan(t *testing.T) {
	testCases := []struct {
		name         string
		planID       int64
		mockResponse *data.Plan
		wantErr      bool
	}{
		{
			name:   "Success",
			planID: 1,
			mockResponse: &data.Plan{
				ID:          1,
				Name:        "Closed Plan",
				IsCompleted: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.ClosePlanFunc = func(ctx context.Context, planID int64) (*data.Plan, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.ClosePlan(ctx, tc.planID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("ClosePlan() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("ClosePlan() unexpected error: %v", err)
				return
			}
			if !result.IsCompleted {
				t.Errorf("ClosePlan() IsCompleted = false, want true")
			}
		})
	}
}

func TestDeletePlan(t *testing.T) {
	testCases := []struct {
		name    string
		planID  int64
		wantErr bool
	}{
		{
			name:   "Success",
			planID: 1,
		},
		{
			name:    "NotFound",
			planID:  999,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			if tc.wantErr {
				mockClient.DeletePlanFunc = func(ctx context.Context, planID int64) error {
					return fmt.Errorf("plan %d not found", planID)
				}
			} else {
				mockClient.DeletePlanFunc = func(ctx context.Context, planID int64) error {
					return nil
				}
			}

			ctx := context.Background()
			err := mockClient.DeletePlan(ctx, tc.planID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("DeletePlan() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("DeletePlan() unexpected error: %v", err)
			}
		})
	}
}

func TestAddPlanEntry(t *testing.T) {
	testCases := []struct {
		name         string
		planID       int64
		request      *data.AddPlanEntryRequest
		mockResponse *data.Plan
		wantErr      bool
	}{
		{
			name:   "Success",
			planID: 1,
			request: &data.AddPlanEntryRequest{
				Name:    "New Entry",
				SuiteID: 1,
			},
			mockResponse: &data.Plan{
				ID:   1,
				Name: "Plan with Entry",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.AddPlanEntryFunc = func(ctx context.Context, planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.AddPlanEntry(ctx, tc.planID, tc.request)

			if tc.wantErr {
				if err == nil {
					t.Errorf("AddPlanEntry() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("AddPlanEntry() unexpected error: %v", err)
				return
			}
			if result.ID != tc.mockResponse.ID {
				t.Errorf("AddPlanEntry() ID = %d, want %d", result.ID, tc.mockResponse.ID)
			}
		})
	}
}

func TestUpdatePlanEntry(t *testing.T) {
	testCases := []struct {
		name         string
		planID       int64
		entryID      string
		request      *data.UpdatePlanEntryRequest
		mockResponse *data.Plan
		wantErr      bool
	}{
		{
			name:    "Success",
			planID:  1,
			entryID: "entry-1",
			request: &data.UpdatePlanEntryRequest{
				Name: "Updated Entry",
			},
			mockResponse: &data.Plan{
				ID:   1,
				Name: "Plan with Updated Entry",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.UpdatePlanEntryFunc = func(ctx context.Context, planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.UpdatePlanEntry(ctx, tc.planID, tc.entryID, tc.request)

			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdatePlanEntry() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdatePlanEntry() unexpected error: %v", err)
				return
			}
			if result.ID != tc.mockResponse.ID {
				t.Errorf("UpdatePlanEntry() ID = %d, want %d", result.ID, tc.mockResponse.ID)
			}
		})
	}
}

func TestDeletePlanEntry(t *testing.T) {
	testCases := []struct {
		name    string
		planID  int64
		entryID string
		wantErr bool
	}{
		{
			name:    "Success",
			planID:  1,
			entryID: "entry-1",
		},
		{
			name:    "NotFound",
			planID:  1,
			entryID: "entry-999",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			if tc.wantErr {
				mockClient.DeletePlanEntryFunc = func(ctx context.Context, planID int64, entryID string) error {
					return fmt.Errorf("entry %s not found", entryID)
				}
			} else {
				mockClient.DeletePlanEntryFunc = func(ctx context.Context, planID int64, entryID string) error {
					return nil
				}
			}

			ctx := context.Background()
			err := mockClient.DeletePlanEntry(ctx, tc.planID, tc.entryID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("DeletePlanEntry() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("DeletePlanEntry() unexpected error: %v", err)
			}
		})
	}
}

// HTTP integration tests
func TestHTTPGetPlan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.Plan{
			ID:   1,
			Name: "Test Plan",
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	ctx := context.Background()
	plan, err := client.GetPlan(ctx, 1)
	if err != nil {
		t.Fatalf("GetPlan() error: %v", err)
	}
	if plan.Name != "Test Plan" {
		t.Errorf("Expected name 'Test Plan', got '%s'", plan.Name)
	}
}

func TestHTTPGetPlans(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]data.Plan{
			{ID: 1, Name: "Plan 1"},
			{ID: 2, Name: "Plan 2"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	ctx := context.Background()
	plans, err := client.GetPlans(ctx, 1)
	if err != nil {
		t.Fatalf("GetPlans() error: %v", err)
	}
	if len(plans) != 2 {
		t.Errorf("Expected 2 plans, got %d", len(plans))
	}
}

func TestHTTPAddPlan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.Plan{
			ID:   3,
			Name: "New Plan",
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	req := &data.AddPlanRequest{
		Name: "New Plan",
	}

	ctx := context.Background()
	plan, err := client.AddPlan(ctx, 1, req)
	if err != nil {
		t.Fatalf("AddPlan() error: %v", err)
	}
	if plan.Name != "New Plan" {
		t.Errorf("Expected name 'New Plan', got '%s'", plan.Name)
	}
}

func TestHTTPDeletePlan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	ctx := context.Background()
	err := client.DeletePlan(ctx, 1)
	if err != nil {
		t.Fatalf("DeletePlan() error: %v", err)
	}
}

func TestHTTPUpdatePlan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.String(), "update_plan/7") {
				t.Fatalf("expected update_plan/7 endpoint, got %s", r.URL.String())
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Plan{ID: 7, Name: "Updated"})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		plan, err := client.UpdatePlan(context.Background(), 7, &data.UpdatePlanRequest{Name: "Updated"})
		if err != nil {
			t.Fatalf("UpdatePlan() error: %v", err)
		}
		if plan.Name != "Updated" {
			t.Fatalf("unexpected plan payload: %+v", plan)
		}
	})

	t.Run("non-200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"bad update"}`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		_, err := client.UpdatePlan(context.Background(), 7, &data.UpdatePlanRequest{Name: "Updated"})
		if err == nil {
			t.Fatalf("expected UpdatePlan() error for non-200 status")
		}
	})
}

func TestHTTPClosePlan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.String(), "close_plan/8") {
				t.Fatalf("expected close_plan/8 endpoint, got %s", r.URL.String())
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Plan{ID: 8, IsCompleted: true})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		plan, err := client.ClosePlan(context.Background(), 8)
		if err != nil {
			t.Fatalf("ClosePlan() error: %v", err)
		}
		if !plan.IsCompleted {
			t.Fatalf("expected closed plan to be completed")
		}
	})
}

func TestHTTPPlanEntries(t *testing.T) {
	t.Run("add entry", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.String(), "add_plan_entry/11") {
				t.Fatalf("expected add_plan_entry/11 endpoint, got %s", r.URL.String())
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Plan{ID: 11, Name: "With Entry"})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		plan, err := client.AddPlanEntry(context.Background(), 11, &data.AddPlanEntryRequest{Name: "E1", SuiteID: 1})
		if err != nil {
			t.Fatalf("AddPlanEntry() error: %v", err)
		}
		if plan.ID != 11 {
			t.Fatalf("unexpected plan payload: %+v", plan)
		}
	})

	t.Run("update entry", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.String(), "update_plan_entry/11/entry-1") {
				t.Fatalf("unexpected endpoint: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Plan{ID: 11, Name: "Updated Entry"})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		plan, err := client.UpdatePlanEntry(context.Background(), 11, "entry-1", &data.UpdatePlanEntryRequest{Name: "Updated"})
		if err != nil {
			t.Fatalf("UpdatePlanEntry() error: %v", err)
		}
		if plan.ID != 11 {
			t.Fatalf("unexpected plan payload: %+v", plan)
		}
	})

	t.Run("delete entry", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.String(), "delete_plan_entry/11/entry-1") {
				t.Fatalf("unexpected endpoint: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		if err := client.DeletePlanEntry(context.Background(), 11, "entry-1"); err != nil {
			t.Fatalf("DeletePlanEntry() error: %v", err)
		}
	})
}

func TestHTTPPlans_ErrorBranches(t *testing.T) {
	t.Run("get plan non-OK and decode", func(t *testing.T) {
		t.Run("non-OK", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"missing plan"}`))
			}))
			defer server.Close()

			client, _ := NewClient(server.URL, "test", "test", false)
			_, err := client.GetPlan(context.Background(), 0)
			if err == nil {
				t.Fatalf("expected GetPlan() error for non-OK status")
			}
		})

		t.Run("decode", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{"))
			}))
			defer server.Close()

			client, _ := NewClient(server.URL, "test", "test", false)
			_, err := client.GetPlan(context.Background(), 0)
			if err == nil {
				t.Fatalf("expected GetPlan() decode error")
			}
			if !strings.Contains(err.Error(), "error decoding plan") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	})

	t.Run("get plans request/non-OK", func(t *testing.T) {
		t.Run("request error", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			client, _ := NewClient(server.URL, "test", "test", false)
			server.Close()

			_, err := client.GetPlans(context.Background(), 0)
			if err == nil {
				t.Fatalf("expected GetPlans() request error")
			}
		})

		t.Run("non-OK", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"bad get plans"}`))
			}))
			defer server.Close()

			client, _ := NewClient(server.URL, "test", "test", false)
			_, err := client.GetPlans(context.Background(), 0)
			if err == nil {
				t.Fatalf("expected GetPlans() non-OK error")
			}
		})
	})

	t.Run("post-based plan methods non-OK", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"bad post"}`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)

		if _, err := client.AddPlan(context.Background(), 0, &data.AddPlanRequest{Name: ""}); err == nil {
			t.Fatalf("expected AddPlan() non-OK error")
		}
		if _, err := client.UpdatePlan(context.Background(), 0, &data.UpdatePlanRequest{Name: ""}); err == nil {
			t.Fatalf("expected UpdatePlan() non-OK error")
		}
		if _, err := client.ClosePlan(context.Background(), 0); err == nil {
			t.Fatalf("expected ClosePlan() non-OK error")
		}
		if err := client.DeletePlan(context.Background(), 0); err == nil {
			t.Fatalf("expected DeletePlan() non-OK error")
		}
		if _, err := client.AddPlanEntry(context.Background(), 0, &data.AddPlanEntryRequest{Name: "", SuiteID: 0}); err == nil {
			t.Fatalf("expected AddPlanEntry() non-OK error")
		}
		if _, err := client.UpdatePlanEntry(context.Background(), 0, "", &data.UpdatePlanEntryRequest{Name: ""}); err == nil {
			t.Fatalf("expected UpdatePlanEntry() non-OK error")
		}
		if err := client.DeletePlanEntry(context.Background(), 0, ""); err == nil {
			t.Fatalf("expected DeletePlanEntry() non-OK error")
		}
	})

	t.Run("post-based plan methods decode errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.String(), "delete_plan/") || strings.Contains(r.URL.String(), "delete_plan_entry/") {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)

		if _, err := client.AddPlan(context.Background(), 1, &data.AddPlanRequest{Name: "A"}); err == nil {
			t.Fatalf("expected AddPlan() decode error")
		}
		if _, err := client.UpdatePlan(context.Background(), 1, &data.UpdatePlanRequest{Name: "A"}); err == nil {
			t.Fatalf("expected UpdatePlan() decode error")
		}
		if _, err := client.ClosePlan(context.Background(), 1); err == nil {
			t.Fatalf("expected ClosePlan() decode error")
		}
		if _, err := client.AddPlanEntry(context.Background(), 1, &data.AddPlanEntryRequest{Name: "A", SuiteID: 1}); err == nil {
			t.Fatalf("expected AddPlanEntry() decode error")
		}
		if _, err := client.UpdatePlanEntry(context.Background(), 1, "", &data.UpdatePlanEntryRequest{Name: "A"}); err == nil {
			t.Fatalf("expected UpdatePlanEntry() decode error")
		}
	})

	t.Run("post-based plan methods request errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		client, _ := NewClient(server.URL, "test", "test", false)
		server.Close()

		if _, err := client.AddPlan(context.Background(), 1, &data.AddPlanRequest{Name: "A"}); err == nil {
			t.Fatalf("expected AddPlan() request error")
		}
		if _, err := client.UpdatePlan(context.Background(), 1, &data.UpdatePlanRequest{Name: "A"}); err == nil {
			t.Fatalf("expected UpdatePlan() request error")
		}
		if _, err := client.ClosePlan(context.Background(), 1); err == nil {
			t.Fatalf("expected ClosePlan() request error")
		}
		if err := client.DeletePlan(context.Background(), 1); err == nil {
			t.Fatalf("expected DeletePlan() request error")
		}
		if _, err := client.AddPlanEntry(context.Background(), 1, &data.AddPlanEntryRequest{Name: "A", SuiteID: 1}); err == nil {
			t.Fatalf("expected AddPlanEntry() request error")
		}
		if _, err := client.UpdatePlanEntry(context.Background(), 1, "entry", &data.UpdatePlanEntryRequest{Name: "A"}); err == nil {
			t.Fatalf("expected UpdatePlanEntry() request error")
		}
		if err := client.DeletePlanEntry(context.Background(), 1, "entry"); err == nil {
			t.Fatalf("expected DeletePlanEntry() request error")
		}
	})
}
