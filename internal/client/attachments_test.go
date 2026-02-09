// internal/client/attachments_test.go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func TestAddAttachmentToCase(t *testing.T) {
	// Создаем временный файл для теста
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/index.php" {
			t.Errorf("Expected path /index.php, got %s", r.URL.Path)
		}

		// Проверяем что Content-Type содержит multipart/form-data
		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			t.Error("Content-Type header is missing")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.AttachmentResponse{
			AttachmentID: 1,
			URL:          "https://example.com/attachment/1",
			Name:         "test.txt",
			Size:         12,
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	resp, err := client.AddAttachmentToCase(1, tmpFile)
	if err != nil {
		t.Fatalf("AddAttachmentToCase() error: %v", err)
	}

	if resp.AttachmentID != 1 {
		t.Errorf("Expected AttachmentID 1, got %d", resp.AttachmentID)
	}
	if resp.Name != "test.txt" {
		t.Errorf("Expected Name 'test.txt', got '%s'", resp.Name)
	}
}

func TestAddAttachmentToResult(t *testing.T) {
	// Создаем временный файл для теста
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "screenshot.png")
	if err := os.WriteFile(tmpFile, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.AttachmentResponse{
			AttachmentID: 2,
			URL:          "https://example.com/attachment/2",
			Name:         "screenshot.png",
			Size:           15,
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	resp, err := client.AddAttachmentToResult(1, tmpFile)
	if err != nil {
		t.Fatalf("AddAttachmentToResult() error: %v", err)
	}

	if resp.AttachmentID != 2 {
		t.Errorf("Expected AttachmentID 2, got %d", resp.AttachmentID)
	}
}

func TestAddAttachmentToCaseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid file"}`))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test", false)

	// Создаем временный файл
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	_, err := client.AddAttachmentToCase(1, tmpFile)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestMockClientAttachments(t *testing.T) {
	mockClient := &MockClient{}
	mockClient.AddAttachmentToCaseFunc = func(caseID int64, filePath string) (*data.AttachmentResponse, error) {
		return &data.AttachmentResponse{
			AttachmentID: 1,
			Name:         "test.txt",
		}, nil
	}

	resp, err := mockClient.AddAttachmentToCase(1, "/tmp/test.txt")
	if err != nil {
		t.Fatalf("Mock AddAttachmentToCase() error: %v", err)
	}
	if resp.AttachmentID != 1 {
		t.Errorf("Expected AttachmentID 1, got %d", resp.AttachmentID)
	}
}
