package sync

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"

	"github.com/stretchr/testify/assert"
)

// resetFullFlags сбрасывает и пересоздаёт флаги для fullCmd
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

// TestSyncFull_DryRun_NoAdds проверяет, что dry-run не вызывает создания сущностей
func TestSyncFull_DryRun_NoAdds(t *testing.T) {
	addShared := false
	addCase := false

	mock := &client.MockClient{
		GetSharedStepsFunc: func(projectID int64) (data.GetSharedStepsResponse, error) {
			if projectID == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "A"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		AddSharedStepFunc: func(projectID int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addShared = true
			return &data.SharedStep{ID: 100}, nil
		},
		GetCasesFunc: func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
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
	assert.False(t, addShared, "AddSharedStep не должен вызываться в dry-run")
	assert.False(t, addCase, "AddCase не должен вызываться в dry-run")
}

// TestSyncFull_AutoApprove_PerformsMigration проверяет, что при авто-подтверждении запускается полный процесс
func TestSyncFull_AutoApprove_PerformsMigration(t *testing.T) {
	addShared := false
	addCase := false

	mock := &client.MockClient{
		GetSharedStepsFunc: func(projectID int64) (data.GetSharedStepsResponse, error) {
			if projectID == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "A"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		AddSharedStepFunc: func(projectID int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addShared = true
			return &data.SharedStep{ID: 100}, nil
		},
		GetCasesFunc: func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 && suiteID == 10 {
				return data.GetCasesResponse{{ID: 1, Title: "Case1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		AddCaseFunc: func(suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
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
	assert.True(t, addShared, "AddSharedStep должен вызываться при авто-подтверждении")
	assert.True(t, addCase, "AddCase должен вызываться при авто-подтверждении")
}
