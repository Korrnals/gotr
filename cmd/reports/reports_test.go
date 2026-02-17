package reports

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для Register ====================

func TestRegister(t *testing.T) {
	root := &cobra.Command{}
	mock := &client.MockClient{}

	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	// Проверяем, что команда reports добавлена
	reportsCmd, _, err := root.Find([]string{"reports"})
	assert.NoError(t, err)
	assert.NotNil(t, reportsCmd)

	// Проверяем наличие подкоманд
	listCmd, _, _ := root.Find([]string{"reports", "list"})
	assert.NotNil(t, listCmd)

	listCrossCmd, _, _ := root.Find([]string{"reports", "list-cross-project"})
	assert.NotNil(t, listCrossCmd)

	runCmd, _, _ := root.Find([]string{"reports", "run"})
	assert.NotNil(t, runCmd)

	runCrossCmd, _, _ := root.Find([]string{"reports", "run-cross-project"})
	assert.NotNil(t, runCrossCmd)
}

func TestRegister_Help(t *testing.T) {
	root := &cobra.Command{}
	mock := &client.MockClient{}

	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	// Проверяем, что вызов без аргументов показывает help
	root.SetArgs([]string{"reports"})
	err := root.Execute()
	assert.NoError(t, err)
}
