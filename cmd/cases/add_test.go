package cases

import (
	"fmt"
	"os"
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

func TestAddCmd_WithJSONFile(t *testing.T) {
	// Create a temp JSON file
	tempDir := t.TempDir()
	jsonFile := tempDir + "/case.json"
	jsonData := `{"title": "JSON Test Case", "template_id": 2, "priority_id": 1}`
	err := os.WriteFile(jsonFile, []byte(jsonData), 0644)
	assert.NoError(t, err)

	mock := &client.MockClient{
		AddCaseFunc: func(sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(100), sectionID)
			assert.Equal(t, "JSON Test Case", req.Title)
			assert.Equal(t, int64(2), req.TemplateID)
			assert.Equal(t, int64(1), req.PriorityID)
			return &data.Case{ID: 555, Title: req.Title}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--json-file=" + jsonFile})

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_JSONFileNotFound(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--json-file=/nonexistent/file.json"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading JSON file")
}

func TestAddCmd_InvalidJSONFile(t *testing.T) {
	// Create a temp file with invalid JSON
	tempDir := t.TempDir()
	jsonFile := tempDir + "/invalid.json"
	err := os.WriteFile(jsonFile, []byte("not valid json"), 0644)
	assert.NoError(t, err)

	mock := &client.MockClient{}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--json-file=" + jsonFile})

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing JSON")
}
