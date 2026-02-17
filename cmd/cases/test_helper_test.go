package cases

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSetupTestCmd(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupTestCmd(t, mock)

	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.Context())

	// Verify the mock client is in the context
	val := cmd.Context().Value(testHTTPClientKey)
	assert.NotNil(t, val)
	_, ok := val.(*client.MockClient)
	assert.True(t, ok, "should be *client.MockClient")
}

func TestGetClientForTests_WithMockClient(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupTestCmd(t, mock)

	result := getClientForTests(cmd)
	assert.NotNil(t, result)
	// Should return the mock client
	_, ok := result.(*client.MockClient)
	assert.True(t, ok, "should return *client.MockClient")
}

func TestGetClientForTests_WithClientInterface(t *testing.T) {
	// Create a mock that implements ClientInterface
	var mock client.ClientInterface = &client.MockClient{}

	// Create command with client.ClientInterface in context
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testHTTPClientKey, mock)
	cmd.SetContext(ctx)

	result := getClientForTests(cmd)
	assert.NotNil(t, result)
}

func TestGetClientForTests_WithNilValue(t *testing.T) {
	// Create command with explicit nil value in context
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testHTTPClientKey, nil)
	cmd.SetContext(ctx)

	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

func TestGetClientForTests_NilCmd(t *testing.T) {
	result := getClientForTests(nil)
	assert.Nil(t, result)
}

func TestGetClientForTests_NilContext(t *testing.T) {
	cmd := &cobra.Command{}
	// No context set

	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

func TestGetClientForTests_NoValue(t *testing.T) {
	cmd := &cobra.Command{}
	ctx := context.Background()
	cmd.SetContext(ctx)

	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

func TestGetClientForTests_InvalidType(t *testing.T) {
	cmd := &cobra.Command{}
	// Set a value with wrong type
	ctx := context.WithValue(context.Background(), testHTTPClientKey, "not a client")
	cmd.SetContext(ctx)

	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

func TestGetClientForTests_IntValue(t *testing.T) {
	cmd := &cobra.Command{}
	// Set a value with int type (not a valid client type)
	ctx := context.WithValue(context.Background(), testHTTPClientKey, 42)
	cmd.SetContext(ctx)

	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

func TestGetClientForTests_StructValue(t *testing.T) {
	cmd := &cobra.Command{}
	// Set a value with a struct that's not a client
	ctx := context.WithValue(context.Background(), testHTTPClientKey, struct{ name string }{name: "test"})
	cmd.SetContext(ctx)

	result := getClientForTests(cmd)
	assert.Nil(t, result)
}
