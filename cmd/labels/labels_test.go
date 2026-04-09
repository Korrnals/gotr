// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// testContextKey is an unexported key type for context values in tests.
type testContextKey string

// httpClientKey is the key for storing the HTTP client in the test context.
// Must match the key used in the main code.
const httpClientKey testContextKey = "httpClient"

// setupTestCmd creates a test command with a mock client in the context.
// Used in tests for mock client injection.
func setupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	cmd.SetContext(ctx)
	return cmd
}

// getClientForTests returns the client from the context for use in tests.
// Returns nil if the client is not found or the context is empty.
func getClientForTests(cmd *cobra.Command) client.ClientInterface {
	if cmd == nil || cmd.Context() == nil {
		return nil
	}
	if mock, ok := cmd.Context().Value(httpClientKey).(*client.MockClient); ok {
		return mock
	}
	// Also try the interface
	if c, ok := cmd.Context().Value(httpClientKey).(client.ClientInterface); ok {
		return c
	}
	return nil
}
