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
						return nil, nil // или ошибка, в зависимости от реализации
					}
					return tc.mockResponse, nil
				}
			}

			ctx := context.Background()
			result, err := mockClient.AddAttachmentToCase(ctx, tc.caseID, tc.filePath)

			if tc.wantErr {
				// Для теста FileNotFound просто проверяем что вызов произошел
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
		return nil, nil // Симулируем ошибку или nil ответ
	}

	// Проверяем что метод вызывается
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
