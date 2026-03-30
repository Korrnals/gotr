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

// resetSectionsFlags сбрасывает и пересоздаёт флаги для sectionsCmd
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

// TestSyncSections_DryRun_NoAddSection проверяет, что при режиме dry-run
// реальные HTTP-вызовы к AddSection не выполняются.
func TestSyncSections_DryRun_NoAddSection(t *testing.T) {
	// Подготавливаем мок-client, который сигнализирует о существовании секции
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

	// Переопределяем фабрику миграции, чтобы она использовала наш мок-client
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	// Подготавливаем команду с флагами и mock clientом
	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "true")

	// Выполняем команду и убеждаемся, что AddSection не вызван
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddSection не должен вызываться в режиме dry-run")
}

// TestSyncSections_Confirm_TriggersAddSection проверяет, что после интерактивного подтверждения
// выполняется вызов AddSection для создания отсутствующих секций
func TestSyncSections_Confirm_TriggersAddSection(t *testing.T) {
	// Подготавливаем мок-client и отслеживаем вызов AddSection
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

	// Переопределяем фабрику миграции на мок, чтобы избежать реальных сетевых вызовов
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

	// Выполняем команду и проверяем, что AddSection был вызван
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled, "AddSection должен вызываться после подтверждения")
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
	assert.False(t, addCalled, "AddSection не должен вызываться в non-interactive")
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
	assert.False(t, addCalled, "AddSection не должен вызываться при отказе подтверждения")
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
	assert.NotEmpty(t, files, "ожидается сохраненный mapping файл")
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
	assert.NotEmpty(t, files, "ожидается сохраненный mapping файл после подтверждения")
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
	assert.Empty(t, files, "mapping файл не должен сохраняться при ошибке confirm")
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
	assert.True(t, addCalled, "AddSection должен вызываться после интерактивного выбора")
}
