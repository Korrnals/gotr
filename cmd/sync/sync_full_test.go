package sync

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"

	"github.com/stretchr/testify/assert"
)

// resetFullFlags resets and recreates flags for fullCmd
func resetFullFlags() {
	fullCmd.ResetFlags()
	fullCmd.Flags().Int64("src-project", 0, "")
	fullCmd.Flags().Int64("src-suite", 0, "")
	fullCmd.Flags().Int64("dst-project", 0, "")
	fullCmd.Flags().Int64("dst-suite", 0, "")
	fullCmd.Flags().String("compare-field", "title", "")
	fullCmd.Flags().Bool("dry-run", false, "")
	fullCmd.Flags().Bool("approve", false, "")
	fullCmd.Flags().Bool("save-mapping", false, "")
}

// TestSyncFull_DryRun_NoAdds verifies that dry-run does not trigger entity creation
func TestSyncFull_DryRun_NoAdds(t *testing.T) {
	addShared := false
	addCase := false

	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			if projectID == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "A"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		AddSharedStepFunc: func(ctx context.Context, projectID int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addShared = true
			return &data.SharedStep{ID: 100}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCase = true
			return &data.Case{ID: 100}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetFullFlags()
	cmd := fullCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "true")

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addShared, "AddSharedStep should not be called in dry-run")
	assert.False(t, addCase, "AddCase should not be called in dry-run")
}

// TestSyncFull_AutoApprove_PerformsMigration verifies that auto-approve triggers the full process
func TestSyncFull_AutoApprove_PerformsMigration(t *testing.T) {
	addShared := false
	addCase := false

	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			if projectID == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "A"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		AddSharedStepFunc: func(ctx context.Context, projectID int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addShared = true
			return &data.SharedStep{ID: 100}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 && suiteID == 10 {
				return data.GetCasesResponse{{ID: 1, Title: "Case1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCase = true
			return &data.Case{ID: 100}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetFullFlags()
	cmd := fullCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "false")
	cmd.Flags().Set("approve", "true")
	cmd.Flags().Set("save-mapping", "true")

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addShared, "AddSharedStep should be called on auto-approve")
	assert.True(t, addCase, "AddCase should be called on auto-approve")
}

func TestSyncFull_NoFlags_NonInteractive_Error(t *testing.T) {
	addShared := false
	addCase := false

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		AddSharedStepFunc: func(ctx context.Context, projectID int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addShared = true
			return &data.SharedStep{ID: 100}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCase = true
			return &data.Case{ID: 100}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetFullFlags()
	cmd := fullCmd
	SetTestClient(cmd, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, addShared, "AddSharedStep should not be called in non-interactive")
	assert.False(t, addCase, "AddCase should not be called in non-interactive")
}

func TestSyncFull_NoFlags_InteractiveSuccess(t *testing.T) {
	addShared := false
	addCase := false

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}, {ID: 2, Name: "P2"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S10"}, {ID: 20, Name: "S20"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			if projectID == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "A"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		AddSharedStepFunc: func(ctx context.Context, projectID int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addShared = true
			return &data.SharedStep{ID: 100}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 && suiteID == 10 {
				return data.GetCasesResponse{{ID: 1, Title: "Case1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(ctx context.Context, suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCase = true
			return &data.Case{ID: 100}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetFullFlags()
	cmd := fullCmd
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
	assert.True(t, addShared, "AddSharedStep should be called after interactive selection")
	assert.True(t, addCase, "AddCase should be called after interactive selection")
}
