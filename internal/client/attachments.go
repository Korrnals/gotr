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

// DeleteAttachment deletes an attachment by ID.
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

// GetAttachment fetches attachment info by ID.
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

// GetAttachmentsForCase fetches attachments for a test case.
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

// GetAttachmentsForPlan fetches attachments for a test plan.
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

// GetAttachmentsForPlanEntry fetches attachments for a plan entry.
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

// GetAttachmentsForRun fetches attachments for a test run.
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

// GetAttachmentsForProject fetches attachments for a project.
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#getattachmentsforproject
func (c *HTTPClient) GetAttachmentsForProject(ctx context.Context, projectID int64) (data.GetAttachmentsResponse, error) {
	endpoint := fmt.Sprintf("get_attachments_for_project/%d", projectID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting attachments for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var attachments data.GetAttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("error decoding attachments for project %d: %w", projectID, err)
	}

	return attachments, nil
}

// GetAttachmentsForTest fetches attachments for a test.
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

// AddAttachmentToCase uploads an attachment to a test case.
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttocase
func (c *HTTPClient) AddAttachmentToCase(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_case/%d", caseID)
	return c.uploadAttachment(ctx, endpoint, filePath)
}

// AddAttachmentToPlan uploads an attachment to a test plan.
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttoplan
func (c *HTTPClient) AddAttachmentToPlan(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_plan/%d", planID)
	return c.uploadAttachment(ctx, endpoint, filePath)
}

// AddAttachmentToPlanEntry uploads an attachment to a plan entry.
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttoplanentry
func (c *HTTPClient) AddAttachmentToPlanEntry(ctx context.Context, planID int64, entryID, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_plan_entry/%d/%s", planID, entryID)
	return c.uploadAttachment(ctx, endpoint, filePath)
}

// AddAttachmentToResult uploads an attachment to a test result.
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttoresult
func (c *HTTPClient) AddAttachmentToResult(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_result/%d", resultID)
	return c.uploadAttachment(ctx, endpoint, filePath)
}

// AddAttachmentToRun uploads an attachment to a test run.
// https://support.testrail.com/hc/en-us/articles/7077990441108-Attachments#addattachmenttorun
func (c *HTTPClient) AddAttachmentToRun(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error) {
	endpoint := fmt.Sprintf("add_attachment_to_run/%d", runID)
	return c.uploadAttachment(ctx, endpoint, filePath)
}

// uploadAttachment is a generic file upload method.
func (c *HTTPClient) uploadAttachment(ctx context.Context, endpoint, filePath string) (*data.AttachmentResponse, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the file to the form
	fileName := filepath.Base(filePath)
	part, _ := writer.CreateFormFile("attachment", fileName)
	_, _ = io.Copy(part, file)
	_ = writer.Close()

	// Use DoRequest for the upload
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

	// Decode response
	var attachmentResp data.AttachmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachmentResp); err != nil {
		return nil, fmt.Errorf("error decoding attachment response: %w", err)
	}

	return &attachmentResp, nil
}
