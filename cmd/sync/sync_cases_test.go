package sync

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/service/migration"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/stretchr/testify/assert"
)

// resetCasesFlags resets and recreates flags for casesCmd
func resetCasesFlags() {
	casesCmd.ResetFlags()
	casesCmd.Flags().Int64("src-project", 0, "")
	casesCmd.Flags().Int64("src-suite", 0, "")
	casesCmd.Flags().Int64("dst-project", 0, "")
	casesCmd.Flags().Int64("dst-suite", 0, "")
	casesCmd.Flags().String("compare-field", "title", "")
	casesCmd.Flags().Bool("dry-run", false, "")
	casesCmd.Flags().String("output", "", "")
	casesCmd.Flags().String("mapping-file", "", "")
}

// TestSyncCases_DryRun_NoAddCase verifies that AddCase is not called in dry-run mode
func TestSyncCases_DryRun_NoAddCase(t *testing.T) {
	addCalled := false

	// Create a mock client that implements both interfaces (client.ClientInterface and migration.ClientInterface)
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	// Override newMigration for test
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	// Set mock client via SetTestClient
	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "true")

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddCase should not be called in dry-run")
}

// TestSyncCases_Confirm_TriggersAddCase verifies that confirmation triggers case import
func TestSyncCases_Confirm_TriggersAddCase(t *testing.T) {
	addCalled := false

	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "false")

	p := interactive.NewMockPrompter().WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled, "AddCase should be called after confirmation")
}

func TestSyncCases_NoFlags_NonInteractive_Error(t *testing.T) {
	addCalled := false

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, addCalled, "AddCase should not be called in non-interactive")
}

func TestSyncCases_InvalidMappingFile_ReturnsError(t *testing.T) {
	addCalled := false
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("mapping-file", "/tmp/does-not-exist-mapping.json")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load mapping")
	assert.False(t, addCalled, "AddCase should not be called on mapping error")
}

func TestSyncCases_ConfirmDeclined_SkipsImportAndWritesLog(t *testing.T) {
	addCalled := false
	outputFile := t.TempDir() + "/sync_cases_log.json"

	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("output", outputFile)

	p := interactive.NewMockPrompter().WithConfirmResponses(false)
	cmd.SetContext(interactive.WithPrompter(context.Background(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddCase should not be called when confirmation is declined")

	data, readErr := os.ReadFile(outputFile)
	assert.NoError(t, readErr)
	assert.Contains(t, string(data), "\"filtered\": 1")
}

func TestSyncCases_ConfirmInNonInteractiveMode_ReturnsError(t *testing.T) {
	addCalled := false
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.SetContext(interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()))
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, addCalled, "AddCase should not be called on confirmation error")
}

func TestSaveLog_WritesStructuredPayload(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "cases_log.json")

	matches := data.GetCasesResponse{{ID: 1, Title: "match"}}
	filtered := data.GetCasesResponse{{ID: 2, Title: "new"}, {ID: 3, Title: "new-2"}}
	mapping := map[int64]int64{10: 20, 30: 40}
	errorsList := []string{"import failed for case 3"}

	saveLog(logFile, matches, filtered, errorsList, mapping, true)

	raw, err := os.ReadFile(logFile)
	assert.NoError(t, err)

	var payload map[string]interface{}
	err = json.Unmarshal(raw, &payload)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), payload["matches"])
	assert.Equal(t, float64(2), payload["filtered"])

	mappingPayload, ok := payload["mapping"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(20), mappingPayload["10"])
	assert.Equal(t, float64(40), mappingPayload["30"])

	errorsPayload, ok := payload["errors"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, errorsPayload, 1)
	assert.Equal(t, "import failed for case 3", errorsPayload[0])
}

func TestSyncCases_DryRun_WithMappingFile_WritesLogWithLoadedMapping(t *testing.T) {
	addCalled := false
	tmpDir := t.TempDir()
	mappingFile := filepath.Join(tmpDir, "mapping.json")
	outputFile := filepath.Join(tmpDir, "sync_cases_log.json")

	err := os.WriteFile(mappingFile, []byte(`{"101":202}`), 0o600)
	assert.NoError(t, err)

	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("mapping-file", mappingFile)
	cmd.Flags().Set("dry-run", "true")
	cmd.Flags().Set("output", outputFile)

	err = cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddCase should not be called in dry-run")

	raw, readErr := os.ReadFile(outputFile)
	assert.NoError(t, readErr)

	var payload map[string]interface{}
	err = json.Unmarshal(raw, &payload)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), payload["filtered"])

	mappingPayload, ok := payload["mapping"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(202), mappingPayload["101"])
}

func TestSyncCases_InvalidMappingJSON_ReturnsError(t *testing.T) {
	addCalled := false
	tmpDir := t.TempDir()
	mappingFile := filepath.Join(tmpDir, "invalid_mapping.json")
	err := os.WriteFile(mappingFile, []byte("{bad-json"), 0o600)
	assert.NoError(t, err)

	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("mapping-file", mappingFile)

	err = cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load mapping")
	assert.False(t, addCalled, "AddCase should not be called on invalid mapping")
}

func TestSyncCases_NewMigrationFactoryError_ReturnsError(t *testing.T) {
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = func(cli client.ClientInterface, srcProject, srcSuite, dstProject, dstSuite int64, compareField, logDir string) (*migration.Migration, error) {
		return nil, assert.AnError
	}

	resetCasesFlags()
	cmd := casesCmd
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

func TestSyncCases_ImportWithErrors_WritesErrorsToLog(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "sync_cases_errors.json")

	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			return nil, assert.AnError
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("output", outputFile)

	p := interactive.NewMockPrompter().WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(context.Background(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	raw, readErr := os.ReadFile(outputFile)
	assert.NoError(t, readErr)
	assert.Contains(t, string(raw), "\"errors\": [")
	assert.Contains(t, string(raw), "assert.AnError")
}

func TestSyncCases_DryRun_DefaultOutputPath_WritesLogFile(t *testing.T) {
	addCalled := false
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "true")

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddCase should not be called in dry-run")

	logsDir := filepath.Join(homeDir, ".gotr", "logs")
	files, globErr := filepath.Glob(filepath.Join(logsDir, "sync_cases_*.json"))
	assert.NoError(t, globErr)
	assert.NotEmpty(t, files, "expected a default log file")
}

func TestSyncCases_NoFlags_InteractiveSelection_DeclineConfirm(t *testing.T) {
	addCalled := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}, {ID: 2, Name: "P2"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S10"}, {ID: 20, Name: "S20"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 && suiteID == 10 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 1}).
		WithSelectResponses(interactive.SelectResponse{Index: 1}).
		WithConfirmResponses(false)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddCase should not be called when confirmation is declined")
}