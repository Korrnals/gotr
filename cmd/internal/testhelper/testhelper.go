// Package testhelper предоставляет общие утилиты для тестирования CLI команд.
// Этот пакет internal и доступен только внутри cmd/.
package testhelper

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// HTTPClientKey — ключ для хранения HTTP клиента в контексте тестов.
// Должен совпадать с ключом, используемым в основном коде.
const HTTPClientKey = "httpClient"

// SetupTestCmd создаёт тестовую команду с mock клиентом в контексте.
// Используется в тестах для инжекции mock клиента.
func SetupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), HTTPClientKey, mock)
	cmd.SetContext(ctx)
	return cmd
}

// GetClientForTests возвращает клиент из контекста для использования в тестах.
// Возвращает nil если клиент не найден или контекст пуст.
func GetClientForTests(cmd *cobra.Command) client.ClientInterface {
	if cmd == nil || cmd.Context() == nil {
		return nil
	}
	if mock, ok := cmd.Context().Value(HTTPClientKey).(*client.MockClient); ok {
		return mock
	}
	// Пробуем также интерфейс
	if c, ok := cmd.Context().Value(HTTPClientKey).(client.ClientInterface); ok {
		return c
	}
	return nil
}

// SetupTestCmdWithBuffer создаёт тестовую команду с mock клиентом и буфером вывода.
// Используется когда нужно проверить вывод команды.
func SetupTestCmdWithBuffer(t *testing.T, mock *client.MockClient) (*cobra.Command, *cobra.Command) {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), HTTPClientKey, mock)
	cmd.SetContext(ctx)
	return cmd, cmd
}
