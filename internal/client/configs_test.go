// internal/client/configs_test.go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func TestGetConfigs(t *testing.T) {
	testCases := []struct {
		name         string
		projectID    int64
		mockResponse data.GetConfigsResponse
		wantErr      bool
		wantCount    int
	}{
		{
			name:      "Success with configs",
			projectID: 1,
			mockResponse: []data.ConfigGroup{
				{
					ID:        1,
					Name:      "OS",
					ProjectID: 1,
					Configs: []data.Config{
						{ID: 1, Name: "Windows", GroupID: 1},
						{ID: 2, Name: "Linux", GroupID: 1},
					},
				},
			},
			wantCount: 1,
		},
		{
			name:         "Success empty",
			projectID:    2,
			mockResponse: []data.ConfigGroup{},
			wantCount:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.GetConfigsFunc = func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.GetConfigs(ctx, tc.projectID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("GetConfigs() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetConfigs() unexpected error: %v", err)
				return
			}
			if len(result) != tc.wantCount {
				t.Errorf("GetConfigs() returned %d groups, want %d", len(result), tc.wantCount)
			}
		})
	}
}

func TestAddConfigGroup(t *testing.T) {
	testCases := []struct {
		name         string
		projectID    int64
		request      *data.AddConfigGroupRequest
		mockResponse *data.ConfigGroup
		wantErr      bool
	}{
		{
			name:      "Success",
			projectID: 1,
			request: &data.AddConfigGroupRequest{
				Name: "Browser",
			},
			mockResponse: &data.ConfigGroup{
				ID:        2,
				Name:      "Browser",
				ProjectID: 1,
				Configs:   []data.Config{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.AddConfigGroupFunc = func(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.AddConfigGroup(ctx, tc.projectID, tc.request)

			if tc.wantErr {
				if err == nil {
					t.Errorf("AddConfigGroup() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("AddConfigGroup() unexpected error: %v", err)
				return
			}
			if result.Name != tc.mockResponse.Name {
				t.Errorf("AddConfigGroup() Name = %s, want %s", result.Name, tc.mockResponse.Name)
			}
		})
	}
}

func TestAddConfig(t *testing.T) {
	testCases := []struct {
		name         string
		groupID      int64
		request      *data.AddConfigRequest
		mockResponse *data.Config
		wantErr      bool
	}{
		{
			name:    "Success",
			groupID: 1,
			request: &data.AddConfigRequest{
				Name: "Chrome",
			},
			mockResponse: &data.Config{
				ID:      3,
				Name:    "Chrome",
				GroupID: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.AddConfigFunc = func(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.AddConfig(ctx, tc.groupID, tc.request)

			if tc.wantErr {
				if err == nil {
					t.Errorf("AddConfig() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("AddConfig() unexpected error: %v", err)
				return
			}
			if result.Name != tc.mockResponse.Name {
				t.Errorf("AddConfig() Name = %s, want %s", result.Name, tc.mockResponse.Name)
			}
		})
	}
}

func TestUpdateConfigGroup(t *testing.T) {
	testCases := []struct {
		name         string
		groupID      int64
		request      *data.UpdateConfigGroupRequest
		mockResponse *data.ConfigGroup
		wantErr      bool
	}{
		{
			name:    "Success",
			groupID: 1,
			request: &data.UpdateConfigGroupRequest{
				Name: "Operating System",
			},
			mockResponse: &data.ConfigGroup{
				ID:        1,
				Name:      "Operating System",
				ProjectID: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.UpdateConfigGroupFunc = func(ctx context.Context, groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
				return tc.mockResponse, nil
			}

			ctx := context.Background()
			result, err := mockClient.UpdateConfigGroup(ctx, tc.groupID, tc.request)

			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateConfigGroup() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateConfigGroup() unexpected error: %v", err)
				return
			}
			if result.Name != tc.mockResponse.Name {
				t.Errorf("UpdateConfigGroup() Name = %s, want %s", result.Name, tc.mockResponse.Name)
			}
		})
	}
}

func TestDeleteConfigGroup(t *testing.T) {
	testCases := []struct {
		name    string
		groupID int64
		wantErr bool
	}{
		{
			name:    "Success",
			groupID: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			mockClient.DeleteConfigGroupFunc = func(ctx context.Context, groupID int64) error {
				return nil
			}

			ctx := context.Background()
			err := mockClient.DeleteConfigGroup(ctx, tc.groupID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteConfigGroup() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteConfigGroup() unexpected error: %v", err)
			}
		})
	}
}

func TestHTTPGetConfigs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]data.ConfigGroup{
			{
				ID:   1,
				Name: "OS",
				Configs: []data.Config{
					{ID: 1, Name: "Windows"},
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	ctx := context.Background()
	configs, err := client.GetConfigs(ctx, 1)
	if err != nil {
		t.Fatalf("GetConfigs() error: %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("Expected 1 config group, got %d", len(configs))
	}
	if configs[0].Name != "OS" {
		t.Errorf("Expected name 'OS', got '%s'", configs[0].Name)
	}
}

func TestHTTPConfigMutations(t *testing.T) {
	t.Run("add config group", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.String(), "add_config_group/1") {
				t.Fatalf("unexpected endpoint: %s", r.URL.String())
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.ConfigGroup{ID: 10, Name: "Browser", ProjectID: 1})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		group, err := client.AddConfigGroup(context.Background(), 1, &data.AddConfigGroupRequest{Name: "Browser"})
		if err != nil {
			t.Fatalf("AddConfigGroup() error: %v", err)
		}
		if group.ID != 10 {
			t.Fatalf("unexpected group payload: %+v", group)
		}
	})

	t.Run("add config", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.String(), "add_config/10") {
				t.Fatalf("unexpected endpoint: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Config{ID: 20, Name: "Chrome", GroupID: 10})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		cfg, err := client.AddConfig(context.Background(), 10, &data.AddConfigRequest{Name: "Chrome"})
		if err != nil {
			t.Fatalf("AddConfig() error: %v", err)
		}
		if cfg.ID != 20 {
			t.Fatalf("unexpected config payload: %+v", cfg)
		}
	})

	t.Run("update config group", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.String(), "update_config_group/10") {
				t.Fatalf("unexpected endpoint: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.ConfigGroup{ID: 10, Name: "Browsers", ProjectID: 1})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		group, err := client.UpdateConfigGroup(context.Background(), 10, &data.UpdateConfigGroupRequest{Name: "Browsers"})
		if err != nil {
			t.Fatalf("UpdateConfigGroup() error: %v", err)
		}
		if group.Name != "Browsers" {
			t.Fatalf("unexpected group payload: %+v", group)
		}
	})

	t.Run("update config", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.String(), "update_config/20") {
				t.Fatalf("unexpected endpoint: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Config{ID: 20, Name: "Firefox", GroupID: 10})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		cfg, err := client.UpdateConfig(context.Background(), 20, &data.UpdateConfigRequest{Name: "Firefox"})
		if err != nil {
			t.Fatalf("UpdateConfig() error: %v", err)
		}
		if cfg.Name != "Firefox" {
			t.Fatalf("unexpected config payload: %+v", cfg)
		}
	})

	t.Run("delete config group and config", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.String(), "delete_config_group/10") && !strings.Contains(r.URL.String(), "delete_config/20") {
				t.Fatalf("unexpected endpoint: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		if err := client.DeleteConfigGroup(context.Background(), 10); err != nil {
			t.Fatalf("DeleteConfigGroup() error: %v", err)
		}
		if err := client.DeleteConfig(context.Background(), 20); err != nil {
			t.Fatalf("DeleteConfig() error: %v", err)
		}
	})
}

func TestHTTPConfigs_ErrorBranches(t *testing.T) {
	t.Run("get configs non-OK", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"server exploded"}`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		_, err := client.GetConfigs(context.Background(), 1)
		if err == nil {
			t.Fatalf("expected GetConfigs() error for non-OK status")
		}
		if !strings.Contains(err.Error(), "error getting configs for project") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("get configs decode error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		_, err := client.GetConfigs(context.Background(), 1)
		if err == nil {
			t.Fatalf("expected GetConfigs() decode error")
		}
		if !strings.Contains(err.Error(), "error decoding configs") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("get configs request error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		client, _ := NewClient(server.URL, "test", "test", false)
		server.Close()

		_, err := client.GetConfigs(context.Background(), 1)
		if err == nil {
			t.Fatalf("expected GetConfigs() request error")
		}
	})

	t.Run("post-based methods non-OK", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"bad request"}`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)

		if _, err := client.AddConfigGroup(context.Background(), 0, &data.AddConfigGroupRequest{Name: ""}); err == nil {
			t.Fatalf("expected AddConfigGroup() error")
		}
		if _, err := client.AddConfig(context.Background(), 0, &data.AddConfigRequest{Name: ""}); err == nil {
			t.Fatalf("expected AddConfig() error")
		}
		if _, err := client.UpdateConfigGroup(context.Background(), 0, &data.UpdateConfigGroupRequest{Name: ""}); err == nil {
			t.Fatalf("expected UpdateConfigGroup() error")
		}
		if _, err := client.UpdateConfig(context.Background(), 0, &data.UpdateConfigRequest{Name: ""}); err == nil {
			t.Fatalf("expected UpdateConfig() error")
		}
		if err := client.DeleteConfigGroup(context.Background(), 0); err == nil {
			t.Fatalf("expected DeleteConfigGroup() error")
		}
		if err := client.DeleteConfig(context.Background(), 0); err == nil {
			t.Fatalf("expected DeleteConfig() error")
		}
	})

	t.Run("post-based methods decode errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.String(), "delete_config_group/") || strings.Contains(r.URL.String(), "delete_config/") {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)

		if _, err := client.AddConfigGroup(context.Background(), 1, &data.AddConfigGroupRequest{Name: "A"}); err == nil {
			t.Fatalf("expected AddConfigGroup() decode error")
		}
		if _, err := client.AddConfig(context.Background(), 1, &data.AddConfigRequest{Name: "A"}); err == nil {
			t.Fatalf("expected AddConfig() decode error")
		}
		if _, err := client.UpdateConfigGroup(context.Background(), 1, &data.UpdateConfigGroupRequest{Name: "A"}); err == nil {
			t.Fatalf("expected UpdateConfigGroup() decode error")
		}
		if _, err := client.UpdateConfig(context.Background(), 1, &data.UpdateConfigRequest{Name: "A"}); err == nil {
			t.Fatalf("expected UpdateConfig() decode error")
		}
	})

	t.Run("post-based methods request errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		client, _ := NewClient(server.URL, "test", "test", false)
		server.Close()

		if _, err := client.AddConfigGroup(context.Background(), 1, &data.AddConfigGroupRequest{Name: "A"}); err == nil {
			t.Fatalf("expected AddConfigGroup() request error")
		}
		if _, err := client.AddConfig(context.Background(), 1, &data.AddConfigRequest{Name: "A"}); err == nil {
			t.Fatalf("expected AddConfig() request error")
		}
		if _, err := client.UpdateConfigGroup(context.Background(), 1, &data.UpdateConfigGroupRequest{Name: "A"}); err == nil {
			t.Fatalf("expected UpdateConfigGroup() request error")
		}
		if _, err := client.UpdateConfig(context.Background(), 1, &data.UpdateConfigRequest{Name: "A"}); err == nil {
			t.Fatalf("expected UpdateConfig() request error")
		}
		if err := client.DeleteConfigGroup(context.Background(), 1); err == nil {
			t.Fatalf("expected DeleteConfigGroup() request error")
		}
		if err := client.DeleteConfig(context.Background(), 1); err == nil {
			t.Fatalf("expected DeleteConfig() request error")
		}
	})
}

func TestGetConfigs_ValidationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"Invalid project ID"}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.GetConfigs(context.Background(), -1)
	if err == nil {
		t.Fatal("GetConfigs with invalid projectID should error")
	}
}

func TestGetConfigs_PermissionDenied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"Permission denied"}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.GetConfigs(context.Background(), 1)
	if err == nil {
		t.Fatal("GetConfigs with 403 should error")
	}
}

func TestGetConfigs_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.GetConfigs(context.Background(), 1)
	if err == nil {
		t.Fatal("GetConfigs with invalid JSON should error")
	}
}

func TestAddConfigGroup_ValidationErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "add_config_group") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"Config group name is required"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.AddConfigGroup(context.Background(), 1, &data.AddConfigGroupRequest{Name: ""})
	if err == nil {
		t.Fatal("AddConfigGroup with empty name should error")
	}
}

func TestUpdateConfigGroup_ConflictDetection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "update_config_group") {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":"Config group already exists"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.UpdateConfigGroup(context.Background(), 1, &data.UpdateConfigGroupRequest{Name: "Dup"})
	if err == nil {
		t.Fatal("UpdateConfigGroup with conflict should error")
	}
}

func TestUpdateConfig_ConflictDetection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "update_config/") {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":"Config conflict"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.UpdateConfig(context.Background(), 999, &data.UpdateConfigRequest{Name: "X"})
	if err == nil {
		t.Fatal("UpdateConfig with conflict should error")
	}
}

func TestDeleteConfigGroup_PermissionDenied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "delete_config_group") {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"Insufficient permissions"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	err := c.DeleteConfigGroup(context.Background(), 1)
	if err == nil {
		t.Fatal("DeleteConfigGroup with 403 should error")
	}
}

func TestDeleteConfig_PermissionDenied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "delete_config/") {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"Insufficient permissions"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	err := c.DeleteConfig(context.Background(), 999)
	if err == nil {
		t.Fatal("DeleteConfig with 403 should error")
	}
}
