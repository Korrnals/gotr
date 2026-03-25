// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для attachments delete ====================

func TestDeleteCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		DeleteAttachmentFunc: func(ctx context.Context, attachmentID int64) error {
			assert.Equal(t, int64(12345), attachmentID)
			return nil
		},
	}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Attachment 12345 deleted successfully")
}

func TestDeleteCmd_InvalidAttachmentID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attachment_id")
}

func TestDeleteCmd_ZeroAttachmentID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attachment_id")
}

func TestDeleteCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		DeleteAttachmentFunc: func(ctx context.Context, attachmentID int64) error {
			return fmt.Errorf("attachment not found")
		},
	}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_NoArgs_Interactive(t *testing.T) {
	deleted := false
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
		DeleteAttachmentFunc: func(ctx context.Context, attachmentID int64) error {
			assert.Equal(t, int64(12345), attachmentID)
			deleted = true
			return nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, deleted)
}

func TestDeleteCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	deleted := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		DeleteAttachmentFunc: func(ctx context.Context, attachmentID int64) error {
			deleted = true
			return nil
		},
	}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, deleted)
}

func TestDeleteCmd_DryRun_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		DeleteAttachmentFunc: func(ctx context.Context, attachmentID int64) error {
			called = true
			return nil
		},
	}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.False(t, called)
}
