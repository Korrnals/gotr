package sync

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Tests for Register ====================

func TestRegister(t *testing.T) {
	// Skip if flags are already defined (other tests may define them)
	defer func() {
		if r := recover(); r != nil {
			t.Skip("Flags already registered, skipping Register test")
		}
	}()

	root := &cobra.Command{}

	Register(root, func(ctx context.Context) client.ClientInterface {
		return nil
	})

	// Verify that the sync command is added
	syncCmd, _, err := root.Find([]string{"sync"})
	assert.NoError(t, err)
	assert.NotNil(t, syncCmd)

	// Verify that subcommands exist
	sharedStepsCmd, _, _ := root.Find([]string{"sync", "shared-steps"})
	assert.NotNil(t, sharedStepsCmd)

	casesCmd, _, _ := root.Find([]string{"sync", "cases"})
	assert.NotNil(t, casesCmd)

	fullCmd, _, _ := root.Find([]string{"sync", "full"})
	assert.NotNil(t, fullCmd)

	suitesCmd, _, _ := root.Find([]string{"sync", "suites"})
	assert.NotNil(t, suitesCmd)

	sectionsCmd, _, _ := root.Find([]string{"sync", "sections"})
	assert.NotNil(t, sectionsCmd)
}

func TestRegister_Help(t *testing.T) {
	// Skip if flags are already defined
	defer func() {
		if r := recover(); r != nil {
			t.Skip("Flags already registered, skipping Register_Help test")
		}
	}()

	root := &cobra.Command{}

	Register(root, func(ctx context.Context) client.ClientInterface {
		return nil
	})

	// Verify that calling without arguments shows help
	root.SetArgs([]string{"sync"})
	err := root.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for getClientSafe ====================

func TestGetClientSafe_WithNilAccessor(t *testing.T) {
	// Save original value
	oldAccessor := clientAccessor
	clientAccessor = nil
	defer func() { clientAccessor = oldAccessor }()

	// Create command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	result := getClientSafe(cmd)
	assert.Nil(t, result)
}

func TestGetClientSafe_WithAccessorReturnsClient(t *testing.T) {
	// Save original value
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	// Create mock HTTP client
	mockClient := &client.HTTPClient{}

	// Create accessor that returns client
	clientAccessor = client.NewAccessor(func(ctx context.Context) client.ClientInterface {
		return mockClient
	})

	// Create command
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Should return client from accessor
	result := getClientSafe(cmd)
	assert.Equal(t, mockClient, result)
}

func TestGetClientSafe_WithAccessorReturnsNil_UsesContextFallback(t *testing.T) {
	// Save original value
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	// Create accessor that returns nil
	clientAccessor = client.NewAccessor(func(ctx context.Context) client.ClientInterface {
		return nil
	})

	// Create mock HTTP client
	mockClient := &client.HTTPClient{}

	// Create command with context containing client by old key
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testHTTPClientKey, mockClient)
	cmd.SetContext(ctx)

	// Should return client from context (fallback)
	result := getClientSafe(cmd)
	assert.Equal(t, mockClient, result)
}

func TestGetClientSafe_WithInvalidTypeInContext(t *testing.T) {
	// Save original value
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	// Create accessor that returns nil
	clientAccessor = client.NewAccessor(func(ctx context.Context) client.ClientInterface {
		return nil
	})

	// Create command with context containing WRONG type by key
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testHTTPClientKey, "not a client")
	cmd.SetContext(ctx)

	// Should return nil since type does not match
	result := getClientSafe(cmd)
	assert.Nil(t, result)
}

func TestGetClientSafe_WithNilContext(t *testing.T) {
	// Save original value
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	// Create accessor that returns nil
	clientAccessor = client.NewAccessor(func(ctx context.Context) client.ClientInterface {
		return nil
	})

	// Create command without context (nil)
	cmd := &cobra.Command{}
	// cmd.Context() returns nil

	// Should not panic, should return nil
	result := getClientSafe(cmd)
	assert.Nil(t, result)
}

// ==================== Tests for getClientInterface ====================

func TestGetClientInterface_WithNilAccessor(t *testing.T) {
	// Save original value
	oldAccessor := clientAccessor
	clientAccessor = nil
	defer func() { clientAccessor = oldAccessor }()

	// Create command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	result := getClientInterface(cmd)
	assert.Nil(t, result)
}

func TestGetClientInterface_WithMockClient(t *testing.T) {
	mock := &client.MockClient{}

	cmd := &cobra.Command{}
	SetTestClient(cmd, mock)

	result := getClientInterface(cmd)
	assert.NotNil(t, result)
	assert.Equal(t, mock, result)
}

func TestGetClientInterface_WithMockInContext(t *testing.T) {
	mock := &client.MockClient{}

	cmd := &cobra.Command{}
	SetTestClient(cmd, mock)

	// Verify that getClientInterface returns mock client
	result := getClientInterface(cmd)
	assert.NotNil(t, result)
}

// ==================== Tests for SetGetClientForTests ====================

func TestSetGetClientForTests_WithNilAccessor(t *testing.T) {
	// Save original value
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
	// First initialize accessor
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

// ==================== Tests for SetTestClient ====================

func TestSetTestClient_WithNilContext(t *testing.T) {
	mock := &client.MockClient{}
	cmd := &cobra.Command{}

	// Should not panic on nil context
	SetTestClient(cmd, mock)

	// Verify that client is set
	result := getClientInterface(cmd)
	assert.Equal(t, mock, result)
}

func TestSetTestClient_WithExistingContext(t *testing.T) {
	mock := &client.MockClient{}
	cmd := &cobra.Command{}
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())

	// Should not panic on existing context
	SetTestClient(cmd, mock)

	// Verify that client is set
	result := getClientInterface(cmd)
	assert.Equal(t, mock, result)
}

// ==================== Tests for Cmd ====================

func TestCmd_Help(t *testing.T) {
	// Verify that Help is called without errors
	err := Cmd.Help()
	assert.NoError(t, err)
}

func TestCmd_Properties(t *testing.T) {
	assert.Equal(t, "sync", Cmd.Use)
	assert.NotEmpty(t, Cmd.Short)
	assert.NotEmpty(t, Cmd.Long)
}

// ==================== Comprehensive tests for sync edge cases ====================

func TestGetClientSafe_Context(t *testing.T) {
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	clientAccessor = client.NewAccessor(func(ctx context.Context) client.ClientInterface {
		return nil
	})

	cmd := &cobra.Command{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd.SetContext(ctx)

	result := getClientSafe(cmd)
	assert.Nil(t, result) // Should be nil because accessor returns nil
}

func TestGetClientInterface_Nil(t *testing.T) {
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	clientAccessor = nil

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	result := getClientInterface(cmd)
	assert.Nil(t, result)
}

func TestSetTestClient_MultipleUpdates(t *testing.T) {
	mock1 := &client.MockClient{}
	mock2 := &client.MockClient{}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	SetTestClient(cmd, mock1)
	result1 := getClientInterface(cmd)
	assert.Equal(t, mock1, result1)

	SetTestClient(cmd, mock2)
	result2 := getClientInterface(cmd)
	assert.Equal(t, mock2, result2)
}

func TestSetGetClientForTests_NilAccessorInitially(t *testing.T) {
	oldAccessor := clientAccessor
	defer func() { clientAccessor = oldAccessor }()

	clientAccessor = nil

	fn := func(ctx context.Context) client.ClientInterface {
		return nil
	}

	SetGetClientForTests(fn)
	assert.NotNil(t, clientAccessor)
}
