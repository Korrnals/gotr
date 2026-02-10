package groups

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// testHTTPClientKey — ключ для хранения HTTP клиента в контексте тестов
const testHTTPClientKey = "httpClient"

// getClientForTests возвращает клиент из контекста для использования в тестах
func getClientForTests(cmd *cobra.Command) client.ClientInterface {
	if cmd == nil || cmd.Context() == nil {
		return nil
	}
	if mock, ok := cmd.Context().Value(testHTTPClientKey).(*client.MockClient); ok {
		return mock
	}
	return nil
}

// setupTestCmd создаёт тестовую команду с mock клиентом в контексте
func setupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	if mock != nil {
		ctx := context.WithValue(context.Background(), testHTTPClientKey, mock)
		cmd.SetContext(ctx)
	}
	return cmd
}
