// internal/client/extended_test.go
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

func TestGetGroups(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.GetGroupsFunc = func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
		return []data.Group{
			{ID: 1, Name: "QA Team", ProjectID: 1, UserIDs: []int64{1, 2}},
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.GetGroups(ctx, 1)
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
	mockClient.GetRolesFunc = func(ctx context.Context) (data.GetRolesResponse, error) {
		return []data.Role{
			{ID: 1, Name: "Lead"},
			{ID: 2, Name: "Tester"},
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.GetRoles(ctx)
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
	mockClient.GetResultFieldsFunc = func(ctx context.Context) (data.GetResultFieldsResponse, error) {
		return []data.ResultField{
			{ID: 1, Name: "Status", SystemName: "status_id", IsActive: true},
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.GetResultFields(ctx)
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
	mockClient.GetDatasetsFunc = func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
		return []data.Dataset{
			{ID: 1, Name: "Browser Data", ProjectID: 1},
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.GetDatasets(ctx, 1)
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
	mockClient.GetVariablesFunc = func(ctx context.Context, datasetID int64) (data.GetVariablesResponse, error) {
		return []data.Variable{
			{ID: 1, Name: "browser", DatasetID: 1},
			{ID: 2, Name: "version", DatasetID: 1},
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.GetVariables(ctx, 1)
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
	mockClient.GetBDDFunc = func(ctx context.Context, caseID int64) (*data.BDD, error) {
		return &data.BDD{
			ID:      1,
			CaseID:  caseID,
			Content: "Given user is logged in",
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.GetBDD(ctx, 1)
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

	ctx := context.Background()
	groups, err := client.GetGroups(ctx, 1)
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

	ctx := context.Background()
	roles, err := client.GetRoles(ctx)
	if err != nil {
		t.Fatalf("GetRoles() error: %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(roles))
	}
}

func TestHTTPExtendedGroupsAndRoles(t *testing.T) {
	t.Run("group methods", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_group/2"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Group{ID: 2, Name: "Ops"})
			case strings.Contains(r.URL.String(), "add_group/1"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Group{ID: 3, Name: "New Group"})
			case strings.Contains(r.URL.String(), "update_group/3"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Group{ID: 3, Name: "Updated Group"})
			case strings.Contains(r.URL.String(), "delete_group/3"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			case strings.Contains(r.URL.String(), "get_role/7"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Role{ID: 7, Name: "Manager"})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "test", "test", false)
		ctx := context.Background()

		g, err := c.GetGroup(ctx, 2)
		if err != nil || g.ID != 2 {
			t.Fatalf("GetGroup() failed: %v, %+v", err, g)
		}

		g, err = c.AddGroup(ctx, 1, "New Group", []int64{1, 2})
		if err != nil || g.ID != 3 {
			t.Fatalf("AddGroup() failed: %v, %+v", err, g)
		}

		g, err = c.UpdateGroup(ctx, 3, "Updated Group", []int64{1})
		if err != nil || g.Name != "Updated Group" {
			t.Fatalf("UpdateGroup() failed: %v, %+v", err, g)
		}

		if err := c.DeleteGroup(ctx, 3); err != nil {
			t.Fatalf("DeleteGroup() failed: %v", err)
		}

		role, err := c.GetRole(ctx, 7)
		if err != nil || role.ID != 7 {
			t.Fatalf("GetRole() failed: %v, %+v", err, role)
		}
	})

	t.Run("result fields", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.String(), "get_result_fields") {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]data.ResultField{{ID: 1, Name: "Status", SystemName: "status_id", IsActive: true}})
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "test", "test", false)
		fields, err := c.GetResultFields(context.Background())
		if err != nil || len(fields) != 1 {
			t.Fatalf("GetResultFields() failed: %v, %+v", err, fields)
		}
	})
}

func TestHTTPExtendedDatasetsVariablesBDDLabels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.String(), "get_datasets/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Dataset{{ID: 1, Name: "DS", ProjectID: 1}})
		case strings.Contains(r.URL.String(), "get_dataset/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Dataset{ID: 1, Name: "DS", ProjectID: 1})
		case strings.Contains(r.URL.String(), "add_dataset/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Dataset{ID: 2, Name: "New DS", ProjectID: 1})
		case strings.Contains(r.URL.String(), "update_dataset/2"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Dataset{ID: 2, Name: "Upd DS", ProjectID: 1})
		case strings.Contains(r.URL.String(), "delete_dataset/2"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.String(), "get_variables/2"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Variable{{ID: 1, Name: "v", DatasetID: 2}})
		case strings.Contains(r.URL.String(), "add_variable/2"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Variable{ID: 3, Name: "v2", DatasetID: 2})
		case strings.Contains(r.URL.String(), "update_variable/3"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Variable{ID: 3, Name: "v3", DatasetID: 2})
		case strings.Contains(r.URL.String(), "delete_variable/3"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.String(), "get_bdd/10"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.BDD{ID: 1, CaseID: 10, Content: "Given"})
		case strings.Contains(r.URL.String(), "add_bdd/10"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.BDD{ID: 2, CaseID: 10, Content: "When"})
		case strings.Contains(r.URL.String(), "update_test_labels/21"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.String(), "update_tests_labels/22"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.String(), "get_labels/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Label{{ID: 1, Name: "smoke"}})
		case strings.Contains(r.URL.String(), "get_label/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Label{ID: 1, Name: "smoke"})
		case strings.Contains(r.URL.String(), "update_label/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Label{ID: 1, Name: "regression"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test", "test", false)
	ctx := context.Background()

	dss, err := c.GetDatasets(ctx, 1)
	if err != nil || len(dss) != 1 {
		t.Fatalf("GetDatasets() failed: %v, %+v", err, dss)
	}

	ds, err := c.GetDataset(ctx, 1)
	if err != nil || ds.ID != 1 {
		t.Fatalf("GetDataset() failed: %v, %+v", err, ds)
	}

	ds, err = c.AddDataset(ctx, 1, "New DS")
	if err != nil || ds.ID != 2 {
		t.Fatalf("AddDataset() failed: %v, %+v", err, ds)
	}

	ds, err = c.UpdateDataset(ctx, 2, "Upd DS")
	if err != nil || ds.Name != "Upd DS" {
		t.Fatalf("UpdateDataset() failed: %v, %+v", err, ds)
	}

	if err := c.DeleteDataset(ctx, 2); err != nil {
		t.Fatalf("DeleteDataset() failed: %v", err)
	}

	vars, err := c.GetVariables(ctx, 2)
	if err != nil || len(vars) != 1 {
		t.Fatalf("GetVariables() failed: %v, %+v", err, vars)
	}

	v, err := c.AddVariable(ctx, 2, "v2")
	if err != nil || v.ID != 3 {
		t.Fatalf("AddVariable() failed: %v, %+v", err, v)
	}

	v, err = c.UpdateVariable(ctx, 3, "v3")
	if err != nil || v.Name != "v3" {
		t.Fatalf("UpdateVariable() failed: %v, %+v", err, v)
	}

	if err := c.DeleteVariable(ctx, 3); err != nil {
		t.Fatalf("DeleteVariable() failed: %v", err)
	}

	bdd, err := c.GetBDD(ctx, 10)
	if err != nil || bdd.CaseID != 10 {
		t.Fatalf("GetBDD() failed: %v, %+v", err, bdd)
	}

	bdd, err = c.AddBDD(ctx, 10, "When")
	if err != nil || bdd.ID != 2 {
		t.Fatalf("AddBDD() failed: %v, %+v", err, bdd)
	}

	if err := c.UpdateTestLabels(ctx, 21, []string{"smoke"}); err != nil {
		t.Fatalf("UpdateTestLabels() failed: %v", err)
	}

	if err := c.UpdateTestsLabels(ctx, 22, []int64{1, 2}, []string{"regression"}); err != nil {
		t.Fatalf("UpdateTestsLabels() failed: %v", err)
	}

	labels, err := c.GetLabels(ctx, 1)
	if err != nil || len(labels) != 1 {
		t.Fatalf("GetLabels() failed: %v, %+v", err, labels)
	}

	label, err := c.GetLabel(ctx, 1)
	if err != nil || label.ID != 1 {
		t.Fatalf("GetLabel() failed: %v, %+v", err, label)
	}

	label, err = c.UpdateLabel(ctx, 1, data.UpdateLabelRequest{ProjectID: 1, Title: "regression"})
	if err != nil || label.Name != "regression" {
		t.Fatalf("UpdateLabel() failed: %v, %+v", err, label)
	}
}
