// internal/client/attachments.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/models/data"
)

// DeleteAttachment удаляет вложение по ID
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#deleteattachment
func (c *HTTPClient) DeleteAttachment(ctx context.Context, attachmentID int64) error {
	endpoint := fmt.Sprintf("delete_attachment/%d", attachmentID)
	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting attachment %d: %w", attachmentID, err)
	}
	defer resp.Body.Close()

	return nil
}

// GetAttachment получает информацию о вложении по ID
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachment
func (c *HTTPClient) GetAttachment(ctx context.Context, attachmentID int64) (*data.Attachment, error) {
	endpoint := fmt.Sprintf("get_attachment/%d", attachmentID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachment %d: %w", attachmentID, err)
	}
	defer resp.Body.Close()

	var attachment data.Attachment
	if err := json.NewDecoder(resp.Body).Decode(&attachment); err != nil {
		return nil, fmt.Errorf("error decoding attachment %d: %w", attachmentID, err)
	}

	return &attachment, nil
}

// GetAttachmentsForCase получает список вложений для тест-case
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsforcase
func (c *HTTPClient) GetAttachmentsForCase(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_case/%d", caseID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for case %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for case %d: %w", caseID, err)
	}

	return attachments, nil
}

// GetAttachmentsForPlan получает список вложений для тест-плана
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsforplan
func (c *HTTPClient) GetAttachmentsForPlan(ctx context.Context, planID int64) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_plan/%d", planID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for plan %d: %w", planID, err)
	}
	defer resp.Body.Close()

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for plan %d: %w", planID, err)
	}

	return attachments, nil
}

// GetAttachmentsForPlanEntry получает список вложений для записи в тест-плане
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsforplanentry
func (c *HTTPClient) GetAttachmentsForPlanEntry(ctx context.Context, planID int64, entryID string) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_plan_entry/%d/%s", planID, entryID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for plan entry %s in plan %d: %w", entryID, planID, err)
	}
	defer resp.Body.Close()

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for plan entry %s: %w", entryID, err)
	}

	return attachments, nil
}

// GetAttachmentsForRun получает список вложений для тест-run
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsforrun
func (c *HTTPClient) GetAttachmentsForRun(ctx context.Context, runID int64) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_run/%d", runID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for run %d: %w", runID, err)
	}
	defer resp.Body.Close()

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for run %d: %w", runID, err)
	}

	return attachments, nil
}

// GetAttachmentsForTest получает список вложений for test
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsfortest
func (c *HTTPClient) GetAttachmentsForTest(ctx context.Context, testID int64) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_test/%d", testID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for test %d: %w", testID, err)
	}
	defer resp.Body.Close()

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for test %d: %w", testID, err)
	}

	return attachments, nil
}

// AddAttachmentToCase добавляет вложение к тест-кейсу
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttocase
func (c *HTTPClient) AddAttachmentToCase(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_case/%d", caseID)
	return c.uploadAttachment(ctx, endpoint, filePath)
}

// AddAttachmentToPlan добавляет вложение к тест-плану
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttoplan
func (c *HTTPClient) AddAttachmentToPlan(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_plan/%d", planID)
	return c.uploadAttachment(ctx, endpoint, filePath)
}

// AddAttachmentToPlanEntry добавляет вложение к entry в плане
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttoplanentry
func (c *HTTPClient) AddAttachmentToPlanEntry(ctx context.Context, planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_plan_entry/%d/%s", planID, entryID)
	return c.uploadAttachment(ctx, endpoint, filePath)
}

// AddAttachmentToResult добавляет вложение к результату test
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttoresult
func (c *HTTPClient) AddAttachmentToResult(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_result/%d", resultID)
	return c.uploadAttachment(ctx, endpoint, filePath)
}

// AddAttachmentToRun добавляет вложение к тест-рану
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttorun
func (c *HTTPClient) AddAttachmentToRun(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_run/%d", runID)
	return c.uploadAttachment(ctx, endpoint, filePath)
}

// uploadAttachment универсальный метод загрузки файла
func (c *HTTPClient) uploadAttachment(ctx context.Context, endpoint, filePath string) (*data.AttachmentResponse, error) {
	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	// Создаем multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Добавляем файл в форму
	fileName := filepath.Base(filePath)
	part, _ := writer.CreateFormFile("attachment", fileName)
	_, _ = io.Copy(part, file)
	_ = writer.Close()

	// Используем DoRequest для выполнения запроса
	resp, err := c.DoRequest(ctx, "POST", endpoint, &requestBody, map[string]string{
		"Content-Type": writer.FormDataContentType(),
	})
	if err != nil {
		return nil, fmt.Errorf("error uploading attachment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	// Декодируем ответ
	var attachmentResp data.AttachmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachmentResp); err != nil {
		return nil, fmt.Errorf("error decoding attachment response: %w", err)
	}

	return &attachmentResp, nil
}
