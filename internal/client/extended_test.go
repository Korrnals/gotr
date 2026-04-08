// internal/client/extended_test.go
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

func TestHTTPExtended_ErrorBranches(t *testing.T) {
	t.Run("groups non-OK and decode", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_groups/1"):
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("boom"))
			case strings.Contains(r.URL.String(), "get_groups/2"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad request"))
			case strings.Contains(r.URL.String(), "get_group/2"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"broken":`))
			case strings.Contains(r.URL.String(), "get_group/3"):
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden"))
			case strings.Contains(r.URL.String(), "add_group/1"):
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte("validation error"))
			case strings.Contains(r.URL.String(), "update_group/1"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{invalid json`))
			case strings.Contains(r.URL.String(), "delete_group/2"):
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not found"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})
		defer s.Close()

		_, err := c.GetGroups(context.Background(), 1)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected non-OK error, got: %v", err)
		}

		_, err = c.GetGroups(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected groups 400 error, got: %v", err)
		}

		_, err = c.GetGroup(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "error decoding group") {
			t.Fatalf("expected decode error, got: %v", err)
		}

		_, err = c.GetGroup(context.Background(), 3)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected group 403 error, got: %v", err)
		}

		_, err = c.AddGroup(context.Background(), 1, "test", []int64{1})
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected add group error, got: %v", err)
		}

		_, err = c.UpdateGroup(context.Background(), 1, "test", []int64{1})
		if err == nil || !strings.Contains(err.Error(), "error decoding group") {
			t.Fatalf("expected update group decode error, got: %v", err)
		}

		err = c.DeleteGroup(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected delete group error, got: %v", err)
		}
	})

	t.Run("delete group non-OK", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad"))
		})
		defer s.Close()

		err := c.DeleteGroup(context.Background(), 1)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected non-OK error, got: %v", err)
		}
	})

	t.Run("roles and result fields decode/non-OK", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_roles"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"broken":`))
			case strings.Contains(r.URL.String(), "get_role/1"):
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden"))
			case strings.Contains(r.URL.String(), "get_role/2"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`invalid`))
			case strings.Contains(r.URL.String(), "get_result_fields"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})
		defer s.Close()

		_, err := c.GetRoles(context.Background())
		if err == nil || !strings.Contains(err.Error(), "error decoding roles") {
			t.Fatalf("expected roles decode error, got: %v", err)
		}

		_, err = c.GetRole(context.Background(), 1)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected role 403 error, got: %v", err)
		}

		_, err = c.GetRole(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "error decoding role") {
			t.Fatalf("expected role decode error, got: %v", err)
		}

		_, err = c.GetResultFields(context.Background())
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected result fields non-OK error, got: %v", err)
		}
	})

	t.Run("datasets and variables decode/non-OK", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_datasets/1"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"broken":`))
			case strings.Contains(r.URL.String(), "get_datasets/2"):
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden"))
			case strings.Contains(r.URL.String(), "get_dataset/1"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			case strings.Contains(r.URL.String(), "get_dataset/2"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`invalid`))
			case strings.Contains(r.URL.String(), "add_dataset/1"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"broken":`))
			case strings.Contains(r.URL.String(), "add_dataset/2"):
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not found"))
			case strings.Contains(r.URL.String(), "update_dataset/1"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			case strings.Contains(r.URL.String(), "update_dataset/2"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{broken`))
			case strings.Contains(r.URL.String(), "delete_dataset/1"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			case strings.Contains(r.URL.String(), "delete_dataset/2"):
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("unauthorized"))
			case strings.Contains(r.URL.String(), "get_variables/1"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"broken":`))
			case strings.Contains(r.URL.String(), "get_variables/2"):
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("unavailable"))
			case strings.Contains(r.URL.String(), "add_variable/1"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			case strings.Contains(r.URL.String(), "add_variable/2"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`invalid`))
			case strings.Contains(r.URL.String(), "update_variable/1"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"broken":`))
			case strings.Contains(r.URL.String(), "update_variable/2"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			case strings.Contains(r.URL.String(), "delete_variable/1"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			case strings.Contains(r.URL.String(), "delete_variable/2"):
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})
		defer s.Close()

		_, err := c.GetDatasets(context.Background(), 1)
		if err == nil || !strings.Contains(err.Error(), "error decoding datasets") {
			t.Fatalf("expected datasets decode error, got: %v", err)
		}

		_, err = c.GetDatasets(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected datasets 403 error, got: %v", err)
		}

		_, err = c.GetDataset(context.Background(), 1)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected dataset non-OK error, got: %v", err)
		}

		_, err = c.GetDataset(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "error decoding dataset") {
			t.Fatalf("expected dataset decode error, got: %v", err)
		}

		_, err = c.AddDataset(context.Background(), 1, "n")
		if err == nil || !strings.Contains(err.Error(), "error decoding dataset") {
			t.Fatalf("expected add dataset decode error, got: %v", err)
		}

		_, err = c.AddDataset(context.Background(), 2, "n")
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected add dataset 404 error, got: %v", err)
		}

		_, err = c.UpdateDataset(context.Background(), 1, "n")
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected update dataset non-OK error, got: %v", err)
		}

		_, err = c.UpdateDataset(context.Background(), 2, "n")
		if err == nil || !strings.Contains(err.Error(), "error decoding dataset") {
			t.Fatalf("expected update dataset decode error, got: %v", err)
		}

		err = c.DeleteDataset(context.Background(), 1)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected delete dataset non-OK error, got: %v", err)
		}

		err = c.DeleteDataset(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected delete dataset 401 error, got: %v", err)
		}

		_, err = c.GetVariables(context.Background(), 1)
		if err == nil || !strings.Contains(err.Error(), "error decoding variables") {
			t.Fatalf("expected variables decode error, got: %v", err)
		}

		_, err = c.GetVariables(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected variables 503 error, got: %v", err)
		}

		_, err = c.AddVariable(context.Background(), 1, "v")
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected add variable non-OK error, got: %v", err)
		}

		_, err = c.AddVariable(context.Background(), 2, "v")
		if err == nil || !strings.Contains(err.Error(), "error decoding variable") {
			t.Fatalf("expected add variable decode error, got: %v", err)
		}

		_, err = c.UpdateVariable(context.Background(), 1, "v")
		if err == nil || !strings.Contains(err.Error(), "error decoding variable") {
			t.Fatalf("expected update variable decode error, got: %v", err)
		}

		_, err = c.UpdateVariable(context.Background(), 2, "v")
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected update variable non-OK error, got: %v", err)
		}

		err = c.DeleteVariable(context.Background(), 1)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected delete variable non-OK error, got: %v", err)
		}

		err = c.DeleteVariable(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected delete variable 403 error, got: %v", err)
		}
	})

	t.Run("bdd and labels decode/non-OK", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_bdd/10"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"broken":`))
			case strings.Contains(r.URL.String(), "get_bdd/11"):
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden"))
			case strings.Contains(r.URL.String(), "add_bdd/10"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			case strings.Contains(r.URL.String(), "add_bdd/11"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`invalid`))
			case strings.Contains(r.URL.String(), "update_test_labels/1"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			case strings.Contains(r.URL.String(), "update_test_labels/2"):
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("unauthorized"))
			case strings.Contains(r.URL.String(), "update_tests_labels/1"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			case strings.Contains(r.URL.String(), "update_tests_labels/2"):
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden"))
			case strings.Contains(r.URL.String(), "get_labels/1"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"broken":`))
			case strings.Contains(r.URL.String(), "get_labels/2"):
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not found"))
			case strings.Contains(r.URL.String(), "get_label/1"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			case strings.Contains(r.URL.String(), "get_label/2"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`invalid`))
			case strings.Contains(r.URL.String(), "update_label/1"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"broken":`))
			case strings.Contains(r.URL.String(), "update_label/2"):
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})
		defer s.Close()

		_, err := c.GetBDD(context.Background(), 10)
		if err == nil || !strings.Contains(err.Error(), "error decoding BDD") {
			t.Fatalf("expected get bdd decode error, got: %v", err)
		}

		_, err = c.GetBDD(context.Background(), 11)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected get bdd 403 error, got: %v", err)
		}

		_, err = c.AddBDD(context.Background(), 10, "x")
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected add bdd non-OK error, got: %v", err)
		}

		_, err = c.AddBDD(context.Background(), 11, "x")
		if err == nil || !strings.Contains(err.Error(), "error decoding BDD") {
			t.Fatalf("expected add bdd decode error, got: %v", err)
		}

		err = c.UpdateTestLabels(context.Background(), 1, []string{"x"})
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected update test labels non-OK error, got: %v", err)
		}

		err = c.UpdateTestLabels(context.Background(), 2, []string{"x"})
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected update test labels 401 error, got: %v", err)
		}

		err = c.UpdateTestsLabels(context.Background(), 1, []int64{1}, []string{"x"})
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected update tests labels non-OK error, got: %v", err)
		}

		err = c.UpdateTestsLabels(context.Background(), 2, []int64{1}, []string{"x"})
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected update tests labels 403 error, got: %v", err)
		}

		_, err = c.GetLabels(context.Background(), 1)
		if err == nil || !strings.Contains(err.Error(), "error decoding labels") {
			t.Fatalf("expected labels decode error, got: %v", err)
		}

		_, err = c.GetLabels(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected labels 404 error, got: %v", err)
		}

		_, err = c.GetLabel(context.Background(), 1)
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected get label non-OK error, got: %v", err)
		}

		_, err = c.GetLabel(context.Background(), 2)
		if err == nil || !strings.Contains(err.Error(), "error decoding label") {
			t.Fatalf("expected get label decode error, got: %v", err)
		}

		_, err = c.UpdateLabel(context.Background(), 1, data.UpdateLabelRequest{ProjectID: 1, Title: "x"})
		if err == nil || !strings.Contains(err.Error(), "error decoding label") {
			t.Fatalf("expected update label decode error, got: %v", err)
		}

		_, err = c.UpdateLabel(context.Background(), 2, data.UpdateLabelRequest{ProjectID: 1, Title: "x"})
		if err == nil || !strings.Contains(err.Error(), "API returned") {
			t.Fatalf("expected update label non-OK error, got: %v", err)
		}
	})
}

// TestHTTPExtended_SuccessCases tests success scenarios for all methods
func TestHTTPExtended_SuccessCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.String(), "get_groups/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Group{
				{ID: 1, Name: "QA", ProjectID: 1, UserIDs: []int64{10, 20}},
				{ID: 2, Name: "Dev", ProjectID: 1, UserIDs: []int64{30}},
			})
		case strings.Contains(r.URL.String(), "get_group/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Group{ID: 1, Name: "QA", ProjectID: 1})
		case strings.Contains(r.URL.String(), "add_group/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Group{ID: 100, Name: "NewGroup", ProjectID: 1})
		case strings.Contains(r.URL.String(), "update_group/100"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Group{ID: 100, Name: "UpdatedGroup", ProjectID: 1})
		case strings.Contains(r.URL.String(), "delete_group/100"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.String(), "get_roles"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Role{
				{ID: 1, Name: "Lead"},
				{ID: 2, Name: "Developer"},
				{ID: 3, Name: "Tester"},
			})
		case strings.Contains(r.URL.String(), "get_role/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Role{ID: 1, Name: "Lead"})
		case strings.Contains(r.URL.String(), "get_result_fields"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.ResultField{
				{ID: 1, Name: "Status", SystemName: "status_id", IsActive: true},
				{ID: 2, Name: "CustomField", SystemName: "custom_1", IsActive: true},
			})
		case strings.Contains(r.URL.String(), "get_datasets/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Dataset{
				{ID: 1, Name: "Browsers", ProjectID: 1},
				{ID: 2, Name: "OS", ProjectID: 1},
			})
		case strings.Contains(r.URL.String(), "get_dataset/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Dataset{ID: 1, Name: "Browsers", ProjectID: 1})
		case strings.Contains(r.URL.String(), "add_dataset/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Dataset{ID: 10, Name: "NewDataset", ProjectID: 1})
		case strings.Contains(r.URL.String(), "update_dataset/10"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Dataset{ID: 10, Name: "UpdatedDataset", ProjectID: 1})
		case strings.Contains(r.URL.String(), "delete_dataset/10"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.String(), "get_variables/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Variable{
				{ID: 1, Name: "version", DatasetID: 1},
				{ID: 2, Name: "browser", DatasetID: 1},
			})
		case strings.Contains(r.URL.String(), "add_variable/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Variable{ID: 20, Name: "NewVar", DatasetID: 1})
		case strings.Contains(r.URL.String(), "update_variable/20"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Variable{ID: 20, Name: "UpdatedVar", DatasetID: 1})
		case strings.Contains(r.URL.String(), "delete_variable/20"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.String(), "get_bdd/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.BDD{ID: 1, CaseID: 1, Content: "Given a user"})
		case strings.Contains(r.URL.String(), "add_bdd/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.BDD{ID: 2, CaseID: 1, Content: "When user clicks"})
		case strings.Contains(r.URL.String(), "update_test_labels/10"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.String(), "update_tests_labels/11"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.String(), "get_labels/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Label{
				{ID: 1, Name: "smoke"},
				{ID: 2, Name: "regression"},
			})
		case strings.Contains(r.URL.String(), "get_label/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Label{ID: 1, Name: "smoke"})
		case strings.Contains(r.URL.String(), "update_label/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Label{ID: 1, Name: "updated_label"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test", "test", false)
	ctx := context.Background()

	// Groups test
	groups, err := c.GetGroups(ctx, 1)
	if err != nil || len(groups) != 2 {
		t.Errorf("GetGroups failed: %v, len=%d", err, len(groups))
	}

	group, err := c.GetGroup(ctx, 1)
	if err != nil || group.ID != 1 {
		t.Errorf("GetGroup failed: %v, id=%d", err, group.ID)
	}

	group, err = c.AddGroup(ctx, 1, "NewGroup", []int64{1, 2})
	if err != nil || group.ID != 100 {
		t.Errorf("AddGroup failed: %v, id=%d", err, group.ID)
	}

	group, err = c.UpdateGroup(ctx, 100, "UpdatedGroup", []int64{1})
	if err != nil || group.Name != "UpdatedGroup" {
		t.Errorf("UpdateGroup failed: %v, name=%s", err, group.Name)
	}

	if err = c.DeleteGroup(ctx, 100); err != nil {
		t.Errorf("DeleteGroup failed: %v", err)
	}

	// Roles test
	roles, err := c.GetRoles(ctx)
	if err != nil || len(roles) != 3 {
		t.Errorf("GetRoles failed: %v, len=%d", err, len(roles))
	}

	role, err := c.GetRole(ctx, 1)
	if err != nil || role.ID != 1 {
		t.Errorf("GetRole failed: %v, id=%d", err, role.ID)
	}

	// Result fields test
	fields, err := c.GetResultFields(ctx)
	if err != nil || len(fields) != 2 {
		t.Errorf("GetResultFields failed: %v, len=%d", err, len(fields))
	}

	// Datasets test
	datasets, err := c.GetDatasets(ctx, 1)
	if err != nil || len(datasets) != 2 {
		t.Errorf("GetDatasets failed: %v, len=%d", err, len(datasets))
	}

	dataset, err := c.GetDataset(ctx, 1)
	if err != nil || dataset.ID != 1 {
		t.Errorf("GetDataset failed: %v, id=%d", err, dataset.ID)
	}

	dataset, err = c.AddDataset(ctx, 1, "NewDataset")
	if err != nil || dataset.ID != 10 {
		t.Errorf("AddDataset failed: %v, id=%d", err, dataset.ID)
	}

	dataset, err = c.UpdateDataset(ctx, 10, "UpdatedDataset")
	if err != nil || dataset.Name != "UpdatedDataset" {
		t.Errorf("UpdateDataset failed: %v, name=%s", err, dataset.Name)
	}

	if err = c.DeleteDataset(ctx, 10); err != nil {
		t.Errorf("DeleteDataset failed: %v", err)
	}

	// Variables test
	vars, err := c.GetVariables(ctx, 1)
	if err != nil || len(vars) != 2 {
		t.Errorf("GetVariables failed: %v, len=%d", err, len(vars))
	}

	v, err := c.AddVariable(ctx, 1, "NewVar")
	if err != nil || v.ID != 20 {
		t.Errorf("AddVariable failed: %v, id=%d", err, v.ID)
	}

	v, err = c.UpdateVariable(ctx, 20, "UpdatedVar")
	if err != nil || v.Name != "UpdatedVar" {
		t.Errorf("UpdateVariable failed: %v, name=%s", err, v.Name)
	}

	if err = c.DeleteVariable(ctx, 20); err != nil {
		t.Errorf("DeleteVariable failed: %v", err)
	}

	// BDD test
	bdd, err := c.GetBDD(ctx, 1)
	if err != nil || bdd.ID != 1 {
		t.Errorf("GetBDD failed: %v, id=%d", err, bdd.ID)
	}

	bdd, err = c.AddBDD(ctx, 1, "When user clicks")
	if err != nil || bdd.ID != 2 {
		t.Errorf("AddBDD failed: %v, id=%d", err, bdd.ID)
	}

	// Labels test
	if err = c.UpdateTestLabels(ctx, 10, []string{"smoke"}); err != nil {
		t.Errorf("UpdateTestLabels failed: %v", err)
	}

	if err = c.UpdateTestsLabels(ctx, 11, []int64{1, 2}, []string{"regression"}); err != nil {
		t.Errorf("UpdateTestsLabels failed: %v", err)
	}

	labels, err := c.GetLabels(ctx, 1)
	if err != nil || len(labels) != 2 {
		t.Errorf("GetLabels failed: %v, len=%d", err, len(labels))
	}

	label, err := c.GetLabel(ctx, 1)
	if err != nil || label.ID != 1 {
		t.Errorf("GetLabel failed: %v, id=%d", err, label.ID)
	}

	label, err = c.UpdateLabel(ctx, 1, data.UpdateLabelRequest{ProjectID: 1, Title: "updated_label"})
	if err != nil || label.Name != "updated_label" {
		t.Errorf("UpdateLabel failed: %v, name=%s", err, label.Name)
	}
}

// TestHTTPExtended_EdgeCases tests edge cases and boundary conditions
func TestHTTPExtended_EdgeCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		switch {
		case strings.Contains(url, "get_groups/"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Group{})
		case strings.Contains(url, "get_roles"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Role{})
		case strings.Contains(url, "get_labels/"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Label{})
		case strings.Contains(url, "get_datasets/"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Dataset{})
		case strings.Contains(url, "get_variables/"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Variable{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test", "test", false)
	ctx := context.Background()

	// Empty responses should still work
	groups, err := c.GetGroups(ctx, 999)
	if err != nil {
		t.Fatalf("GetGroups returned unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected empty groups, got %d", len(groups))
	}
}

// TestParallel_ExtendedAPI_AllMethods - параллельное тестирование всех extended API методов
func TestParallel_ExtendedAPI_AllMethods(t *testing.T) {
	// Группы методов для параллельного тестирования
	t.Run("Parallel_Groups", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_groups/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]data.Group{{ID: 1, Name: "G1"}, {ID: 2, Name: "G2"}})
			case strings.Contains(r.URL.String(), "get_group/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Group{ID: 5, Name: "G5"})
			case strings.Contains(r.URL.String(), "add_group/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Group{ID: 99, Name: "NewG"})
			case strings.Contains(r.URL.String(), "update_group/99"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Group{ID: 99, Name: "UpdG"})
			case strings.Contains(r.URL.String(), "delete_group/99"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "t", "t", false)
		ctx := context.Background()

		if g, err := c.GetGroups(ctx, 5); err != nil || len(g) != 2 {
			t.Errorf("GetGroups: %v, len=%d", err, len(g))
		}
		if g, err := c.GetGroup(ctx, 5); err != nil || g.ID != 5 {
			t.Errorf("GetGroup: %v, id=%d", err, g.ID)
		}
		if g, err := c.AddGroup(ctx, 5, "NewG", []int64{1}); err != nil || g.ID != 99 {
			t.Errorf("AddGroup: %v", err)
		}
		if g, err := c.UpdateGroup(ctx, 99, "UpdG", []int64{1}); err != nil || g.Name != "UpdG" {
			t.Errorf("UpdateGroup: %v", err)
		}
		if err := c.DeleteGroup(ctx, 99); err != nil {
			t.Errorf("DeleteGroup: %v", err)
		}
	})

	t.Run("Parallel_Roles", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_roles"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]data.Role{{ID: 1, Name: "R1"}, {ID: 2, Name: "R2"}})
			case strings.Contains(r.URL.String(), "get_role/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Role{ID: 5, Name: "R5"})
			case strings.Contains(r.URL.String(), "get_result_fields"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]data.ResultField{{ID: 1, Name: "F1", SystemName: "f1", IsActive: true}})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "t", "t", false)
		ctx := context.Background()

		if r, err := c.GetRoles(ctx); err != nil || len(r) != 2 {
			t.Errorf("GetRoles: %v", err)
		}
		if r, err := c.GetRole(ctx, 5); err != nil || r.ID != 5 {
			t.Errorf("GetRole: %v", err)
		}
		if f, err := c.GetResultFields(ctx); err != nil || len(f) != 1 {
			t.Errorf("GetResultFields: %v", err)
		}
	})

	t.Run("Parallel_Datasets", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_datasets/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]data.Dataset{{ID: 1, Name: "D1", ProjectID: 5}})
			case strings.Contains(r.URL.String(), "get_dataset/1"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Dataset{ID: 1, Name: "D1", ProjectID: 5})
			case strings.Contains(r.URL.String(), "add_dataset/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Dataset{ID: 99, Name: "NewD", ProjectID: 5})
			case strings.Contains(r.URL.String(), "update_dataset/99"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Dataset{ID: 99, Name: "UpdD", ProjectID: 5})
			case strings.Contains(r.URL.String(), "delete_dataset/99"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "t", "t", false)
		ctx := context.Background()

		if d, err := c.GetDatasets(ctx, 5); err != nil || len(d) != 1 {
			t.Errorf("GetDatasets: %v", err)
		}
		if d, err := c.GetDataset(ctx, 1); err != nil || d.ID != 1 {
			t.Errorf("GetDataset: %v", err)
		}
		if d, err := c.AddDataset(ctx, 5, "NewD"); err != nil || d.ID != 99 {
			t.Errorf("AddDataset: %v", err)
		}
		if d, err := c.UpdateDataset(ctx, 99, "UpdD"); err != nil || d.Name != "UpdD" {
			t.Errorf("UpdateDataset: %v", err)
		}
		if err := c.DeleteDataset(ctx, 99); err != nil {
			t.Errorf("DeleteDataset: %v", err)
		}
	})

	t.Run("Parallel_Variables", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_variables/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]data.Variable{{ID: 1, Name: "V1", DatasetID: 5}})
			case strings.Contains(r.URL.String(), "add_variable/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Variable{ID: 99, Name: "NewV", DatasetID: 5})
			case strings.Contains(r.URL.String(), "update_variable/99"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Variable{ID: 99, Name: "UpdV", DatasetID: 5})
			case strings.Contains(r.URL.String(), "delete_variable/99"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "t", "t", false)
		ctx := context.Background()

		if v, err := c.GetVariables(ctx, 5); err != nil || len(v) != 1 {
			t.Errorf("GetVariables: %v", err)
		}
		if v, err := c.AddVariable(ctx, 5, "NewV"); err != nil || v.ID != 99 {
			t.Errorf("AddVariable: %v", err)
		}
		if v, err := c.UpdateVariable(ctx, 99, "UpdV"); err != nil || v.Name != "UpdV" {
			t.Errorf("UpdateVariable: %v", err)
		}
		if err := c.DeleteVariable(ctx, 99); err != nil {
			t.Errorf("DeleteVariable: %v", err)
		}
	})

	t.Run("Parallel_BDD", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_bdd/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.BDD{ID: 1, CaseID: 5, Content: "Given"})
			case strings.Contains(r.URL.String(), "add_bdd/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.BDD{ID: 99, CaseID: 5, Content: "When"})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "t", "t", false)
		ctx := context.Background()

		if b, err := c.GetBDD(ctx, 5); err != nil || b.ID != 1 {
			t.Errorf("GetBDD: %v", err)
		}
		if b, err := c.AddBDD(ctx, 5, "When"); err != nil || b.ID != 99 {
			t.Errorf("AddBDD: %v", err)
		}
	})

	t.Run("Parallel_Labels", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "get_labels/5"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]data.Label{{ID: 1, Name: "L1"}})
			case strings.Contains(r.URL.String(), "get_label/1"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Label{ID: 1, Name: "L1"})
			case strings.Contains(r.URL.String(), "update_label/1"):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Label{ID: 1, Name: "Updated"})
			case strings.Contains(r.URL.String(), "update_test_labels/5"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			case strings.Contains(r.URL.String(), "update_tests_labels/5"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "t", "t", false)
		ctx := context.Background()

		if l, err := c.GetLabels(ctx, 5); err != nil || len(l) != 1 {
			t.Errorf("GetLabels: %v", err)
		}
		if l, err := c.GetLabel(ctx, 1); err != nil || l.ID != 1 {
			t.Errorf("GetLabel: %v", err)
		}
		if l, err := c.UpdateLabel(ctx, 1, data.UpdateLabelRequest{ProjectID: 5}); err != nil || l.Name != "Updated" {
			t.Errorf("UpdateLabel: %v", err)
		}
		if err := c.UpdateTestLabels(ctx, 5, []string{"L1"}); err != nil {
			t.Errorf("UpdateTestLabels: %v", err)
		}
		if err := c.UpdateTestsLabels(ctx, 5, []int64{1, 2}, []string{"L1"}); err != nil {
			t.Errorf("UpdateTestsLabels: %v", err)
		}
	})
}

func TestExtended_GetGroups_MissingFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_groups/1") {
			w.WriteHeader(http.StatusOK)
			// Sparse data: missing some fields
			_ = json.NewEncoder(w).Encode([]data.Group{
				{ID: 1, Name: "Sparse Group"},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	groups, err := c.GetGroups(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetGroups with sparse data should not error: %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("GetGroups expected 1 group, got %d", len(groups))
	}
}

func TestExtended_GetRoles_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_roles") {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.GetRoles(context.Background())
	if err == nil {
		t.Fatal("GetRoles with 401 should error")
	}
}

func TestExtended_GetResultFields_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_result_fields") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{invalid`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.GetResultFields(context.Background())
	if err == nil {
		t.Fatal("GetResultFields with invalid JSON should error")
	}
}

func TestExtended_GetDatasets_LargeDataset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_datasets/1") {
			w.WriteHeader(http.StatusOK)
			// Large dataset response
			datasets := make([]data.Dataset, 50)
			for i := 0; i < 50; i++ {
				datasets[i] = data.Dataset{ID: int64(i), Name: fmt.Sprintf("Dataset %d", i), ProjectID: 1}
			}
			_ = json.NewEncoder(w).Encode(datasets)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	datasets, err := c.GetDatasets(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDatasets with large data should not error: %v", err)
	}
	if len(datasets) != 50 {
		t.Errorf("GetDatasets expected 50 datasets, got %d", len(datasets))
	}
}

func TestExtended_AddDataset_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "add_dataset/1") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"Dataset name is required"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.AddDataset(context.Background(), 1, "")
	if err == nil {
		t.Fatal("AddDataset with validation error should error")
	}
}

func TestExtended_UpdateDataset_ConflictError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "update_dataset/1") {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":"Dataset already exists"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.UpdateDataset(context.Background(), 1, "NewName")
	if err == nil {
		t.Fatal("UpdateDataset with conflict should error")
	}
}

func TestExtended_DeleteDataset_ForbiddenError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "delete_dataset/1") {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"Cannot delete"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	err := c.DeleteDataset(context.Background(), 1)
	if err == nil {
		t.Fatal("DeleteDataset with 403 should error")
	}
}

func TestExtended_GetVariables_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_variables/1") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Variable{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	vars, err := c.GetVariables(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetVariables should not error on empty list: %v", err)
	}
	if len(vars) != 0 {
		t.Errorf("GetVariables expected empty list, got %d items", len(vars))
	}
}

func TestExtended_AddVariable_InvalidDatasetID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "add_variable") {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"Dataset not found"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	_, err := c.AddVariable(context.Background(), 999, "NewVar")
	if err == nil {
		t.Fatal("AddVariable with invalid dataset should error")
	}
}

func TestExtended_BDD_ErrorCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_bdd/999") {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"BDD not found"}`))
			return
		}
		if strings.Contains(r.URL.String(), "add_bdd/invalid") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"Invalid case ID"}`))
			return
		}
		if strings.Contains(r.URL.String(), "get_bdd/1") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.BDD{ID: 1, CaseID: 1, Content: "BDD"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	ctx := context.Background()

	_, err := c.GetBDD(ctx, 999)
	if err == nil {
		t.Fatal("GetBDD with invalid ID should error")
	}

	bdd, err := c.GetBDD(ctx, 1)
	if err != nil {
		t.Fatalf("GetBDD should work for valid ID: %v", err)
	}
	if bdd.ID != 1 {
		t.Errorf("GetBDD returned wrong ID: %d", bdd.ID)
	}
}

func TestExtended_DeleteMethods_RequestAndSuccessBranches(t *testing.T) {
	t.Run("delete methods success", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.String(), "delete_group/17"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			case strings.Contains(r.URL.String(), "delete_dataset/18"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			case strings.Contains(r.URL.String(), "delete_variable/19"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			default:
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
		})
		defer s.Close()

		ctx := context.Background()
		if err := c.DeleteGroup(ctx, 17); err != nil {
			t.Fatalf("DeleteGroup success branch failed: %v", err)
		}
		if err := c.DeleteDataset(ctx, 18); err != nil {
			t.Fatalf("DeleteDataset success branch failed: %v", err)
		}
		if err := c.DeleteVariable(ctx, 19); err != nil {
			t.Fatalf("DeleteVariable success branch failed: %v", err)
		}
	})

	t.Run("delete methods request error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		ctx := context.Background()
		if err := c.DeleteGroup(ctx, 17); err == nil {
			t.Fatal("expected DeleteGroup request error")
		}
		if err := c.DeleteDataset(ctx, 18); err == nil {
			t.Fatal("expected DeleteDataset request error")
		}
		if err := c.DeleteVariable(ctx, 19); err == nil {
			t.Fatal("expected DeleteVariable request error")
		}
	})
}
