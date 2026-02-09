// internal/client/users_test.go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func TestGetUsers(t *testing.T) {
	testCases := []struct {
		name         string
		mockResponse data.GetUsersResponse
		wantErr      bool
		wantCount    int
	}{
		{
			name: "Success with users",
			mockResponse: []data.User{
				{ID: 1, Name: "John Doe", Email: "john@example.com", IsActive: true},
				{ID: 2, Name: "Jane Smith", Email: "jane@example.com", IsActive: true},
			},
			wantCount: 2,
		},
		{
			name:         "Success empty",
			mockResponse: []data.User{},
			wantCount:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.GetUsersFunc = func() (data.GetUsersResponse, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.GetUsers()

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetUsers() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetUsers() unexpected error: %v", err)
				return
			}
			if len(result) != tc.wantCount {
				t.Errorf("GetUsers() returned %d users, want %d", len(result), tc.wantCount)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	testCases := []struct {
		name         string
		userID       int64
		mockResponse *data.User
		wantErr      bool
	}{
		{
			name:   "Success",
			userID: 1,
			mockResponse: &data.User{
				ID:       1,
				Name:     "John Doe",
				Email:    "john@example.com",
				IsActive: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.GetUserFunc = func(userID int64) (*data.User, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.GetUser(tc.userID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetUser() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetUser() unexpected error: %v", err)
				return
			}
			if result.ID != tc.mockResponse.ID {
				t.Errorf("GetUser() ID = %d, want %d", result.ID, tc.mockResponse.ID)
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	testCases := []struct {
		name         string
		email        string
		mockResponse *data.User
		wantErr      bool
	}{
		{
			name:  "Success",
			email: "john@example.com",
			mockResponse: &data.User{
				ID:       1,
				Name:     "John Doe",
				Email:    "john@example.com",
				IsActive: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.GetUserByEmailFunc = func(email string) (*data.User, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.GetUserByEmail(tc.email)

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetUserByEmail() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetUserByEmail() unexpected error: %v", err)
				return
			}
			if result.Email != tc.mockResponse.Email {
				t.Errorf("GetUserByEmail() Email = %s, want %s", result.Email, tc.mockResponse.Email)
			}
		})
	}
}

func TestGetPriorities(t *testing.T) {
	testCases := []struct {
		name         string
		mockResponse data.GetPrioritiesResponse
		wantErr      bool
		wantCount    int
	}{
		{
			name: "Success",
			mockResponse: []data.Priority{
				{ID: 1, Name: "Critical", ShortName: "C", IsDefault: false, Priority: 1},
				{ID: 2, Name: "High", ShortName: "H", IsDefault: false, Priority: 2},
				{ID: 3, Name: "Medium", ShortName: "M", IsDefault: true, Priority: 3},
				{ID: 4, Name: "Low", ShortName: "L", IsDefault: false, Priority: 4},
			},
			wantCount: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.GetPrioritiesFunc = func() (data.GetPrioritiesResponse, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.GetPriorities()

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetPriorities() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetPriorities() unexpected error: %v", err)
				return
			}
			if len(result) != tc.wantCount {
				t.Errorf("GetPriorities() returned %d priorities, want %d", len(result), tc.wantCount)
			}
		})
	}
}

func TestGetStatuses(t *testing.T) {
	testCases := []struct {
		name         string
		mockResponse data.GetStatusesResponse
		wantErr      bool
		wantCount    int
	}{
		{
			name: "Success",
			mockResponse: []data.Status{
				{ID: 1, Name: "passed", Label: "Passed", IsFinal: true, IsSystem: true, IsUntested: false},
				{ID: 5, Name: "failed", Label: "Failed", IsFinal: true, IsSystem: true, IsUntested: false},
				{ID: 3, Name: "untested", Label: "Untested", IsFinal: false, IsSystem: true, IsUntested: true},
			},
			wantCount: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.GetStatusesFunc = func() (data.GetStatusesResponse, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.GetStatuses()

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetStatuses() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetStatuses() unexpected error: %v", err)
				return
			}
			if len(result) != tc.wantCount {
				t.Errorf("GetStatuses() returned %d statuses, want %d", len(result), tc.wantCount)
			}
		})
	}
}

func TestGetTemplates(t *testing.T) {
	testCases := []struct {
		name         string
		projectID    int64
		mockResponse data.GetTemplatesResponse
		wantErr      bool
		wantCount    int
	}{
		{
			name:      "Success",
			projectID: 1,
			mockResponse: []data.Template{
				{ID: 1, Name: "Test Case (Text)", IsDefault: true},
				{ID: 2, Name: "Test Case (Steps)", IsDefault: false},
			},
			wantCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.GetTemplatesFunc = func(projectID int64) (data.GetTemplatesResponse, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.GetTemplates(tc.projectID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetTemplates() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetTemplates() unexpected error: %v", err)
				return
			}
			if len(result) != tc.wantCount {
				t.Errorf("GetTemplates() returned %d templates, want %d", len(result), tc.wantCount)
			}
		})
	}
}

func TestHTTPGetUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]data.User{
			{ID: 1, Name: "John Doe", Email: "john@example.com"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	users, err := client.GetUsers()
	if err != nil {
		t.Fatalf("GetUsers() error: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
}

func TestHTTPGetPriorities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]data.Priority{
			{ID: 1, Name: "Critical", ShortName: "C"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	priorities, err := client.GetPriorities()
	if err != nil {
		t.Fatalf("GetPriorities() error: %v", err)
	}
	if len(priorities) != 1 {
		t.Errorf("Expected 1 priority, got %d", len(priorities))
	}
}
