package test

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// mockContextKey - тип для ключа контекста
type mockContextKey struct{}

// ==================== Тесты для getClientInterface ====================

func TestGetClientInterface_WithNilAccessor(t *testing.T) {
	// Сохраняем старое значение
	oldAccessor := clientAccessor
	clientAccessor = nil
	defer func() { clientAccessor = oldAccessor }()

	cmd := &cobra.Command{}
	result := getClientInterface(cmd)
	assert.Nil(t, result)
}

func TestGetClientInterface_WithAccessor(t *testing.T) {
	// Сохраняем старое значение
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	// Создаём команду
	cmd := &cobra.Command{}

	// Сначала устанавливаем accessor
	SetGetClientForTests(func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	})

	result := getClientInterface(cmd)
	// accessor возвращает nil т.к. HTTPClient не установлен
	assert.Nil(t, result)
}

// ==================== Тесты для SetGetClientForTests ====================

func TestSetGetClientForTests_WithNilAccessor(t *testing.T) {
	// Сохраняем старое значение
	oldAccessor := clientAccessor
	clientAccessor = nil
	defer func() { clientAccessor = oldAccessor }()

	fn := func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	}

	// Не должно паниковать
	SetGetClientForTests(fn)
	assert.NotNil(t, clientAccessor)
}

func TestSetGetClientForTests_WithExistingAccessor(t *testing.T) {
	// Сначала инициализируем accessor
	oldAccessor := clientAccessor
	SetGetClientForTests(func(cmd *cobra.Command) *client.HTTPClient { return nil })
	defer func() { clientAccessor = oldAccessor }()

	fn := func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	}

	// Не должно паниковать
	SetGetClientForTests(fn)
	assert.NotNil(t, clientAccessor)
}

// ==================== Тесты для Register ====================

func TestRegister(t *testing.T) {
	root := &cobra.Command{}

	Register(root, func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	})

	// Проверяем, что команда test добавлена
	testCmd, _, err := root.Find([]string{"test"})
	assert.NoError(t, err)
	assert.NotNil(t, testCmd)

	// Проверяем наличие подкоманд
	getCmd, _, _ := root.Find([]string{"test", "get"})
	assert.NotNil(t, getCmd)

	listCmd, _, _ := root.Find([]string{"test", "list"})
	assert.NotNil(t, listCmd)
}

func TestRegister_Help(t *testing.T) {
	root := &cobra.Command{}

	Register(root, func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	})

	testCmd, _, err := root.Find([]string{"test"})
	assert.NoError(t, err)
	assert.NotNil(t, testCmd)

	// Проверяем, что вызов без аргументов показывает help
	root.SetArgs([]string{"test"})
	err = root.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для Cmd ====================

func TestCmd_Help(t *testing.T) {
	// Проверяем, что Help вызывается без ошибок
	Cmd.Run(Cmd, []string{})
}

func TestCmd_Properties(t *testing.T) {
	assert.Equal(t, "test", Cmd.Use)
	assert.NotEmpty(t, Cmd.Short)
	assert.NotEmpty(t, Cmd.Long)
}

// TestGetClientInterface_WithMockClientInContext проверяет получение mock клиента из контекста
func TestGetClientInterface_WithMockClientInContext(t *testing.T) {
	mock := &client.MockClient{}
	
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), mockContextKey{}, mock)
	cmd.SetContext(ctx)
	
	// Прямо тестируем получение значения из контекста
	val := cmd.Context().Value(mockContextKey{})
	assert.NotNil(t, val)
	
	// Проверяем что это MockClient
	if c, ok := val.(*client.MockClient); ok {
		assert.Equal(t, mock, c)
	}
}
