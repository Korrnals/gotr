package sync

import (
	"context"
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
