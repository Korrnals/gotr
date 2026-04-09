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

// resetSuitesFlags resets and recreates flags for suitesCmd
func resetSuitesFlags() {
	suitesCmd.ResetFlags()
	suitesCmd.Flags().Int64("src-project", 0, "")
	suitesCmd.Flags().Int64("dst-project", 0, "")
	suitesCmd.Flags().String("compare-field", "title", "")
	suitesCmd.Flags().Bool("dry-run", false, "")
	suitesCmd.Flags().Bool("approve", false, "")
	suitesCmd.Flags().Bool("save-mapping", false, "")
}

// TestSyncSuites_DryRun_NoAddSuite verifies the command behavior in dry-run mode.
// In dry-run mode, the client's AddSuite method should not be called.
func TestSyncSuites_DryRun_NoAddSuite(t *testing.T) {
	// Prepare mock client: source contains one suite
	addCalled := false
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
			}
			return data.GetSuitesResponse{}, nil
		},
		AddSuiteFunc: func(ctx context.Context, projectID int64, r *data.AddSuiteRequest) (*data.Suite, error) {
			addCalled = true
			return &data.Suite{ID: 100, Name: r.Name}, nil
		},
	}

	// Override migration factory to use mock and temp directory for logs
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	// Prepare the command and set flags (dry-run = true)
	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dry-run", "true")

	// Execute the command and verify behavior
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddSuite should not be called in dry-run mode")
}

// TestSyncSuites_Confirm_TriggersAddSuite verifies that after interactive confirmation
// AddSuite is called to create the required suites in target.
func TestSyncSuites_Confirm_TriggersAddSuite(t *testing.T) {
	// Prepare mock client and track AddSuite call
	addCalled := false
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
			}
			return data.GetSuitesResponse{}, nil
		},
		AddSuiteFunc: func(ctx context.Context, projectID int64, r *data.AddSuiteRequest) (*data.Suite, error) {
			addCalled = true
			return &data.Suite{ID: 100, Name: r.Name}, nil
		},
	}

	// Override migration factory with mock to avoid real requests and logs
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dry-run", "false")

	p := interactive.NewMockPrompter().WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	// Execute the command and verify that AddSuite was called
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled, "AddSuite should be called after confirmation")
}

func TestSyncSuites_Confirm_NonInteractive_Error(t *testing.T) {
	addCalled := false
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
			}
			return data.GetSuitesResponse{}, nil
		},
		AddSuiteFunc: func(ctx context.Context, projectID int64, r *data.AddSuiteRequest) (*data.Suite, error) {
			addCalled = true
			return &data.Suite{ID: 100, Name: r.Name}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dry-run", "false")
	cmd.Flags().Set("approve", "false")
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, addCalled, "AddSuite should not be called in non-interactive")
}

func TestSyncSuites_RequiredIDs_ReturnsError(t *testing.T) {
	resetSuitesFlags()
	cmd := suitesCmd

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required IDs")
}

func TestSyncSuites_ConfirmDeclined_SkipsImport(t *testing.T) {
	addCalled := false
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
			}
			return data.GetSuitesResponse{}, nil
		},
		AddSuiteFunc: func(ctx context.Context, projectID int64, r *data.AddSuiteRequest) (*data.Suite, error) {
			addCalled = true
			return &data.Suite{ID: 100, Name: r.Name}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")

	p := interactive.NewMockPrompter().WithConfirmResponses(false)
	cmd.SetContext(interactive.WithPrompter(context.Background(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddSuite should not be called when confirmation is declined")
}

func TestSyncSuites_SaveMappingPromptAccepted_WritesMappingFile(t *testing.T) {
	addCalled := false
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
			}
			return data.GetSuitesResponse{}, nil
		},
		AddSuiteFunc: func(ctx context.Context, projectID int64, r *data.AddSuiteRequest) (*data.Suite, error) {
			addCalled = true
			return &data.Suite{ID: 100, Name: r.Name}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")
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
