package plans

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// testHTTPClientKey — ключ для контекста в тестах (должен совпадать с cmd.httpClientKey)
const testHTTPClientKey = "httpClient"

// setupTestCmd создаёт команду с mock клиентом в контексте
func setupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testHTTPClientKey, mock)
	cmd.SetContext(ctx)
	return cmd
}

// getClientForTests извлекает клиент из контекста для тестов
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
