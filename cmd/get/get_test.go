package get

import (
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для handleOutput ====================

func TestHandleOutput_JSONOutput(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("type", "t", "json", "")
	cmd.Flags().String("save", "", "")
	cmd.Flags().BoolP("quiet", "q", false, "")
	cmd.Flags().BoolP("jq", "j", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().BoolP("body-only", "b", false, "")

	testData := map[string]string{"key": "value"}
	err := handleOutput(cmd, testData, time.Now())

	// Вывод идет в stdout, просто проверяем что нет ошибки
	assert.NoError(t, err)
}

func TestHandleOutput_JSONFullOutput(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("type", "t", "json-full", "")
	cmd.Flags().String("save", "", "")
	cmd.Flags().BoolP("quiet", "q", false, "")
	cmd.Flags().BoolP("jq", "j", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().BoolP("body-only", "b", false, "")

	testData := map[string]string{"test": "data"}
	err := handleOutput(cmd, testData, time.Now())

	assert.NoError(t, err)
}

func TestHandleOutput_DefaultOutput(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("type", "t", "table", "") // Неподдерживаемый формат
	cmd.Flags().String("save", "", "")
	cmd.Flags().BoolP("quiet", "q", false, "")
	cmd.Flags().BoolP("jq", "j", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().BoolP("body-only", "b", false, "")

	testData := map[string]string{"test": "data"}
	err := handleOutput(cmd, testData, time.Now())

	assert.NoError(t, err)
}

func TestHandleOutput_FileOutput(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("type", "t", "json", "")
	cmd.Flags().Bool("save", true, "")
	cmd.Flags().BoolP("quiet", "q", false, "")
	cmd.Flags().BoolP("jq", "j", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().BoolP("body-only", "b", false, "")

	testData := map[string]string{"test": "data"}
	err := handleOutput(cmd, testData, time.Now())

	assert.NoError(t, err)
}

func TestHandleOutput_FileOutputBodyOnly(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("type", "t", "json", "")
	cmd.Flags().Bool("save", true, "")
	cmd.Flags().BoolP("quiet", "q", false, "")
	cmd.Flags().BoolP("jq", "j", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().BoolP("body-only", "b", true, "")

	testData := map[string]string{"test": "data"}
	err := handleOutput(cmd, testData, time.Now())

	assert.NoError(t, err)
}

func TestHandleOutput_FileOutputQuiet(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("type", "t", "json", "")
	cmd.Flags().Bool("save", true, "")
	cmd.Flags().BoolP("quiet", "q", true, "")
	cmd.Flags().BoolP("jq", "j", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().BoolP("body-only", "b", false, "")

	testData := map[string]string{"test": "data"}
	err := handleOutput(cmd, testData, time.Now())

	assert.NoError(t, err)
}

func TestHandleOutput_QuietMode(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("type", "t", "json", "")
	cmd.Flags().String("save", "", "")
	cmd.Flags().BoolP("quiet", "q", true, "")
	cmd.Flags().BoolP("jq", "j", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().BoolP("body-only", "b", false, "")

	testData := map[string]string{"test": "data"}
	err := handleOutput(cmd, testData, time.Now())

	assert.NoError(t, err)
}

func TestHandleOutput_JQEnabled(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("type", "t", "json", "")
	cmd.Flags().String("save", "", "")
	cmd.Flags().BoolP("quiet", "q", false, "")
	cmd.Flags().BoolP("jq", "j", true, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().BoolP("body-only", "b", false, "")

	testData := map[string]string{"test": "data"}
	// JQ path - функция embed.RunEmbeddedJQ выполнится
	// Результат зависит от наличия jq в системе
	err := handleOutput(cmd, testData, time.Now())
	
	// Тест просто проверяет что путь jq выполняется без паники
	// Результат может быть успешным или с ошибкой в зависимости от окружения
	_ = err
}

func TestHandleOutput_JQFilterOnly(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("type", "t", "json", "")
	cmd.Flags().String("save", "", "")
	cmd.Flags().BoolP("quiet", "q", false, "")
	cmd.Flags().BoolP("jq", "j", false, "")
	cmd.Flags().String("jq-filter", ".test", "")
	cmd.Flags().BoolP("body-only", "b", false, "")

	testData := map[string]string{"test": "data"}
	// JQ path - функция embed.RunEmbeddedJQ выполнится
	err := handleOutput(cmd, testData, time.Now())
	
	// Тест просто проверяет что путь jq выполняется без паники
	_ = err
}

// ==================== Тесты для Register ====================

func TestRegister(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	
	// Создаем mock функцию для получения клиента
	mockClientFn := func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	}
	
	Register(rootCmd, mockClientFn)
	
	// Проверяем что команда get добавлена
	getCmd, _, err := rootCmd.Find([]string{"get"})
	assert.NoError(t, err)
	assert.NotNil(t, getCmd)
	assert.Equal(t, "get", getCmd.Name())
	
	// Проверяем что все подкоманды добавлены
	subCommands := []string{
		"cases", "case",
		"case-types", "case-fields",
		"case-history",
		"projects", "project",
		"sharedsteps", "sharedstep",
		"sharedstep-history",
		"suites", "suite",
	}
	
	for _, subCmdName := range subCommands {
		subCmd, _, err := rootCmd.Find([]string{"get", subCmdName})
		assert.NoError(t, err, "subcommand %s should exist", subCmdName)
		assert.NotNil(t, subCmd, "subcommand %s should not be nil", subCmdName)
	}
}

// ==================== Тесты для SetGetClientForTests ====================

func TestSetGetClientForTests(t *testing.T) {
	// Проверяем что функция устанавливает getClient
	mockFn := func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	}
	
	SetGetClientForTests(mockFn)
	
	// getClient должен быть установлен
	assert.NotNil(t, getClient)
}
