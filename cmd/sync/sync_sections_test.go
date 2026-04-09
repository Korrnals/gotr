package sync

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/stretchr/testify/assert"
)

// resetSectionsFlags resets and recreates flags for sectionsCmd
func resetSectionsFlags() {
	sectionsCmd.ResetFlags()
	sectionsCmd.Flags().Int64("src-project", 0, "")
	sectionsCmd.Flags().Int64("src-suite", 0, "")
	sectionsCmd.Flags().Int64("dst-project", 0, "")
	sectionsCmd.Flags().Int64("dst-suite", 0, "")
	sectionsCmd.Flags().String("compare-field", "title", "")
	sectionsCmd.Flags().Bool("dry-run", false, "")
	sectionsCmd.Flags().Bool("approve", false, "")
	sectionsCmd.Flags().Bool("save-mapping", false, "")
}

// TestSyncSections_DryRun_NoAddSection verifies that in dry-run mode
// no real HTTP calls to AddSection are made.
func TestSyncSections_DryRun_NoAddSection(t *testing.T) {
	// Prepare a mock client that signals the existence of a section
	addCalled := false
	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "Sec 1"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
		AddSectionFunc: func(ctx context.Context, projectID int64, r *data.AddSectionRequest) (*data.Section, error) {
			addCalled = true
			return &data.Section{ID: 200, Name: r.Name}, nil
		},
	}

	// Override migration factory to use our mock client
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	// Prepare the command with flags and mock client
	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "true")

	// Execute the command and verify that AddSection was not called
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddSection should not be called in dry-run mode")
}

// TestSyncSections_Confirm_TriggersAddSection verifies that after interactive confirmation
// AddSection is called to create missing sections
func TestSyncSections_Confirm_TriggersAddSection(t *testing.T) {
	// Prepare mock client and track AddSection call
	addCalled := false
	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "Sec 1"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
		AddSectionFunc: func(ctx context.Context, projectID int64, r *data.AddSectionRequest) (*data.Section, error) {
			addCalled = true
			return &data.Section{ID: 200, Name: r.Name}, nil
		},
	}

	// Override migration factory with mock to avoid real network calls
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "false")

	p := interactive.NewMockPrompter().WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	// Execute the command and verify that AddSection was called
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled, "AddSection should be called after confirmation")
}

func TestSyncSections_NoFlags_NonInteractive_Error(t *testing.T) {
	addCalled := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		AddSectionFunc: func(ctx context.Context, projectID int64, r *data.AddSectionRequest) (*data.Section, error) {
			addCalled = true
			return &data.Section{ID: 200, Name: r.Name}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, addCalled, "AddSection should not be called in non-interactive")
}

func TestSyncSections_ConfirmDeclined_SkipsImport(t *testing.T) {
	addCalled := false
	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "Sec 1"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
		AddSectionFunc: func(ctx context.Context, projectID int64, r *data.AddSectionRequest) (*data.Section, error) {
			addCalled = true
			return &data.Section{ID: 200, Name: r.Name}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")

	p := interactive.NewMockPrompter().WithConfirmResponses(false)
	cmd.SetContext(interactive.WithPrompter(context.Background(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddSection should not be called when confirmation is declined")
}

func TestSyncSections_SaveMappingFlag_WritesMappingFile(t *testing.T) {
	addCalled := false
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "Sec 1"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
		AddSectionFunc: func(ctx context.Context, projectID int64, r *data.AddSectionRequest) (*data.Section, error) {
			addCalled = true
			return &data.Section{ID: 201, Name: r.Name}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("approve", "true")
	cmd.Flags().Set("save-mapping", "true")

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled)

	logsDir := filepath.Join(homeDir, ".gotr", "logs")
	files, globErr := filepath.Glob(filepath.Join(logsDir, "mapping_*.json"))
	assert.NoError(t, globErr)
	assert.NotEmpty(t, files, "expected a saved mapping file")
}

func TestSyncSections_SaveMappingPromptAccepted_WritesMappingFile(t *testing.T) {
	addCalled := false
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "Sec 1"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
		AddSectionFunc: func(ctx context.Context, projectID int64, r *data.AddSectionRequest) (*data.Section, error) {
			addCalled = true
			return &data.Section{ID: 301, Name: r.Name}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("approve", "true")

	p := interactive.NewMockPrompter().WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(context.Background(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled)

	logsDir := filepath.Join(homeDir, ".gotr", "logs")
	files, globErr := filepath.Glob(filepath.Join(logsDir, "mapping_*.json"))
	assert.NoError(t, globErr)
	assert.NotEmpty(t, files, "expected a saved mapping file after confirmation")
}

func TestSyncSections_SaveMappingPromptErrorIgnored_NoMappingFile(t *testing.T) {
	addCalled := false
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "Sec 1"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
		AddSectionFunc: func(ctx context.Context, projectID int64, r *data.AddSectionRequest) (*data.Section, error) {
			addCalled = true
			return &data.Section{ID: 302, Name: r.Name}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("approve", "true")
	cmd.SetContext(interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled)

	logsDir := filepath.Join(homeDir, ".gotr", "logs")
	files, globErr := filepath.Glob(filepath.Join(logsDir, "mapping_*.json"))
	assert.NoError(t, globErr)
	assert.Empty(t, files, "mapping file should not be saved on confirm error")
}

func TestSyncSections_NoFlags_InteractiveSuccess(t *testing.T) {
	addCalled := false
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}, {ID: 2, Name: "P2"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S10"}, {ID: 20, Name: "S20"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "Sec 1"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
		AddSectionFunc: func(ctx context.Context, projectID int64, r *data.AddSectionRequest) (*data.Section, error) {
			addCalled = true
			return &data.Section{ID: 202, Name: r.Name}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("approve", "true")

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 1}).
		WithSelectResponses(interactive.SelectResponse{Index: 1})
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled, "AddSection should be called after interactive selection")
}
