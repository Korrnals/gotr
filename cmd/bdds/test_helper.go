package bdds

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

const testHTTPClientKey = "httpClient"

func getClientForTests(cmd *cobra.Command) client.ClientInterface {
	if cmd == nil || cmd.Context() == nil {
		return nil
	}
	if mock, ok := cmd.Context().Value(testHTTPClientKey).(*client.MockClient); ok {
		return mock
	}
	return nil
}

func setupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	if mock != nil {
		ctx := context.WithValue(context.Background(), testHTTPClientKey, mock)
		cmd.SetContext(ctx)
	}
	return cmd
}
