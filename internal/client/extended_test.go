// internal/client/extended_test.go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func TestGetGroups(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.GetGroupsFunc = func(projectID int64) (data.GetGroupsResponse, error) {
		return []data.Group{
			{ID: 1, Name: "QA Team", ProjectID: 1, UserIDs: []int64{1, 2}},
		}, nil
	}

	result, err := mockClient.GetGroups(1)
	if err != nil {
		t.Errorf("GetGroups() unexpected error: %v", err)
		return
	}
	if len(result) != 1 {
		t.Errorf("GetGroups() returned %d groups, want 1", len(result))
	}
}

func TestGetRoles(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.GetRolesFunc = func() (data.GetRolesResponse, error) {
		return []data.Role{
			{ID: 1, Name: "Lead"},
			{ID: 2, Name: "Tester"},
		}, nil
	}

	result, err := mockClient.GetRoles()
	if err != nil {
		t.Errorf("GetRoles() unexpected error: %v", err)
		return
	}
	if len(result) != 2 {
		t.Errorf("GetRoles() returned %d roles, want 2", len(result))
	}
}

func TestGetResultFields(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.GetResultFieldsFunc = func() (data.GetResultFieldsResponse, error) {
		return []data.ResultField{
			{ID: 1, Name: "Status", SystemName: "status_id", IsActive: true},
		}, nil
	}

	result, err := mockClient.GetResultFields()
	if err != nil {
		t.Errorf("GetResultFields() unexpected error: %v", err)
		return
	}
	if len(result) != 1 {
		t.Errorf("GetResultFields() returned %d fields, want 1", len(result))
	}
}

func TestGetDatasets(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.GetDatasetsFunc = func(projectID int64) (data.GetDatasetsResponse, error) {
		return []data.Dataset{
			{ID: 1, Name: "Browser Data", ProjectID: 1},
		}, nil
	}

	result, err := mockClient.GetDatasets(1)
	if err != nil {
		t.Errorf("GetDatasets() unexpected error: %v", err)
		return
	}
	if len(result) != 1 {
		t.Errorf("GetDatasets() returned %d datasets, want 1", len(result))
	}
}

func TestGetVariables(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.GetVariablesFunc = func(datasetID int64) (data.GetVariablesResponse, error) {
		return []data.Variable{
			{ID: 1, Name: "browser", DatasetID: 1},
			{ID: 2, Name: "version", DatasetID: 1},
		}, nil
	}

	result, err := mockClient.GetVariables(1)
	if err != nil {
		t.Errorf("GetVariables() unexpected error: %v", err)
		return
	}
	if len(result) != 2 {
		t.Errorf("GetVariables() returned %d variables, want 2", len(result))
	}
}

func TestGetBDD(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.GetBDDFunc = func(caseID int64) (*data.BDD, error) {
		return &data.BDD{
			ID:      1,
			CaseID:  caseID,
			Content: "Given user is logged in",
		}, nil
	}

	result, err := mockClient.GetBDD(1)
	if err != nil {
		t.Errorf("GetBDD() unexpected error: %v", err)
		return
	}
	if result.CaseID != 1 {
		t.Errorf("GetBDD() CaseID = %d, want 1", result.CaseID)
	}
}

func TestHTTPGetGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]data.Group{
			{ID: 1, Name: "QA Team"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	groups, err := client.GetGroups(1)
	if err != nil {
		t.Fatalf("GetGroups() error: %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}
}

func TestHTTPGetRoles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]data.Role{
			{ID: 1, Name: "Lead"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	roles, err := client.GetRoles()
	if err != nil {
		t.Fatalf("GetRoles() error: %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(roles))
	}
}
