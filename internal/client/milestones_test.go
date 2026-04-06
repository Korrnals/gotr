// internal/client/milestones_test.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
				mockClient.GetMilestoneFunc = func(ctx context.Context, id int64) (*data.Milestone, error) {
					return tc.mockResponse, nil
				}
			} else {
				mockClient.GetMilestoneFunc = func(ctx context.Context, id int64) (*data.Milestone, error) {
					return nil, fmt.Errorf("milestone %d not found", id)
				}
			}

			ctx := context.Background()
			result, err := mockClient.GetMilestone(ctx, tc.milestoneID)

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
			mockClient.GetMilestonesFunc = func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.GetMilestones(ctx, tc.projectID)

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
			mockClient.AddMilestoneFunc = func(ctx context.Context, projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.AddMilestone(ctx, tc.projectID, tc.request)

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
			mockClient.UpdateMilestoneFunc = func(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.UpdateMilestone(ctx, tc.milestoneID, tc.request)

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
				mockClient.DeleteMilestoneFunc = func(ctx context.Context, milestoneID int64) error {
					return fmt.Errorf("milestone %d not found", milestoneID)
				}
			} else {
				mockClient.DeleteMilestoneFunc = func(ctx context.Context, milestoneID int64) error {
					return nil
				}
			}

			ctx := context.Background()
			err := mockClient.DeleteMilestone(ctx, tc.milestoneID)

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

	ctx := context.Background()
	milestone, err := client.GetMilestone(ctx, 1)
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

	ctx := context.Background()
	milestones, err := client.GetMilestones(ctx, 1)
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

	ctx := context.Background()
	milestone, err := client.AddMilestone(ctx, 1, req)
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

	ctx := context.Background()
	milestone, err := client.UpdateMilestone(ctx, 1, req)
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

	ctx := context.Background()
	err := client.DeleteMilestone(ctx, 1)
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
		DueOn:       data.Timestamp{Time: time.Now()},
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

func TestAddMilestone_DuplicateCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "add_milestone") {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":"Milestone with this name already exists"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.AddMilestone(context.Background(), 1, &data.AddMilestoneRequest{Name: "DupMilestone"})
	if err == nil {
		t.Fatal("AddMilestone with duplicate should error")
	}
}

func TestAddMilestone_ValidationErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "add_milestone") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"Milestone name is required"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.AddMilestone(context.Background(), 1, &data.AddMilestoneRequest{Name: ""})
	if err == nil {
		t.Fatal("AddMilestone with empty name should error")
	}
}

func TestAddMilestone_NilRequest(t *testing.T) {
	c, _ := NewClient("http://unused", "t", "t", false)
	_, err := c.AddMilestone(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("AddMilestone with nil request should error")
	}
}

func TestUpdateMilestone_ConflictDetection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "update_milestone") {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":"Milestone has been modified"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.UpdateMilestone(context.Background(), 1, &data.UpdateMilestoneRequest{Name: "New"})
	if err == nil {
		t.Fatal("UpdateMilestone with conflict should error")
	}
}

func TestUpdateMilestone_PartialUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "update_milestone/5") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Milestone{ID: 5, Name: "PartiallyUpdated", ProjectID: 1})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	m, err := c.UpdateMilestone(context.Background(), 5, &data.UpdateMilestoneRequest{Name: "PartiallyUpdated"})
	if err != nil {
		t.Fatalf("UpdateMilestone with partial data should succeed: %v", err)
	}
	if m.Name != "PartiallyUpdated" {
		t.Errorf("Expected name 'PartiallyUpdated', got '%s'", m.Name)
	}
}

func TestUpdateMilestone_NilRequest(t *testing.T) {
	c, _ := NewClient("http://unused", "t", "t", false)
	_, err := c.UpdateMilestone(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("UpdateMilestone with nil request should error")
	}
}

func TestGetMilestones_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_milestones/1") {
			callCount++
			w.WriteHeader(http.StatusOK)
			// Simulate minimal response
			_ = json.NewEncoder(w).Encode([]data.Milestone{
				{ID: int64(callCount), Name: fmt.Sprintf("Milestone %d", callCount), ProjectID: 1},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	milestones, err := c.GetMilestones(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetMilestones should succeed: %v", err)
	}
	if len(milestones) == 0 {
		t.Fatal("GetMilestones should return milestones")
	}
}

func TestDeleteMilestone_ErrorCase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "delete_milestone") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Internal server error"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	err := c.DeleteMilestone(context.Background(), 999)
	if err == nil {
		t.Fatal("DeleteMilestone with server error should error")
	}
}

func TestGetMilestone_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_milestone") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{invalid}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.GetMilestone(context.Background(), 1)
	if err == nil {
		t.Fatal("GetMilestone with invalid JSON should error")
	}
}

func TestAddAndUpdateMilestone_HTTPDecodeAndRequestBranches(t *testing.T) {
	t.Run("add milestone decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.String(), "add_milestone/7") {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.AddMilestone(context.Background(), 7, &data.AddMilestoneRequest{Name: "M"})
		if err == nil {
			t.Fatal("expected decode error for AddMilestone")
		}
	})

	t.Run("add milestone request error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.AddMilestone(context.Background(), 7, &data.AddMilestoneRequest{Name: "M"})
		if err == nil {
			t.Fatal("expected request error for AddMilestone")
		}
	})

	t.Run("update milestone decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.String(), "update_milestone/9") {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.UpdateMilestone(context.Background(), 9, &data.UpdateMilestoneRequest{Name: "U"})
		if err == nil {
			t.Fatal("expected decode error for UpdateMilestone")
		}
	})

	t.Run("update milestone request error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.UpdateMilestone(context.Background(), 9, &data.UpdateMilestoneRequest{Name: "U"})
		if err == nil {
			t.Fatal("expected request error for UpdateMilestone")
		}
	})
}
