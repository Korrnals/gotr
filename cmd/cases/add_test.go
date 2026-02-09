package cases

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Dry Run Tests ====================

func TestAddCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--title=Test Case", "--template-id=1", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_DryRun_NoTitle(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--dry-run"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title")
}

// ==================== Functional Tests with Mock ====================

func TestAddCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddCaseFunc: func(sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(100), sectionID)
			assert.Equal(t, "Test Case Title", req.Title)
			assert.Equal(t, int64(1), req.TemplateID)
			assert.Equal(t, int64(2), req.TypeID)
			return &data.Case{ID: 12345, Title: req.Title}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--title=Test Case Title", "--template-id=1", "--type-id=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_Success_Minimal(t *testing.T) {
	mock := &client.MockClient{
		AddCaseFunc: func(sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(50), sectionID)
			assert.Equal(t, "Minimal Case", req.Title)
			return &data.Case{ID: 999, Title: req.Title}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"50", "--title=Minimal Case"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		AddCaseFunc: func(sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			return nil, fmt.Errorf("API error: section not found")
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--title=Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

// ==================== Validation Tests ====================

func TestAddCmd_InvalidSectionID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--title=Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid section_id")
}

func TestAddCmd_ZeroSectionID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "--title=Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid section_id")
}

func TestAddCmd_MissingTitle(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title is required")
}

// ==================== Flag Tests ====================

func TestAddCmd_AllFlags(t *testing.T) {
	mock := &client.MockClient{
		AddCaseFunc: func(sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(100), sectionID)
			assert.Equal(t, "Full Test Case", req.Title)
			assert.Equal(t, int64(1), req.TemplateID)
			assert.Equal(t, int64(2), req.TypeID)
			assert.Equal(t, int64(3), req.PriorityID)
			assert.Equal(t, "JIRA-123", req.Refs)
			return &data.Case{ID: 11111, Title: req.Title}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{
		"100",
		"--title=Full Test Case",
		"--template-id=1",
		"--type-id=2",
		"--priority-id=3",
		"--refs=JIRA-123",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}
