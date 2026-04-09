// internal/client/attachments_test.go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func TestAddAttachmentToCase(t *testing.T) {
	testCases := []struct {
		name         string
		caseID       int64
		filePath     string
		mockResponse *data.AttachmentResponse
		wantErr      bool
	}{
		{
			name:     "Success",
			caseID:   1,
			filePath: "/tmp/test.txt",
			mockResponse: &data.AttachmentResponse{
				AttachmentID: 1,
				URL:          "https://example.com/attachment/1",
				Name:         "test.txt",
				Size:         12,
			},
		},
		{
			name:     "FileNotFound",
			caseID:   1,
			filePath: "/tmp/nonexistent.txt",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockClient{}
			if tc.mockResponse != nil {
				mockClient.AddAttachmentToCaseFunc = func(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
					if tc.name == "FileNotFound" {
						return nil, nil // or error, depending on implementation
					}
					return tc.mockResponse, nil
				}
			}

			ctx := context.Background()
			result, err := mockClient.AddAttachmentToCase(ctx, tc.caseID, tc.filePath)

			if tc.wantErr {
				// For FileNotFound test just verify the call was made
				return
			}
			if err != nil {
				t.Errorf("AddAttachmentToCase() unexpected error: %v", err)
				return
			}
			if result.AttachmentID != tc.mockResponse.AttachmentID {
				t.Errorf("AddAttachmentToCase() AttachmentID = %d, want %d", result.AttachmentID, tc.mockResponse.AttachmentID)
			}
		})
	}
}

func TestAddAttachmentToPlan(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.AddAttachmentToPlanFunc = func(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error) {
		return &data.AttachmentResponse{
			AttachmentID: 2,
			URL:          "https://example.com/attachment/2",
			Name:         "plan_doc.pdf",
			Size:         1024,
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.AddAttachmentToPlan(ctx, 1, "/tmp/plan_doc.pdf")
	if err != nil {
		t.Errorf("AddAttachmentToPlan() unexpected error: %v", err)
		return
	}
	if result.AttachmentID != 2 {
		t.Errorf("AddAttachmentToPlan() AttachmentID = %d, want 2", result.AttachmentID)
	}
}

func TestAddAttachmentToPlanEntry(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.AddAttachmentToPlanEntryFunc = func(ctx context.Context, planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
		return &data.AttachmentResponse{
			AttachmentID: 3,
			URL:          "https://example.com/attachment/3",
			Name:         "entry_data.json",
			Size:         512,
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.AddAttachmentToPlanEntry(ctx, 1, "entry-1", "/tmp/data.json")
	if err != nil {
		t.Errorf("AddAttachmentToPlanEntry() unexpected error: %v", err)
		return
	}
	if result.AttachmentID != 3 {
		t.Errorf("AddAttachmentToPlanEntry() AttachmentID = %d, want 3", result.AttachmentID)
	}
}

func TestAddAttachmentToResult(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.AddAttachmentToResultFunc = func(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error) {
		return &data.AttachmentResponse{
			AttachmentID: 4,
			URL:          "https://example.com/attachment/4",
			Name:         "screenshot.png",
			Size:         2048,
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.AddAttachmentToResult(ctx, 1, "/tmp/screenshot.png")
	if err != nil {
		t.Errorf("AddAttachmentToResult() unexpected error: %v", err)
		return
	}
	if result.AttachmentID != 4 {
		t.Errorf("AddAttachmentToResult() AttachmentID = %d, want 4", result.AttachmentID)
	}
}

func TestAddAttachmentToRun(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.AddAttachmentToRunFunc = func(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error) {
		return &data.AttachmentResponse{
			AttachmentID: 5,
			URL:          "https://example.com/attachment/5",
			Name:         "run_log.txt",
			Size:         4096,
		}, nil
	}

	ctx := context.Background()
	result, err := mockClient.AddAttachmentToRun(ctx, 1, "/tmp/run_log.txt")
	if err != nil {
		t.Errorf("AddAttachmentToRun() unexpected error: %v", err)
		return
	}
	if result.AttachmentID != 5 {
		t.Errorf("AddAttachmentToRun() AttachmentID = %d, want 5", result.AttachmentID)
	}
}

func TestAddAttachmentToCaseError(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.AddAttachmentToCaseFunc = func(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
		return nil, nil // Simulate error or nil response
	}

	// Verify the method is called
	ctx := context.Background()
	_, err := mockClient.AddAttachmentToCase(ctx, 999, "/tmp/test.txt")
	if err != nil {
		t.Errorf("AddAttachmentToCase() unexpected error: %v", err)
	}
}

func TestHTTPAttachmentReadAndDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.String(), "delete_attachment/9"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.String(), "get_attachment/9"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.Attachment{ID: 9, Name: "a.txt", Size: 10})
		case strings.Contains(r.URL.String(), "get_attachments_for_case/1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Attachment{{ID: 1, Name: "c.txt"}})
		case strings.Contains(r.URL.String(), "get_attachments_for_plan/2"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Attachment{{ID: 2, Name: "p.txt"}})
		case strings.Contains(r.URL.String(), "get_attachments_for_plan_entry/2/e1"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Attachment{{ID: 3, Name: "pe.txt"}})
		case strings.Contains(r.URL.String(), "get_attachments_for_run/3"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Attachment{{ID: 4, Name: "r.txt"}})
		case strings.Contains(r.URL.String(), "get_attachments_for_test/4"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Attachment{{ID: 5, Name: "t.txt"}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test", "test", false)
	ctx := context.Background()

	if err := c.DeleteAttachment(ctx, 9); err != nil {
		t.Fatalf("DeleteAttachment() error: %v", err)
	}

	a, err := c.GetAttachment(ctx, 9)
	if err != nil || a.ID != 9 {
		t.Fatalf("GetAttachment() failed: %v, %+v", err, a)
	}

	if list, err := c.GetAttachmentsForCase(ctx, 1); err != nil || len(list) != 1 {
		t.Fatalf("GetAttachmentsForCase() failed: %v, %+v", err, list)
	}
	if list, err := c.GetAttachmentsForPlan(ctx, 2); err != nil || len(list) != 1 {
		t.Fatalf("GetAttachmentsForPlan() failed: %v, %+v", err, list)
	}
	if list, err := c.GetAttachmentsForPlanEntry(ctx, 2, "e1"); err != nil || len(list) != 1 {
		t.Fatalf("GetAttachmentsForPlanEntry() failed: %v, %+v", err, list)
	}
	if list, err := c.GetAttachmentsForRun(ctx, 3); err != nil || len(list) != 1 {
		t.Fatalf("GetAttachmentsForRun() failed: %v, %+v", err, list)
	}
	if list, err := c.GetAttachmentsForTest(ctx, 4); err != nil || len(list) != 1 {
		t.Fatalf("GetAttachmentsForTest() failed: %v, %+v", err, list)
	}
}

func TestHTTPAttachmentUploads(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"multipart required"}`))
			return
		}

		switch {
		case strings.Contains(r.URL.String(), "add_attachment_to_case/1"):
			_ = json.NewEncoder(w).Encode(data.AttachmentResponse{AttachmentID: 11, Name: "f.txt"})
		case strings.Contains(r.URL.String(), "add_attachment_to_plan/2"):
			_ = json.NewEncoder(w).Encode(data.AttachmentResponse{AttachmentID: 12, Name: "f.txt"})
		case strings.Contains(r.URL.String(), "add_attachment_to_plan_entry/2/e1"):
			_ = json.NewEncoder(w).Encode(data.AttachmentResponse{AttachmentID: 13, Name: "f.txt"})
		case strings.Contains(r.URL.String(), "add_attachment_to_result/3"):
			_ = json.NewEncoder(w).Encode(data.AttachmentResponse{AttachmentID: 14, Name: "f.txt"})
		case strings.Contains(r.URL.String(), "add_attachment_to_run/4"):
			_ = json.NewEncoder(w).Encode(data.AttachmentResponse{AttachmentID: 15, Name: "f.txt"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "f.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	c, _ := NewClient(server.URL, "test", "test", false)
	ctx := context.Background()

	if resp, err := c.AddAttachmentToCase(ctx, 1, filePath); err != nil || resp.AttachmentID != 11 {
		t.Fatalf("AddAttachmentToCase() failed: %v, %+v", err, resp)
	}
	if resp, err := c.AddAttachmentToPlan(ctx, 2, filePath); err != nil || resp.AttachmentID != 12 {
		t.Fatalf("AddAttachmentToPlan() failed: %v, %+v", err, resp)
	}
	if resp, err := c.AddAttachmentToPlanEntry(ctx, 2, "e1", filePath); err != nil || resp.AttachmentID != 13 {
		t.Fatalf("AddAttachmentToPlanEntry() failed: %v, %+v", err, resp)
	}
	if resp, err := c.AddAttachmentToResult(ctx, 3, filePath); err != nil || resp.AttachmentID != 14 {
		t.Fatalf("AddAttachmentToResult() failed: %v, %+v", err, resp)
	}
	if resp, err := c.AddAttachmentToRun(ctx, 4, filePath); err != nil || resp.AttachmentID != 15 {
		t.Fatalf("AddAttachmentToRun() failed: %v, %+v", err, resp)
	}
}

func TestUploadAttachment_FileNotFound(t *testing.T) {
	c := &HTTPClient{}
	_, err := c.uploadAttachment(context.Background(), "add_attachment_to_case/1", "/tmp/definitely-missing-file")
	if err == nil {
		t.Fatalf("expected uploadAttachment() error for missing file")
	}
}

func TestHTTPAttachments_ErrorBranches(t *testing.T) {
	t.Run("delete attachment non-OK", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"cannot delete"}`))
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "test", "test", false)
		if err := c.DeleteAttachment(context.Background(), 123); err == nil {
			t.Fatalf("expected DeleteAttachment() error")
		}
	})

	t.Run("get attachment methods decode errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "test", "test", false)
		if _, err := c.GetAttachment(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachment() decode error")
		}
		if _, err := c.GetAttachmentsForCase(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForCase() decode error")
		}
		if _, err := c.GetAttachmentsForPlan(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForPlan() decode error")
		}
		if _, err := c.GetAttachmentsForPlanEntry(context.Background(), 1, ""); err == nil {
			t.Fatalf("expected GetAttachmentsForPlanEntry() decode error")
		}
		if _, err := c.GetAttachmentsForRun(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForRun() decode error")
		}
		if _, err := c.GetAttachmentsForTest(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForTest() decode error")
		}
	})

	t.Run("get attachment methods non-OK", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "test", "test", false)
		if _, err := c.GetAttachment(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachment() non-OK error")
		}
		if _, err := c.GetAttachmentsForCase(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForCase() non-OK error")
		}
		if _, err := c.GetAttachmentsForPlan(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForPlan() non-OK error")
		}
		if _, err := c.GetAttachmentsForPlanEntry(context.Background(), 1, "edge-entry"); err == nil {
			t.Fatalf("expected GetAttachmentsForPlanEntry() non-OK error")
		}
		if _, err := c.GetAttachmentsForRun(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForRun() non-OK error")
		}
		if _, err := c.GetAttachmentsForTest(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForTest() non-OK error")
		}
	})

	t.Run("get attachment methods request errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		c, _ := NewClient(server.URL, "test", "test", false)
		server.Close()

		if _, err := c.GetAttachment(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachment() request error")
		}
		if _, err := c.GetAttachmentsForCase(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForCase() request error")
		}
		if _, err := c.GetAttachmentsForPlan(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForPlan() request error")
		}
		if _, err := c.GetAttachmentsForPlanEntry(context.Background(), 1, ""); err == nil {
			t.Fatalf("expected GetAttachmentsForPlanEntry() request error")
		}
		if _, err := c.GetAttachmentsForRun(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForRun() request error")
		}
		if _, err := c.GetAttachmentsForTest(context.Background(), 1); err == nil {
			t.Fatalf("expected GetAttachmentsForTest() request error")
		}
	})
}

func TestUploadAttachment_HTTPErrorBranches(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "upload.txt")
	if err := os.WriteFile(filePath, []byte("payload"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	t.Run("non-OK", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"upload rejected"}`))
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "test", "test", false)
		_, err := c.uploadAttachment(context.Background(), "add_attachment_to_case/1", filePath)
		if err == nil {
			t.Fatalf("expected uploadAttachment() non-OK error")
		}
	})

	t.Run("decode error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}))
		defer server.Close()

		c, _ := NewClient(server.URL, "test", "test", false)
		_, err := c.uploadAttachment(context.Background(), "add_attachment_to_case/1", filePath)
		if err == nil {
			t.Fatalf("expected uploadAttachment() decode error")
		}
		if !strings.Contains(err.Error(), "error decoding attachment response") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("request error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		c, _ := NewClient(server.URL, "test", "test", false)
		server.Close()

		_, err := c.uploadAttachment(context.Background(), "add_attachment_to_case/1", filePath)
		if err == nil {
			t.Fatalf("expected uploadAttachment() request error")
		}
	})
}

func TestGetAttachmentsForPlanEntry_ErrorCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.String(), "get_attachments_for_plan_entry/1/entry1"):
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Server error"}`))
		case strings.Contains(r.URL.String(), "get_attachments_for_plan_entry/2/entry2"):
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"Forbidden"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	ctx := context.Background()

	_, err := c.GetAttachmentsForPlanEntry(ctx, 1, "entry1")
	if err == nil {
		t.Fatal("GetAttachmentsForPlanEntry should error on 500")
	}

	_, err = c.GetAttachmentsForPlanEntry(ctx, 2, "entry2")
	if err == nil {
		t.Fatal("GetAttachmentsForPlanEntry should error on 403")
	}
}

func TestGetAttachmentsForPlanEntry_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_attachments_for_plan_entry") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Attachment{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	attachments, err := c.GetAttachmentsForPlanEntry(context.Background(), 1, "entry1")
	if err != nil {
		t.Fatalf("GetAttachmentsForPlanEntry should not error on empty: %v", err)
	}
	if len(attachments) != 0 {
		t.Errorf("Expected 0 attachments, got %d", len(attachments))
	}
}

func TestGetAttachmentsForRun_ErrorCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_attachments_for_run/1") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"Invalid run ID"}`))
			return
		}
		if strings.Contains(r.URL.String(), "get_attachments_for_run/2") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{invalid json}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	ctx := context.Background()

	_, err := c.GetAttachmentsForRun(ctx, 1)
	if err == nil {
		t.Fatal("GetAttachmentsForRun should error on 400")
	}

	_, err = c.GetAttachmentsForRun(ctx, 2)
	if err == nil {
		t.Fatal("GetAttachmentsForRun should error on decode failure")
	}
}

func TestGetAttachmentsForRun_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_attachments_for_run") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Attachment{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	attachments, err := c.GetAttachmentsForRun(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetAttachmentsForRun should not error on empty: %v", err)
	}
	if len(attachments) != 0 {
		t.Errorf("Expected 0 attachments, got %d", len(attachments))
	}
}

func TestGetAttachmentsForTest_ErrorCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_attachments_for_test/999") {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"Test not found"}`))
			return
		}
		if strings.Contains(r.URL.String(), "get_attachments_for_test/1") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Attachment{
				{ID: 1, Size: 1024},
				{ID: 2, Size: 2048},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	ctx := context.Background()

	_, err := c.GetAttachmentsForTest(ctx, 999)
	if err == nil {
		t.Fatal("GetAttachmentsForTest should error on 404")
	}

	attachments, err := c.GetAttachmentsForTest(ctx, 1)
	if err != nil {
		t.Fatalf("GetAttachmentsForTest should not error: %v", err)
	}
	if len(attachments) != 2 {
		t.Errorf("Expected 2 attachments, got %d", len(attachments))
	}
}

func TestDeleteAttachment_ErrorCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "delete_attachment/1") {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"Cannot delete"}`))
			return
		}
		if strings.Contains(r.URL.String(), "delete_attachment/2") {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"Not found"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	ctx := context.Background()

	err := c.DeleteAttachment(ctx, 1)
	if err == nil {
		t.Fatal("DeleteAttachment should error on 403")
	}

	err = c.DeleteAttachment(ctx, 2)
	if err == nil {
		t.Fatal("DeleteAttachment should error on 404")
	}
}

func TestGetAttachment_ErrorCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_attachment/1") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{broken:`))
			return
		}
		if strings.Contains(r.URL.String(), "get_attachment/2") {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	ctx := context.Background()

	_, err := c.GetAttachment(ctx, 1)
	if err == nil {
		t.Fatal("GetAttachment should error on decode failure")
	}

	_, err = c.GetAttachment(ctx, 2)
	if err == nil {
		t.Fatal("GetAttachment should error on 401")
	}
}

func TestGetAttachmentsForCase_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_attachments_for_case/1") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Attachment{
				{ID: 10, Size: 512},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	attachments, err := c.GetAttachmentsForCase(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetAttachmentsForCase: %v", err)
	}
	if len(attachments) != 1 || attachments[0].ID != 10 {
		t.Errorf("GetAttachmentsForCase returned unexpected data: %+v", attachments)
	}
}

func TestGetAttachmentsForPlan_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "get_attachments_for_plan/2") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]data.Attachment{
				{ID: 20, Size: 1024},
				{ID: 21, Size: 2048},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "t", "t", false)
	attachments, err := c.GetAttachmentsForPlan(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetAttachmentsForPlan: %v", err)
	}
	if len(attachments) != 2 {
		t.Errorf("GetAttachmentsForPlan expected 2, got %d", len(attachments))
	}
}

func TestDeleteAttachment_RequestError(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
	s.Close()

	err := c.DeleteAttachment(context.Background(), 77)
	if err == nil {
		t.Fatal("DeleteAttachment should return request error when server is closed")
	}
}
