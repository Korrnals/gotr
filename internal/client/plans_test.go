// internal/client/plans_test.go
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
				mockClient.GetPlanFunc = func(id int64) (*data.Plan, error) {
					return tc.mockResponse, nil
				}
			} else {
				mockClient.GetPlanFunc = func(id int64) (*data.Plan, error) {
					return nil, fmt.Errorf("plan %d not found", id)
				}
			}

			result, err := mockClient.GetPlan(tc.planID)

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
			mockClient.GetPlansFunc = func(projectID int64) (data.GetPlansResponse, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.GetPlans(tc.projectID)

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
			mockClient.AddPlanFunc = func(projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.AddPlan(tc.projectID, tc.request)

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
			mockClient.UpdatePlanFunc = func(planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.UpdatePlan(tc.planID, tc.request)

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
			mockClient.ClosePlanFunc = func(planID int64) (*data.Plan, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.ClosePlan(tc.planID)

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
		name   string
		planID int64
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
				mockClient.DeletePlanFunc = func(planID int64) error {
					return fmt.Errorf("plan %d not found", planID)
				}
			} else {
				mockClient.DeletePlanFunc = func(planID int64) error {
					return nil
				}
			}

			err := mockClient.DeletePlan(tc.planID)

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
			mockClient.AddPlanEntryFunc = func(planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.AddPlanEntry(tc.planID, tc.request)

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
			mockClient.UpdatePlanEntryFunc = func(planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.UpdatePlanEntry(tc.planID, tc.entryID, tc.request)

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
				mockClient.DeletePlanEntryFunc = func(planID int64, entryID string) error {
					return fmt.Errorf("entry %s not found", entryID)
				}
			} else {
				mockClient.DeletePlanEntryFunc = func(planID int64, entryID string) error {
					return nil
				}
			}

			err := mockClient.DeletePlanEntry(tc.planID, tc.entryID)

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

	plan, err := client.GetPlan(1)
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

	plans, err := client.GetPlans(1)
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

	plan, err := client.AddPlan(1, req)
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

	err := client.DeletePlan(1)
	if err != nil {
		t.Fatalf("DeletePlan() error: %v", err)
	}
}
