// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// httpClientKey — ключ для хранения HTTP clientа в контексте тестов.
// Должен совпадать с ключом, используемым в основном коде.
const httpClientKey = "httpClient"

// setupTestCmd создаёт тестовую команду с mock clientом в контексте.
// Используется в тестах для инжекции mock clientа.
func setupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	cmd.SetContext(ctx)
	return cmd
}

// getClientForTests возвращает client из контекста для использования в тестах.
// Возвращает nil если client не найден или контекст пуст.
func getClientForTests(cmd *cobra.Command) client.ClientInterface {
	if cmd == nil || cmd.Context() == nil {
		return nil
	}
	if mock, ok := cmd.Context().Value(httpClientKey).(*client.MockClient); ok {
		return mock
	}
	// Пробуем также интерфейс
	if c, ok := cmd.Context().Value(httpClientKey).(client.ClientInterface); ok {
		return c
	}
	return nil
}
