package sync

import (
	"context"
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/service/migration"
	"github.com/Korrnals/gotr/internal/models/data"

	"github.com/stretchr/testify/assert"
)

// Используем общий mockClient из sync_common_test.go — не дублируем реализацию

// TestSyncSections_DryRun_NoAddSection проверяет, что при режиме dry-run
// реальные HTTP-вызовы к AddSection не выполняются.
func TestSyncSections_DryRun_NoAddSection(t *testing.T) {
	// Подготавливаем мок-клиент, который сигнализирует о существовании секции
	addCalled := false
	mock := &mockClient{
		getSections: func(p, s int64) (data.GetSectionsResponse, error) {
			if p == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "Sec 1"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
		addSection: func(p int64, r *data.AddSectionRequest) (*data.Section, error) {
			addCalled = true
			return &data.Section{ID: 200, Name: r.Name}, nil
		},
	}

	// Переопределяем фабрику миграции, чтобы она использовала наш мок-клиент
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = func(client migration.ClientInterface, srcProject, srcSuite, dstProject, dstSuite int64, compareField, logDir string) (*migration.Migration, error) {
		tmp := t.TempDir()
		return migration.NewMigration(mock, srcProject, srcSuite, dstProject, dstSuite, compareField, tmp)
	}

	// Подготавливаем команду с флагами и dummy-клиентом в контексте
	cmd := sectionsCmd
	cmd.SetContext(context.WithValue(context.Background(), testHTTPClientKey, &client.HTTPClient{}))
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
	// Подготавливаем мок-клиент и отслеживаем вызов AddSection
	addCalled := false
	mock := &mockClient{
		getSections: func(p, s int64) (data.GetSectionsResponse, error) {
			if p == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "Sec 1"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
		addSection: func(p int64, r *data.AddSectionRequest) (*data.Section, error) {
			addCalled = true
			return &data.Section{ID: 200, Name: r.Name}, nil
		},
	}

	// Переопределяем фабрику миграции на мок, чтобы избежать реальных сетевых вызовов
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = func(client migration.ClientInterface, srcProject, srcSuite, dstProject, dstSuite int64, compareField, logDir string) (*migration.Migration, error) {
		tmp := t.TempDir()
		return migration.NewMigration(mock, srcProject, srcSuite, dstProject, dstSuite, compareField, tmp)
	}

	// Устанавливаем dummy-клиент в контекст команды (чтобы GetClient не завершал процесс)
	dummy, _ := client.NewClient("http://example.com", "u", "k", false)
	cmd := sectionsCmd
	cmd.SetContext(context.WithValue(context.Background(), testHTTPClientKey, dummy))
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "false")

	// Симулируем ввод пользователя: подтверждение 'y'
	r, w, _ := os.Pipe()
	_, _ = w.Write([]byte("y\n"))
	_ = w.Close()
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	// Выполняем команду и проверяем, что AddSection был вызван
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled, "AddSection должен вызываться после подтверждения")
}
