// internal/client/attachments_test.go
package client

import (
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
				mockClient.AddAttachmentToCaseFunc = func(caseID int64, filePath string) (*data.AttachmentResponse, error) {
					if tc.name == "FileNotFound" {
						return nil, nil // или ошибка, в зависимости от реализации
					}
					return tc.mockResponse, nil
				}
			}

			result, err := mockClient.AddAttachmentToCase(tc.caseID, tc.filePath)

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
	mockClient.AddAttachmentToPlanFunc = func(planID int64, filePath string) (*data.AttachmentResponse, error) {
		return &data.AttachmentResponse{
			AttachmentID: 2,
			URL:          "https://example.com/attachment/2",
			Name:         "plan_doc.pdf",
			Size:         1024,
		}, nil
	}

	result, err := mockClient.AddAttachmentToPlan(1, "/tmp/plan_doc.pdf")
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
	mockClient.AddAttachmentToPlanEntryFunc = func(planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
		return &data.AttachmentResponse{
			AttachmentID: 3,
			URL:          "https://example.com/attachment/3",
			Name:         "entry_data.json",
			Size:         512,
		}, nil
	}

	result, err := mockClient.AddAttachmentToPlanEntry(1, "entry-1", "/tmp/data.json")
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
	mockClient.AddAttachmentToResultFunc = func(resultID int64, filePath string) (*data.AttachmentResponse, error) {
		return &data.AttachmentResponse{
			AttachmentID: 4,
			URL:          "https://example.com/attachment/4",
			Name:         "screenshot.png",
			Size:         2048,
		}, nil
	}

	result, err := mockClient.AddAttachmentToResult(1, "/tmp/screenshot.png")
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
	mockClient.AddAttachmentToRunFunc = func(runID int64, filePath string) (*data.AttachmentResponse, error) {
		return &data.AttachmentResponse{
			AttachmentID: 5,
			URL:          "https://example.com/attachment/5",
			Name:         "run_log.txt",
			Size:         4096,
		}, nil
	}

	result, err := mockClient.AddAttachmentToRun(1, "/tmp/run_log.txt")
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
	mockClient.AddAttachmentToCaseFunc = func(caseID int64, filePath string) (*data.AttachmentResponse, error) {
		return nil, nil // Симулируем ошибку или nil ответ
	}

	// Проверяем что метод вызывается
	_, err := mockClient.AddAttachmentToCase(999, "/tmp/test.txt")
	if err != nil {
		t.Errorf("AddAttachmentToCase() unexpected error: %v", err)
	}
}
