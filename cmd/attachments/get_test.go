// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Tests for attachments get ====================

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentFunc: func(ctx context.Context, attachmentID int64) (*data.Attachment, error) {
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

func TestGetCmd_WithSaveFlag(t *testing.T) {
	// Create temp home directory for exports
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	mock := &client.MockClient{
		GetAttachmentFunc: func(ctx context.Context, attachmentID int64) (*data.Attachment, error) {
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
	cmd.SetArgs([]string{"12345", "--save"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify file was created in exports directory
	exportsDir := filepath.Join(tempHome, ".gotr", "exports", "attachments")
	files, err := os.ReadDir(exportsDir)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Verify content
	content, err := os.ReadFile(filepath.Join(exportsDir, files[0].Name()))
	require.NoError(t, err)
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
	assert.Contains(t, err.Error(), "attachment_id")
}

func TestGetCmd_ZeroAttachmentID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attachment_id")
}

func TestGetCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetAttachmentFunc: func(ctx context.Context, attachmentID int64) (*data.Attachment, error) {
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

func TestGetCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "Suite 10", ProjectID: 1}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: 100, Title: "Case 100"}}, nil
		},
		GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{{ID: 12345, Name: "test-file.pdf", Size: 1024, CreatedOn: 1704067200}}, nil
		},
		GetAttachmentFunc: func(ctx context.Context, attachmentID int64) (*data.Attachment, error) {
			assert.Equal(t, int64(12345), attachmentID)
			return &data.Attachment{ID: 12345, Name: "test-file.pdf", Size: 1024, ContentType: "application/pdf"}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "test-file.pdf")
}

func TestGetCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
