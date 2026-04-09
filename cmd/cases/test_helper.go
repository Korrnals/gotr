package cases

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// testContextKey is an unexported key type for context values in tests.
type testContextKey string

// testHTTPClientKey is the context key for tests (must match cmd.httpClientKey).
const testHTTPClientKey testContextKey = "httpClient"

// setupTestCmd creates a command with a mock client in the context.
func setupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testHTTPClientKey, mock)
	cmd.SetContext(ctx)
	return cmd
}

// getClientForTests extracts the client from context for tests.
func getClientForTests(cmd *cobra.Command) client.ClientInterface {
	if cmd == nil || cmd.Context() == nil {
		return nil
	}
	val := cmd.Context().Value(testHTTPClientKey)
	if val == nil {
		return nil
	}
	if c, ok := val.(client.ClientInterface); ok {
		return c
	}
	return nil
}
