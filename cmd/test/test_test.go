package test

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// mockContextKey - type for context key
type mockContextKey struct{}

// ==================== Tests for getClientInterface ====================

func TestGetClientInterface_WithNilAccessor(t *testing.T) {
	// Save the old value
	oldAccessor := clientAccessor
	clientAccessor = nil
	defer func() { clientAccessor = oldAccessor }()

	cmd := &cobra.Command{}
	result := getClientInterface(cmd)
	assert.Nil(t, result)
}

func TestGetClientInterface_WithAccessor(t *testing.T) {
	// Save the old value
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	// Create a command
	cmd := &cobra.Command{}

	// First set the accessor
	SetGetClientForTests(func(ctx context.Context) client.ClientInterface {
		return nil
	})

	result := getClientInterface(cmd)
	// accessor returns nil because HTTPClient is not set
	assert.Nil(t, result)
}

// ==================== Tests for SetGetClientForTests ====================

func TestSetGetClientForTests_WithNilAccessor(t *testing.T) {
	// Save the old value
	oldAccessor := clientAccessor
	clientAccessor = nil
	defer func() { clientAccessor = oldAccessor }()

	fn := func(ctx context.Context) client.ClientInterface {
		return nil
	}

	// Should not panic
	SetGetClientForTests(fn)
	assert.NotNil(t, clientAccessor)
}

func TestSetGetClientForTests_WithExistingAccessor(t *testing.T) {
	// First initialize the accessor
	oldAccessor := clientAccessor
	SetGetClientForTests(func(ctx context.Context) client.ClientInterface { return nil })
	defer func() { clientAccessor = oldAccessor }()

	fn := func(ctx context.Context) client.ClientInterface {
		return nil
	}

	// Should not panic
	SetGetClientForTests(fn)
	assert.NotNil(t, clientAccessor)
}

// ==================== Tests for Register ====================

func TestRegister(t *testing.T) {
	root := &cobra.Command{}

	Register(root, func(ctx context.Context) client.ClientInterface {
		return nil
	})

	// Verify that the test command is added
	testCmd, _, err := root.Find([]string{"test"})
	assert.NoError(t, err)
	assert.NotNil(t, testCmd)

	// Verify subcommands exist
	getCmd, _, _ := root.Find([]string{"test", "get"})
	assert.NotNil(t, getCmd)

	listCmd, _, _ := root.Find([]string{"test", "list"})
	assert.NotNil(t, listCmd)
}

func TestRegister_Help(t *testing.T) {
	root := &cobra.Command{}

	Register(root, func(ctx context.Context) client.ClientInterface {
		return nil
	})

	testCmd, _, err := root.Find([]string{"test"})
	assert.NoError(t, err)
	assert.NotNil(t, testCmd)

	// Verify that calling without arguments shows help
	root.SetArgs([]string{"test"})
	err = root.Execute()
	assert.NoError(t, err)
}

func TestRegister_NoLocalQuietFlags(t *testing.T) {
	root := &cobra.Command{}

	Register(root, func(ctx context.Context) client.ClientInterface {
		return nil
	})

	testCmd, _, err := root.Find([]string{"test"})
	assert.NoError(t, err)
	assert.NotNil(t, testCmd)

	// Verify that quiet flag is not declared locally on subcommands.
	// Global quiet should be inherited from root persistent flags.
	for _, sub := range testCmd.Commands() {
		quietFlag := sub.Flags().Lookup("quiet")
		assert.Nil(t, quietFlag, "quiet should not be declared locally on subcommand %s", sub.Name())
	}
}

// ==================== Tests for Cmd ====================

func TestCmd_Help(t *testing.T) {
	// Verify that Help is called without errors
	err := Cmd.Help()
	assert.NoError(t, err)
}

func TestCmd_Properties(t *testing.T) {
	assert.Equal(t, "test", Cmd.Use)
	assert.NotEmpty(t, Cmd.Short)
	assert.NotEmpty(t, Cmd.Long)
}

// TestGetClientInterface_WithMockClientInContext verifies getting a mock client from context
func TestGetClientInterface_WithMockClientInContext(t *testing.T) {
	mock := &client.MockClient{}

	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), mockContextKey{}, mock)
	cmd.SetContext(ctx)

	// Directly test getting a value from context
	val := cmd.Context().Value(mockContextKey{})
	assert.NotNil(t, val)

	// Verify it is a MockClient
	if c, ok := val.(*client.MockClient); ok {
		assert.Equal(t, mock, c)
	}
}
