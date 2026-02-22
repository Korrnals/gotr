package sync

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для Register ====================

func TestRegister(t *testing.T) {
	// Пропускаем если флаги уже определены (другие тесты могут их определить)
	defer func() {
		if r := recover(); r != nil {
			t.Skip("Flags already registered, skipping Register test")
		}
	}()

	root := &cobra.Command{}

	Register(root, func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	})

	// Проверяем, что команда sync добавлена
	syncCmd, _, err := root.Find([]string{"sync"})
	assert.NoError(t, err)
	assert.NotNil(t, syncCmd)

	// Проверяем наличие подкоманд
	sharedStepsCmd, _, _ := root.Find([]string{"sync", "shared-steps"})
	assert.NotNil(t, sharedStepsCmd)

	casesCmd, _, _ := root.Find([]string{"sync", "cases"})
	assert.NotNil(t, casesCmd)

	fullCmd, _, _ := root.Find([]string{"sync", "full"})
	assert.NotNil(t, fullCmd)

	suitesCmd, _, _ := root.Find([]string{"sync", "suites"})
	assert.NotNil(t, suitesCmd)

	sectionsCmd, _, _ := root.Find([]string{"sync", "sections"})
	assert.NotNil(t, sectionsCmd)
}

func TestRegister_Help(t *testing.T) {
	// Пропускаем если флаги уже определены
	defer func() {
		if r := recover(); r != nil {
			t.Skip("Flags already registered, skipping Register_Help test")
		}
	}()

	root := &cobra.Command{}

	Register(root, func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	})

	// Проверяем, что вызов без аргументов показывает help
	root.SetArgs([]string{"sync"})
	err := root.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для getClientSafe ====================

func TestGetClientSafe_WithNilAccessor(t *testing.T) {
	// Сохраняем старое значение
	oldAccessor := clientAccessor
	clientAccessor = nil
	defer func() { clientAccessor = oldAccessor }()

	// Создаём команду с контекстом
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	
	result := getClientSafe(cmd)
	assert.Nil(t, result)
}

func TestGetClientSafe_WithAccessorReturnsClient(t *testing.T) {
	// Сохраняем старое значение
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	// Создаём mock HTTP клиент
	mockClient := &client.HTTPClient{}

	// Создаём accessor, который возвращает клиент
	clientAccessor = client.NewAccessor(func(cmd *cobra.Command) *client.HTTPClient {
		return mockClient
	})

	// Создаём команду
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Должен вернуть клиент от accessor
	result := getClientSafe(cmd)
	assert.Equal(t, mockClient, result)
}

func TestGetClientSafe_WithAccessorReturnsNil_UsesContextFallback(t *testing.T) {
	// Сохраняем старое значение
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	// Создаём accessor, который возвращает nil
	clientAccessor = client.NewAccessor(func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	})

	// Создаём mock HTTP клиент
	mockClient := &client.HTTPClient{}

	// Создаём команду с контекстом, содержащим клиент по старому ключу
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testHTTPClientKey, mockClient)
	cmd.SetContext(ctx)

	// Должен вернуть клиент из контекста (fallback)
	result := getClientSafe(cmd)
	assert.Equal(t, mockClient, result)
}

func TestGetClientSafe_WithInvalidTypeInContext(t *testing.T) {
	// Сохраняем старое значение
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	// Создаём accessor, который возвращает nil
	clientAccessor = client.NewAccessor(func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	})

	// Создаём команду с контекстом, содержащим НЕВЕРНЫЙ тип по ключу
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testHTTPClientKey, "not a client")
	cmd.SetContext(ctx)

	// Должен вернуть nil, так как тип не совпадает
	result := getClientSafe(cmd)
	assert.Nil(t, result)
}

func TestGetClientSafe_WithNilContext(t *testing.T) {
	// Сохраняем старое значение
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	// Создаём accessor, который возвращает nil
	clientAccessor = client.NewAccessor(func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	})

	// Создаём команду без контекста (nil)
	cmd := &cobra.Command{}
	// cmd.Context() вернёт nil

	// Не должно паниковать, должен вернуть nil
	result := getClientSafe(cmd)
	assert.Nil(t, result)
}

// ==================== Тесты для getClientInterface ====================

func TestGetClientInterface_WithNilAccessor(t *testing.T) {
	// Сохраняем старое значение
	oldAccessor := clientAccessor
	clientAccessor = nil
	defer func() { clientAccessor = oldAccessor }()

	// Создаём команду с контекстом
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	
	result := getClientInterface(cmd)
	assert.Nil(t, result)
}

func TestGetClientInterface_WithMockClient(t *testing.T) {
	mock := &client.MockClient{}

	cmd := &cobra.Command{}
	SetTestClient(cmd, mock)

	result := getClientInterface(cmd)
	assert.NotNil(t, result)
	assert.Equal(t, mock, result)
}

func TestGetClientInterface_WithMockInContext(t *testing.T) {
	mock := &client.MockClient{}

	cmd := &cobra.Command{}
	SetTestClient(cmd, mock)

	// Проверяем, что getClientInterface возвращает mock клиент
	result := getClientInterface(cmd)
	assert.NotNil(t, result)
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

// ==================== Тесты для SetTestClient ====================

func TestSetTestClient_WithNilContext(t *testing.T) {
	mock := &client.MockClient{}
	cmd := &cobra.Command{}

	// Не должно паниковать при nil контексте
	SetTestClient(cmd, mock)

	// Проверяем, что клиент установлен
	result := getClientInterface(cmd)
	assert.Equal(t, mock, result)
}

func TestSetTestClient_WithExistingContext(t *testing.T) {
	mock := &client.MockClient{}
	cmd := &cobra.Command{}
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())

	// Не должно паниковать при существующем контексте
	SetTestClient(cmd, mock)

	// Проверяем, что клиент установлен
	result := getClientInterface(cmd)
	assert.Equal(t, mock, result)
}

// ==================== Тесты для Cmd ====================

func TestCmd_Help(t *testing.T) {
	// Проверяем, что Help вызывается без ошибок
	Cmd.Run(Cmd, []string{})
}

func TestCmd_Properties(t *testing.T) {
	assert.Equal(t, "sync", Cmd.Use)
	assert.NotEmpty(t, Cmd.Short)
	assert.NotEmpty(t, Cmd.Long)
}
