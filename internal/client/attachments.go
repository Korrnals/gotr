// internal/client/attachments.go
package client

import (
	"bytes"
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
func (c *HTTPClient) DeleteAttachment(attachmentID int64) error {
	endpoint := fmt.Sprintf("delete_attachment/%d", attachmentID)
	resp, err := c.Post(endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting attachment %d: %w", attachmentID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s when deleting attachment %d: %s", resp.Status, attachmentID, string(body))
	}

	return nil
}

// GetAttachment получает информацию о вложении по ID
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachment
func (c *HTTPClient) GetAttachment(attachmentID int64) (*data.Attachment, error) {
	endpoint := fmt.Sprintf("get_attachment/%d", attachmentID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachment %d: %w", attachmentID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s when getting attachment %d: %s", resp.Status, attachmentID, string(body))
	}

	var attachment data.Attachment
	if err := json.NewDecoder(resp.Body).Decode(&attachment); err != nil {
		return nil, fmt.Errorf("error decoding attachment %d: %w", attachmentID, err)
	}

	return &attachment, nil
}

// GetAttachmentsForCase получает список вложений для тест-кейса
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsforcase
func (c *HTTPClient) GetAttachmentsForCase(caseID int64) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_case/%d", caseID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for case %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s when getting attachments for case %d: %s", resp.Status, caseID, string(body))
	}

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for case %d: %w", caseID, err)
	}

	return attachments, nil
}

// GetAttachmentsForPlan получает список вложений для тест-плана
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsforplan
func (c *HTTPClient) GetAttachmentsForPlan(planID int64) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_plan/%d", planID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for plan %d: %w", planID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s when getting attachments for plan %d: %s", resp.Status, planID, string(body))
	}

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for plan %d: %w", planID, err)
	}

	return attachments, nil
}

// GetAttachmentsForPlanEntry получает список вложений для записи в тест-плане
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsforplanentry
func (c *HTTPClient) GetAttachmentsForPlanEntry(planID int64, entryID string) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_plan_entry/%d/%s", planID, entryID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for plan entry %s in plan %d: %w", entryID, planID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s when getting attachments for plan entry %s: %s", resp.Status, entryID, string(body))
	}

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for plan entry %s: %w", entryID, err)
	}

	return attachments, nil
}

// GetAttachmentsForRun получает список вложений для тест-рана
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsforrun
func (c *HTTPClient) GetAttachmentsForRun(runID int64) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_run/%d", runID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for run %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s when getting attachments for run %d: %s", resp.Status, runID, string(body))
	}

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for run %d: %w", runID, err)
	}

	return attachments, nil
}

// GetAttachmentsForTest получает список вложений для теста
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsfortest
func (c *HTTPClient) GetAttachmentsForTest(testID int64) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_test/%d", testID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for test %d: %w", testID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s when getting attachments for test %d: %s", resp.Status, testID, string(body))
	}

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for test %d: %w", testID, err)
	}

	return attachments, nil
}

// AddAttachmentToCase добавляет вложение к тест-кейсу
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttocase
func (c *HTTPClient) AddAttachmentToCase(caseID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_case/%d", caseID)
	return c.uploadAttachment(endpoint, filePath)
}

// AddAttachmentToPlan добавляет вложение к тест-плану
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttoplan
func (c *HTTPClient) AddAttachmentToPlan(planID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_plan/%d", planID)
	return c.uploadAttachment(endpoint, filePath)
}

// AddAttachmentToPlanEntry добавляет вложение к entry в плане
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttoplanentry
func (c *HTTPClient) AddAttachmentToPlanEntry(planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_plan_entry/%d/%s", planID, entryID)
	return c.uploadAttachment(endpoint, filePath)
}

// AddAttachmentToResult добавляет вложение к результату теста
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttoresult
func (c *HTTPClient) AddAttachmentToResult(resultID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_result/%d", resultID)
	return c.uploadAttachment(endpoint, filePath)
}

// AddAttachmentToRun добавляет вложение к тест-рану
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttorun
func (c *HTTPClient) AddAttachmentToRun(runID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_run/%d", runID)
	return c.uploadAttachment(endpoint, filePath)
}

// uploadAttachment универсальный метод загрузки файла
func (c *HTTPClient) uploadAttachment(endpoint, filePath string) (*data.AttachmentResponse, error) {
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
	part, err := writer.CreateFormFile("attachment", fileName)
	if err != nil {
		return nil, fmt.Errorf("error creating form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("error copying file to form: %w", err)
	}

	// Закрываем writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing multipart writer: %w", err)
	}

	// Используем DoRequest для выполнения запроса
	resp, err := c.DoRequest("POST", endpoint, &requestBody, map[string]string{
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
