// internal/client/milestones_test.go
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
)

func TestGetMilestone(t *testing.T) {
	testCases := []struct {
		name           string
		milestoneID    int64
		mockStatusCode int
		mockResponse   *data.Milestone
		wantErr        bool
	}{
		{
			name:           "Success",
			milestoneID:    1,
			mockStatusCode: http.StatusOK,
			mockResponse: &data.Milestone{
				ID:          1,
				Name:        "Release 1.0",
				Description: "First release",
				ProjectID:   1,
				URL:         "https://example.com/index.php?/milestones/view/1",
			},
		},
		{
			name:           "MilestoneNotFound",
			milestoneID:    999,
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			if tc.mockResponse != nil {
				mockClient.GetMilestoneFunc = func(id int64) (*data.Milestone, error) {
					return tc.mockResponse, nil
				}
			} else {
				mockClient.GetMilestoneFunc = func(id int64) (*data.Milestone, error) {
					return nil, fmt.Errorf("milestone %d not found", id)
				}
			}

			result, err := mockClient.GetMilestone(tc.milestoneID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetMilestone() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetMilestone() unexpected error: %v", err)
				return
			}
			if result.ID != tc.mockResponse.ID {
				t.Errorf("GetMilestone() ID = %d, want %d", result.ID, tc.mockResponse.ID)
			}
		})
	}
}

func TestGetMilestones(t *testing.T) {
	testCases := []struct {
		name         string
		projectID    int64
		mockResponse data.GetMilestonesResponse
		wantErr      bool
		wantCount    int
	}{
		{
			name:      "Success with milestones",
			projectID: 1,
			mockResponse: []data.Milestone{
				{ID: 1, Name: "Release 1.0", ProjectID: 1},
				{ID: 2, Name: "Release 2.0", ProjectID: 1},
			},
			wantCount: 2,
		},
		{
			name:         "Success empty",
			projectID:    2,
			mockResponse: []data.Milestone{},
			wantCount:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.GetMilestonesFunc = func(projectID int64) ([]data.Milestone, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.GetMilestones(tc.projectID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetMilestones() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetMilestones() unexpected error: %v", err)
				return
			}
			if len(result) != tc.wantCount {
				t.Errorf("GetMilestones() returned %d milestones, want %d", len(result), tc.wantCount)
			}
		})
	}
}

func TestAddMilestone(t *testing.T) {
	testCases := []struct {
		name         string
		projectID    int64
		request      *data.AddMilestoneRequest
		mockResponse *data.Milestone
		wantErr      bool
	}{
		{
			name:      "Success",
			projectID: 1,
			request: &data.AddMilestoneRequest{
				Name:        "Release 3.0",
				Description: "Third release",
			},
			mockResponse: &data.Milestone{
				ID:          3,
				Name:        "Release 3.0",
				Description: "Third release",
				ProjectID:   1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.AddMilestoneFunc = func(projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.AddMilestone(tc.projectID, tc.request)

			if tc.wantErr {
				if err == nil {
					t.Errorf("AddMilestone() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("AddMilestone() unexpected error: %v", err)
				return
			}
			if result.ID != tc.mockResponse.ID {
				t.Errorf("AddMilestone() ID = %d, want %d", result.ID, tc.mockResponse.ID)
			}
		})
	}
}

func TestUpdateMilestone(t *testing.T) {
	testCases := []struct {
		name         string
		milestoneID  int64
		request      *data.UpdateMilestoneRequest
		mockResponse *data.Milestone
		wantErr      bool
	}{
		{
			name:        "Success",
			milestoneID: 1,
			request: &data.UpdateMilestoneRequest{
				Name:        "Release 1.1",
				IsCompleted: true,
			},
			mockResponse: &data.Milestone{
				ID:          1,
				Name:        "Release 1.1",
				IsCompleted: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.UpdateMilestoneFunc = func(milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.UpdateMilestone(tc.milestoneID, tc.request)

			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateMilestone() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateMilestone() unexpected error: %v", err)
				return
			}
			if result.Name != tc.mockResponse.Name {
				t.Errorf("UpdateMilestone() Name = %s, want %s", result.Name, tc.mockResponse.Name)
			}
		})
	}
}

func TestDeleteMilestone(t *testing.T) {
	testCases := []struct {
		name        string
		milestoneID int64
		wantErr     bool
	}{
		{
			name:        "Success",
			milestoneID: 1,
		},
		{
			name:        "NotFound",
			milestoneID: 999,
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			if tc.wantErr {
				mockClient.DeleteMilestoneFunc = func(milestoneID int64) error {
					return fmt.Errorf("milestone %d not found", milestoneID)
				}
			} else {
				mockClient.DeleteMilestoneFunc = func(milestoneID int64) error {
					return nil
				}
			}

			err := mockClient.DeleteMilestone(tc.milestoneID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteMilestone() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteMilestone() unexpected error: %v", err)
			}
		})
	}
}

// HTTP integration tests for milestone endpoints
func TestHTTPGetMilestone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.Milestone{
			ID:   1,
			Name: "Release 1.0",
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	milestone, err := client.GetMilestone(1)
	if err != nil {
		t.Fatalf("GetMilestone() error: %v", err)
	}
	if milestone.Name != "Release 1.0" {
		t.Errorf("Expected name 'Release 1.0', got '%s'", milestone.Name)
	}
}

func TestHTTPGetMilestones(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]data.Milestone{
			{ID: 1, Name: "Milestone 1"},
			{ID: 2, Name: "Milestone 2"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	milestones, err := client.GetMilestones(1)
	if err != nil {
		t.Fatalf("GetMilestones() error: %v", err)
	}
	if len(milestones) != 2 {
		t.Errorf("Expected 2 milestones, got %d", len(milestones))
	}
}

func TestHTTPAddMilestone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.Milestone{
			ID:   3,
			Name: "New Milestone",
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	req := &data.AddMilestoneRequest{
		Name: "New Milestone",
	}

	milestone, err := client.AddMilestone(1, req)
	if err != nil {
		t.Fatalf("AddMilestone() error: %v", err)
	}
	if milestone.Name != "New Milestone" {
		t.Errorf("Expected name 'New Milestone', got '%s'", milestone.Name)
	}
}

func TestHTTPUpdateMilestone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.Milestone{
			ID:   1,
			Name: "Updated Milestone",
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	req := &data.UpdateMilestoneRequest{
		Name: "Updated Milestone",
	}

	milestone, err := client.UpdateMilestone(1, req)
	if err != nil {
		t.Fatalf("UpdateMilestone() error: %v", err)
	}
	if milestone.Name != "Updated Milestone" {
		t.Errorf("Expected name 'Updated Milestone', got '%s'", milestone.Name)
	}
}

func TestHTTPDeleteMilestone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	err := client.DeleteMilestone(1)
	if err != nil {
		t.Fatalf("DeleteMilestone() error: %v", err)
	}
}

func TestMilestoneModel(t *testing.T) {
	milestone := data.Milestone{
		ID:          1,
		Name:        "Test Milestone",
		Description: "Test description",
		ProjectID:   1,
		DueOn:       time.Now(),
		IsCompleted: false,
		URL:         "https://example.com/milestone/1",
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(milestone)
	if err != nil {
		t.Fatalf("Failed to marshal milestone: %v", err)
	}

	// Test JSON deserialization
	var decoded data.Milestone
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal milestone: %v", err)
	}

	if decoded.Name != milestone.Name {
		t.Errorf("Name mismatch: expected %s, got %s", milestone.Name, decoded.Name)
	}
}
