// internal/client/reports_test.go
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

func TestGetReports(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.GetReportsFunc = func(ctx context.Context, projectID int64) (data.GetReportsResponse, error) {
		return []data.ReportTemplate{
			{ID: 1, Name: "Test Summary", Description: "Summary report"},
			{ID: 2, Name: "Detailed Report", Description: "Detailed report"},
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.GetReports(ctx, 1)
	if err != nil {
		t.Errorf("GetReports() unexpected error: %v", err)
		return
	}
	if len(result) != 2 {
		t.Errorf("GetReports() returned %d reports, want 2", len(result))
	}
}

func TestRunReport(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.RunReportFunc = func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
		return &data.RunReportResponse{
			ReportID: 123,
			URL:      "https://example.com/report/123",
			Status:   "pending",
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.RunReport(ctx, 1)
	if err != nil {
		t.Errorf("RunReport() unexpected error: %v", err)
		return
	}
	if result.ReportID != 123 {
		t.Errorf("RunReport() ReportID = %d, want 123", result.ReportID)
	}
}

func TestRunCrossProjectReport(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.RunCrossProjectReportFunc = func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
		return &data.RunReportResponse{
			ReportID: 456,
			URL:      "https://example.com/report/456",
			Status:   "completed",
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.RunCrossProjectReport(ctx, 1)
	if err != nil {
		t.Errorf("RunCrossProjectReport() unexpected error: %v", err)
		return
	}
	if result.ReportID != 456 {
		t.Errorf("RunCrossProjectReport() ReportID = %d, want 456", result.ReportID)
	}
}

func TestHTTPGetReports(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]data.ReportTemplate{
			{ID: 1, Name: "Test Report"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	ctx := context.Background()
	reports, err := client.GetReports(ctx, 1)
	if err != nil {
		t.Fatalf("GetReports() error: %v", err)
	}
	if len(reports) != 1 {
		t.Errorf("Expected 1 report template, got %d", len(reports))
	}
}

func TestHTTPGetReports_ErrorBranches(t *testing.T) {
	t.Run("transport error", func(t *testing.T) {
		client, _ := NewClient("http://127.0.0.1:1", "test", "test", false)
		_, err := client.GetReports(context.Background(), 1)
		if err == nil {
			t.Fatalf("expected transport error")
		}
	})

	t.Run("non-200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"bad request"}`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		_, err := client.GetReports(context.Background(), 1)
		if err == nil {
			t.Fatalf("expected error for non-200 response")
		}
	})

	t.Run("decode error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{invalid`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		_, err := client.GetReports(context.Background(), 1)
		if err == nil {
			t.Fatalf("expected decode error")
		}
	})
}

func TestHTTPRunReport(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.String(), "run_report/10") {
				t.Fatalf("expected run_report/10 endpoint, got %s", r.URL.String())
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.RunReportResponse{ReportID: 10, URL: "https://example/report/10", Status: "queued"})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		resp, err := client.RunReport(context.Background(), 10)
		if err != nil {
			t.Fatalf("RunReport() error = %v", err)
		}
		if resp.ReportID != 10 {
			t.Fatalf("RunReport() ReportID = %d, want 10", resp.ReportID)
		}
	})

	t.Run("non-200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"bad report"}`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		_, err := client.RunReport(context.Background(), 10)
		if err == nil {
			t.Fatalf("expected RunReport() error for non-200 status")
		}
	})
}

func TestHTTPRunCrossProjectReport(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.String(), "run_cross_project_report/20") {
				t.Fatalf("expected run_cross_project_report/20 endpoint, got %s", r.URL.String())
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.RunReportResponse{ReportID: 20, URL: "https://example/report/20", Status: "running"})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		resp, err := client.RunCrossProjectReport(context.Background(), 20)
		if err != nil {
			t.Fatalf("RunCrossProjectReport() error = %v", err)
		}
		if resp.ReportID != 20 {
			t.Fatalf("RunCrossProjectReport() ReportID = %d, want 20", resp.ReportID)
		}
	})

	t.Run("decode error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{invalid`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		_, err := client.RunCrossProjectReport(context.Background(), 20)
		if err == nil {
			t.Fatalf("expected decode error")
		}
	})
}

func TestHTTPGetCrossProjectReports(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.String(), "get_cross_project_reports") {
				t.Fatalf("expected get_cross_project_reports endpoint, got %s", r.URL.String())
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.ReportTemplate{{ID: 31, Name: "Cross"}})
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		reports, err := client.GetCrossProjectReports(context.Background())
		if err != nil {
			t.Fatalf("GetCrossProjectReports() error = %v", err)
		}
		if len(reports) != 1 || reports[0].ID != 31 {
			t.Fatalf("unexpected reports payload: %+v", reports)
		}
	})

	t.Run("non-200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"forbidden"}`))
		}))
		defer server.Close()

		client, _ := NewClient(server.URL, "test", "test", false)
		_, err := client.GetCrossProjectReports(context.Background())
		if err == nil {
			t.Fatalf("expected error for non-200 status")
		}
	})
}
