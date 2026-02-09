// internal/client/reports_test.go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func TestGetReports(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.GetReportsFunc = func(projectID int64) (data.GetReportsResponse, error) {
		return []data.ReportTemplate{
			{ID: 1, Name: "Test Summary", Description: "Summary report"},
			{ID: 2, Name: "Detailed Report", Description: "Detailed report"},
		}, nil
	}

	result, err := mockClient.GetReports(1)
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
	mockClient.RunReportFunc = func(templateID int64) (*data.RunReportResponse, error) {
		return &data.RunReportResponse{
			ReportID: 123,
			URL:      "https://example.com/report/123",
			Status:   "pending",
		}, nil
	}

	result, err := mockClient.RunReport(1)
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
	mockClient.RunCrossProjectReportFunc = func(templateID int64) (*data.RunReportResponse, error) {
		return &data.RunReportResponse{
			ReportID: 456,
			URL:      "https://example.com/report/456",
			Status:   "completed",
		}, nil
	}

	result, err := mockClient.RunCrossProjectReport(1)
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

	reports, err := client.GetReports(1)
	if err != nil {
		t.Fatalf("GetReports() error: %v", err)
	}
	if len(reports) != 1 {
		t.Errorf("Expected 1 report template, got %d", len(reports))
	}
}
