package sync

import (
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"

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
// Ожидается, что в режиме dry-run не будет вызван метод AddSuite клиента.
func TestSyncSuites_DryRun_NoAddSuite(t *testing.T) {
	// Подготавливаем мок-клиент: source содержит одну suite
	addCalled := false
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
			}
			return data.GetSuitesResponse{}, nil
		},
		AddSuiteFunc: func(projectID int64, r *data.AddSuiteRequest) (*data.Suite, error) {
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
	// Подготавливаем мок-клиент и отмечаем факт вызова AddSuite
	addCalled := false
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
			}
			return data.GetSuitesResponse{}, nil
		},
		AddSuiteFunc: func(projectID int64, r *data.AddSuiteRequest) (*data.Suite, error) {
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

	// Симулируем ввод пользователя: подтверждение 'y'
	r, w, _ := os.Pipe()
	_, _ = w.Write([]byte("y\n"))
	_ = w.Close()
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	// Выполняем команду и проверяем, что AddSuite был вызван
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled, "AddSuite должен вызываться после подтверждения")
}
