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

// resetSuitesFlags сбрасывает и пересоздаёт флаги для suitesCmd
func resetSuitesFlags() {
	suitesCmd.ResetFlags()
	suitesCmd.Flags().Int64("src-project", 0, "")
	suitesCmd.Flags().Int64("dst-project", 0, "")
	suitesCmd.Flags().String("compare-field", "title", "")
	suitesCmd.Flags().Bool("dry-run", false, "")
	suitesCmd.Flags().Bool("approve", false, "")
	suitesCmd.Flags().Bool("save-mapping", false, "")
}

// TestSyncSuites_DryRun_NoAddSuite проверяет поведение команды при режиме dry-run.
// Ожидается, что в режиме dry-run не будет вызван метод AddSuite clientа.
func TestSyncSuites_DryRun_NoAddSuite(t *testing.T) {
	// Подготавливаем мок-client: source содержит одну suite
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

	// Переопределяем фабрику миграции, чтобы использовать мок и временную директорию для логов
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	// Готовим команду и устанавливаем флаги (dry-run = true)
	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dry-run", "true")

	// Выполняем команду и проверяем поведение
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddSuite не должен вызываться в режиме dry-run")
}

// TestSyncSuites_Confirm_TriggersAddSuite проверяет, что после интерактивного подтверждения
// выполняется вызов AddSuite для создания необходимых suites в target.
func TestSyncSuites_Confirm_TriggersAddSuite(t *testing.T) {
	// Подготавливаем мок-client и отмечаем факт вызова AddSuite
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

	// Переопределяем фабрику миграции на мок, чтобы избежать реальных запросов и логов
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

	// Выполняем команду и проверяем, что AddSuite был вызван
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled, "AddSuite должен вызываться после подтверждения")
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
	assert.False(t, addCalled, "AddSuite не должен вызываться в non-interactive")
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
	assert.False(t, addCalled, "AddSuite не должен вызываться при отказе подтверждения")
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
	assert.NotEmpty(t, files, "ожидается сохраненный mapping файл после подтверждения")
}
