package cases

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
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
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
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
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
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
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
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

func TestAddCmd_NoSection_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "Suite 10", ProjectID: 1}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 100, Name: "Section 100", SuiteID: 10}}, nil
		},
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(100), sectionID)
			assert.Equal(t, "Interactive Case", req.Title)
			return &data.Case{ID: 12345, Title: req.Title}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--title=Interactive Case"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_NoSection_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"--title=Interactive Case"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestAddCmd_NoSection_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--title=Case without section"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section_id required")
}

// ==================== Flag Tests ====================

func TestAddCmd_AllFlags(t *testing.T) {
	mock := &client.MockClient{
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
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
	err := os.WriteFile(jsonFile, []byte(jsonData), 0o644)
	assert.NoError(t, err)

	mock := &client.MockClient{
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
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
	err := os.WriteFile(jsonFile, []byte("not valid json"), 0o644)
	assert.NoError(t, err)

	mock := &client.MockClient{}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--json-file=" + jsonFile})

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing JSON")
}
