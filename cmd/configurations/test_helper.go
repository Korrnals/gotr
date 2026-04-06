package configurations

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// testHTTPClientKey is the context key for storing the HTTP client in tests.
const testHTTPClientKey = "httpClient"

// getClientForTests returns the client from the command context for use in tests.
func getClientForTests(cmd *cobra.Command) client.ClientInterface {
	if cmd == nil || cmd.Context() == nil {
		return nil
	}
	if mock, ok := cmd.Context().Value(testHTTPClientKey).(*client.MockClient); ok {
		return mock
	}
	return nil
}

// setupTestCmd creates a test command with a mock client in the context.
func setupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	if mock != nil {
		ctx := context.WithValue(context.Background(), testHTTPClientKey, mock)
		cmd.SetContext(ctx)
	}
	return cmd
}
