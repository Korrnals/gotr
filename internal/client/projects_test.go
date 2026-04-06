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

func TestHTTPGetProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.String(), "get_projects") {
			t.Fatalf("expected get_projects endpoint, got %s", r.URL.String())
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]data.Project{{ID: 1, Name: "P1"}, {ID: 2, Name: "P2"}})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test", "test", false)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	projects, err := client.GetProjects(context.Background())
	if err != nil {
		t.Fatalf("GetProjects() error = %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("GetProjects() returned %d items, want 2", len(projects))
	}
}

func TestHTTPGetProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.String(), "get_project/42") {
			t.Fatalf("expected get_project/42 endpoint, got %s", r.URL.String())
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(data.Project{ID: 42, Name: "Backend"})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test", "test", false)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	project, err := client.GetProject(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}
	if project.ID != 42 || project.Name != "Backend" {
		t.Fatalf("GetProject() unexpected project: %+v", project)
	}
}

func TestHTTPAddProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.String(), "add_project") {
			t.Fatalf("expected add_project endpoint, got %s", r.URL.String())
		}

		var req data.AddProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request error: %v", err)
		}
		if req.Name != "New Project" {
			t.Fatalf("expected request name New Project, got %q", req.Name)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(data.Project{ID: 11, Name: req.Name})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test", "test", false)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	project, err := client.AddProject(context.Background(), &data.AddProjectRequest{Name: "New Project"})
	if err != nil {
		t.Fatalf("AddProject() error = %v", err)
	}
	if project.ID != 11 || project.Name != "New Project" {
		t.Fatalf("AddProject() unexpected project: %+v", project)
	}
}

func TestHTTPUpdateProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.String(), "update_project/7") {
			t.Fatalf("expected update_project/7 endpoint, got %s", r.URL.String())
		}

		var req data.UpdateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request error: %v", err)
		}
		if req.Name != "Updated" {
			t.Fatalf("expected request name Updated, got %q", req.Name)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(data.Project{ID: 7, Name: req.Name})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test", "test", false)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	project, err := client.UpdateProject(context.Background(), 7, &data.UpdateProjectRequest{Name: "Updated"})
	if err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}
	if project.ID != 7 || project.Name != "Updated" {
		t.Fatalf("UpdateProject() unexpected project: %+v", project)
	}
}

func TestHTTPDeleteProject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.String(), "delete_project/9") {
				t.Fatalf("expected delete_project/9 endpoint, got %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client, err := NewClient(server.URL, "test", "test", false)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}

		if err := client.DeleteProject(context.Background(), 9); err != nil {
			t.Fatalf("DeleteProject() error = %v", err)
		}
	})

	t.Run("non-200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"cannot delete"}`))
		}))
		defer server.Close()

		client, err := NewClient(server.URL, "test", "test", false)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}

		err = client.DeleteProject(context.Background(), 9)
		if err == nil {
			t.Fatalf("expected DeleteProject() error for non-200 response")
		}
		if !strings.Contains(err.Error(), "cannot delete") {
			t.Fatalf("expected error to include response body, got: %v", err)
		}
	})
}
