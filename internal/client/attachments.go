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
