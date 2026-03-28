// internal/client/users_test.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
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
			mockClient.GetUsersFunc = func(ctx context.Context) (data.GetUsersResponse, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.GetUsers(ctx)

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
			mockClient.GetUserFunc = func(ctx context.Context, userID int64) (*data.User, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.GetUser(ctx, tc.userID)

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
			mockClient.GetUserByEmailFunc = func(ctx context.Context, email string) (*data.User, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.GetUserByEmail(ctx, tc.email)

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
			mockClient.GetPrioritiesFunc = func(ctx context.Context) (data.GetPrioritiesResponse, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.GetPriorities(ctx)

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
			mockClient.GetStatusesFunc = func(ctx context.Context) (data.GetStatusesResponse, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.GetStatuses(ctx)

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
			mockClient.GetTemplatesFunc = func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.GetTemplates(ctx, tc.projectID)

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

	ctx := context.Background()
	users, err := client.GetUsers(ctx)
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

	ctx := context.Background()
	priorities, err := client.GetPriorities(ctx)
	if err != nil {
		t.Fatalf("GetPriorities() error: %v", err)
	}
	if len(priorities) != 1 {
		t.Errorf("Expected 1 priority, got %d", len(priorities))
	}
}

func TestHTTPUsersProjectAndLookupEndpoints(t *testing.T) {
	t.Run("GetUsersByProject success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_users/77", r.URL.String())
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetUsersResponse{{ID: 1, Name: "Ann"}})
		})
		defer server.Close()

		users, err := client.GetUsersByProject(context.Background(), 77)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
	})

	t.Run("GetUser success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_user/11", r.URL.String())
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.User{ID: 11, Name: "John"})
		})
		defer server.Close()

		user, err := client.GetUser(context.Background(), 11)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(11), user.ID)
	})

	t.Run("GetUserByEmail success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php", r.URL.Path)
			assert.Contains(t, r.URL.String(), "get_user_by_email")
			assert.Equal(t, "john@example.com", r.URL.Query().Get("email"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.User{ID: 12, Email: "john@example.com"})
		})
		defer server.Close()

		user, err := client.GetUserByEmail(context.Background(), "john@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(12), user.ID)
	})
}

func TestHTTPUsersWriteAndMetaEndpoints(t *testing.T) {
	t.Run("AddUser success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/add_user", r.URL.String())

			var req data.AddUserRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "Jane", req.Name)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.User{ID: 21, Name: req.Name, Email: req.Email})
		})
		defer server.Close()

		user, err := client.AddUser(context.Background(), data.AddUserRequest{Name: "Jane", Email: "jane@example.com"})
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(21), user.ID)
	})

	t.Run("UpdateUser success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/update_user/21", r.URL.String())

			var req data.UpdateUserRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "Jane Updated", req.Name)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.User{ID: 21, Name: req.Name})
		})
		defer server.Close()

		user, err := client.UpdateUser(context.Background(), 21, data.UpdateUserRequest{Name: "Jane Updated"})
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Jane Updated", user.Name)
	})

	t.Run("GetStatuses success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php?/api/v2/get_statuses", r.URL.String())
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetStatusesResponse{{ID: 1, Name: "passed"}})
		})
		defer server.Close()

		statuses, err := client.GetStatuses(context.Background())
		assert.NoError(t, err)
		assert.Len(t, statuses, 1)
	})

	t.Run("GetTemplates success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expected := fmt.Sprintf("/index.php?/api/v2/get_templates/%d", int64(30))
			assert.Equal(t, expected, r.URL.String())
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetTemplatesResponse{{ID: 1, Name: "Template"}})
		})
		defer server.Close()

		templates, err := client.GetTemplates(context.Background(), 30)
		assert.NoError(t, err)
		assert.Len(t, templates, 1)
	})
}
