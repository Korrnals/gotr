package sync

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/stretchr/testify/assert"
)

// resetSharedStepsFlags сбрасывает и пересоздаёт флаги для sharedStepsCmd
func resetSharedStepsFlags() {
	sharedStepsCmd.ResetFlags()
	sharedStepsCmd.Flags().Int64("src-project", 0, "")
	sharedStepsCmd.Flags().Int64("src-suite", 0, "")
	sharedStepsCmd.Flags().Int64("dst-project", 0, "")
	sharedStepsCmd.Flags().String("compare-field", "title", "")
	sharedStepsCmd.Flags().Bool("dry-run", false, "")
	sharedStepsCmd.Flags().Bool("approve", false, "")
	sharedStepsCmd.Flags().Bool("save-mapping", false, "")
	sharedStepsCmd.Flags().Bool("save-filtered", false, "")
	sharedStepsCmd.Flags().String("output", "", "")
}

// TestSyncSharedSteps_DryRun_NoAddSharedSteps проверяет, что dry-run не вызовет AddSharedStep
func TestSyncSharedSteps_DryRun_NoAddSharedSteps(t *testing.T) {
	addCalled := false

	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			if projectID == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "A"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
			}
			return data.GetSuitesResponse{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
		AddSharedStepFunc: func(ctx context.Context, projectID int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addCalled = true
			return &data.SharedStep{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10") // Явно указываем suite, чтобы избежать интерактивного выбора
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dry-run", "true")

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddSharedStep не должен вызываться в dry-run")
}

// TestSyncSharedSteps_Confirm_TriggersAddSharedStep проверяет, что подтверждение запускает импорт shared steps
func TestSyncSharedSteps_Confirm_TriggersAddSharedStep(t *testing.T) {
	addCalled := false

	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			if projectID == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "A"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
		AddSharedStepFunc: func(ctx context.Context, projectID int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addCalled = true
			return &data.SharedStep{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10") // Явно указываем suite, чтобы избежать интерактивного выбора
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dry-run", "false")

	p := interactive.NewMockPrompter().WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled, "AddSharedStep должен вызываться после подтверждения")
}

func TestSyncSharedSteps_NoFlags_NonInteractive_Error(t *testing.T) {
	addCalled := false

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		AddSharedStepFunc: func(ctx context.Context, projectID int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addCalled = true
			return &data.SharedStep{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, addCalled, "AddSharedStep не должен вызываться в non-interactive")
}
