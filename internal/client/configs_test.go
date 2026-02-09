// internal/client/configs_test.go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
			mockClient.GetConfigsFunc = func(projectID int64) (data.GetConfigsResponse, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.GetConfigs(tc.projectID)

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
			mockClient.AddConfigGroupFunc = func(projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.AddConfigGroup(tc.projectID, tc.request)

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
			mockClient.AddConfigFunc = func(groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.AddConfig(tc.groupID, tc.request)

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
			mockClient.UpdateConfigGroupFunc = func(groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
				return tc.mockResponse, nil
			}

			result, err := mockClient.UpdateConfigGroup(tc.groupID, tc.request)

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
			mockClient.DeleteConfigGroupFunc = func(groupID int64) error {
				return nil
			}

			err := mockClient.DeleteConfigGroup(tc.groupID)

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

	configs, err := client.GetConfigs(1)
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
