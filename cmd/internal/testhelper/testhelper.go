// Package testhelper provides common utilities for testing CLI commands.
// This is an internal package, accessible only within cmd/.
package testhelper

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// HTTPClientKey is the context key for storing the HTTP client in tests.
// Must match the key used in the main code.
const HTTPClientKey = "httpClient"

// SetupTestCmd creates a test command with a mock client in its context.
func SetupTestCmd(t *testing.T, mock *client.MockClient) *cobra.Command {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), HTTPClientKey, mock)
	cmd.SetContext(ctx)
	return cmd
}

// GetClientForTests retrieves the client from context for use in tests.
// Returns nil if the client is not found or context is empty.
func GetClientForTests(cmd *cobra.Command) client.ClientInterface {
	if cmd == nil || cmd.Context() == nil {
		return nil
	}
	if mock, ok := cmd.Context().Value(HTTPClientKey).(*client.MockClient); ok {
		return mock
	}
	// Also try interface assertion
	if c, ok := cmd.Context().Value(HTTPClientKey).(client.ClientInterface); ok {
		return c
	}
	return nil
}

// SetupTestCmdWithBuffer creates a test command with a mock client and output buffer.
// Used when command output needs to be verified.
func SetupTestCmdWithBuffer(t *testing.T, mock *client.MockClient) (*cobra.Command, *cobra.Command) {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), HTTPClientKey, mock)
	cmd.SetContext(ctx)
	return cmd, cmd
}
