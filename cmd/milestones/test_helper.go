package milestones

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// testHTTPClientKey — ключ для хранения HTTP клиента в контексте тестов
const testHTTPClientKey = "httpClient"

// setupTestCmd создаёт тестовую команду с mock клиентом в контексте
func setupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testHTTPClientKey, mock)
	cmd.SetContext(ctx)
	return cmd
}

// getClientForTests возвращает клиент из контекста для использования в тестах
func getClientForTests(cmd *cobra.Command) client.ClientInterface {
	val := cmd.Context().Value(testHTTPClientKey)
	if val == nil {
		return nil
	}
	if c, ok := val.(client.ClientInterface); ok {
		return c
	}
	if c, ok := val.(*client.MockClient); ok {
		return c
	}
	return nil
}
