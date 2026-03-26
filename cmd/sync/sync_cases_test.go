package sync

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/stretchr/testify/assert"
)

// resetCasesFlags сбрасывает и пересоздаёт флаги для casesCmd
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

// TestSyncCases_DryRun_NoAddCase проверяет, что в режиме dry-run не вызывается AddCase
func TestSyncCases_DryRun_NoAddCase(t *testing.T) {
	addCalled := false

	// Создаём mock client который реализует оба интерфейса (client.ClientInterface и migration.ClientInterface)
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

	// Подменяем newMigration для теста
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	// Устанавливаем mock client через SetTestClient
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
	assert.False(t, addCalled, "AddCase не должен вызываться в dry-run")
}

// TestSyncCases_Confirm_TriggersAddCase проверяет, что подтверждение запускает импорт кейсов
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
	assert.True(t, addCalled, "AddCase должен вызываться после подтверждения")
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
	assert.False(t, addCalled, "AddCase не должен вызываться в non-interactive")
}
