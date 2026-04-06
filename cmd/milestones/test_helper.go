package milestones

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// testHTTPClientKey is the context key for storing the HTTP client in tests.
const testHTTPClientKey = "httpClient"

// setupTestCmd creates a test command with a mock client in context.
func setupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testHTTPClientKey, mock)
	cmd.SetContext(ctx)
	return cmd
}

// getClientForTests retrieves the client from context for use in tests.
func getClientForTests(cmd *cobra.Command) client.ClientInterface {
	val := cmd.Context().Value(testHTTPClientKey)
	if val == nil {
		return nil
	}
	if c, ok := val.(client.ClientInterface); ok {
		return c
	}
	return nil
}
