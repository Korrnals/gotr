// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для attachments get ====================

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentFunc: func(attachmentID int64) (*data.Attachment, error) {
			assert.Equal(t, int64(12345), attachmentID)
			return &data.Attachment{
				ID:          12345,
				Name:        "test-file.pdf",
				Filename:    "test-file.pdf",
				Size:        1024,
				ContentType: "application/pdf",
				CreatedOn:   1704067200,
				ProjectID:   1,
				CaseID:      100,
				UserID:      5,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "test-file.pdf")
	assert.Contains(t, output, "12345")
}

func TestGetCmd_WithJSONOutput(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentFunc: func(attachmentID int64) (*data.Attachment, error) {
			return &data.Attachment{
				ID:          12345,
				Name:        "test-file.pdf",
				Filename:    "test-file.pdf",
				Size:        1024,
				ContentType: "application/pdf",
				CreatedOn:   1704067200,
				ProjectID:   1,
				CaseID:      100,
				UserID:      5,
			}, nil
		},
	}

	// Create a temp directory for output file
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.json")

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "-o", outputFile})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify the file was created and contains the expected data
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "test-file.pdf")
	assert.Contains(t, string(content), "12345")
	assert.Contains(t, string(content), "application/pdf")
}

func TestGetCmd_InvalidAttachmentID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid attachment_id")
}

func TestGetCmd_ZeroAttachmentID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid attachment_id")
}

func TestGetCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentFunc: func(attachmentID int64) (*data.Attachment, error) {
			return nil, fmt.Errorf("attachment not found")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
