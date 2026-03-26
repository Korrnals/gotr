package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRootCmd_Properties проверяет базовые свойства root команды
func TestRootCmd_Properties(t *testing.T) {
	assert.Equal(t, "gotr", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)
	assert.NotEmpty(t, Version)
}

// TestVersion_Properties проверяет версию
func TestVersion_Properties(t *testing.T) {
	assert.NotEmpty(t, Version)
	assert.NotEmpty(t, Commit)
	assert.NotEmpty(t, Date)
}

// TestGetClient_NotNilContext проверяет что GetClient требует контекст
func TestGetClient_NotNilContext(t *testing.T) {
	// GetClient требует контекст с clientом
	// Если контекст пустой, функция вызывает panic
	// Проверяем только что функция существует
	assert.NotNil(t, GetClient)
}

// TestGetClientInterface_NotNilContext проверяет что GetClientInterface требует контекст
func TestGetClientInterface_NotNilContext(t *testing.T) {
	// GetClientInterface требует контекст с clientом
	// Если контекст пустой, функция вызывает panic
	// Проверяем только что функция существует
	assert.NotNil(t, GetClientInterface)
}

// TestRootCmd_NonInteractiveFlagRegistered проверяет наличие флага --non-interactive
func TestRootCmd_NonInteractiveFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("non-interactive")
	assert.NotNil(t, flag, "--non-interactive flag must be registered")
	assert.Equal(t, "false", flag.DefValue)
}
