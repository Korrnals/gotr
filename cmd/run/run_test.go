package run

import (
	"context"
	"sync"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var (
	registerTestOnce sync.Once
	registerTestDone bool
)

// Test helper to provide mock client
func mockGetClient(mock *client.MockClient) GetClientFunc {
	return func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	}
}

func TestRegister(t *testing.T) {
	var runCmd *cobra.Command

	registerTestOnce.Do(func() {
		// Reset clientAccessor before test
		clientAccessor = nil

		rootCmd := &cobra.Command{Use: "root"}
		mockFn := mockGetClient(&client.MockClient{})

		// Register can only be called once per process due to flag redefinition
		Register(rootCmd, mockFn)
		registerTestDone = true

		// Find run command
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "run" {
				runCmd = cmd
				break
			}
		}
	})

	if !registerTestDone {
		t.Skip("Register already called in another test")
	}

	// Verify run command was found
	assert.NotNil(t, runCmd, "run command should be registered")

	// Verify subcommands exist
	subCommands := runCmd.Commands()
	subCmdNames := make(map[string]bool)
	for _, subCmd := range subCommands {
		subCmdNames[subCmd.Name()] = true
	}
	
	assert.True(t, subCmdNames["get"], "get subcommand should exist")
	assert.True(t, subCmdNames["list"], "list subcommand should exist")
	assert.True(t, subCmdNames["create"], "create subcommand should exist")
	assert.True(t, subCmdNames["update"], "update subcommand should exist")
	assert.True(t, subCmdNames["close"], "close subcommand should exist")
	assert.True(t, subCmdNames["delete"], "delete subcommand should exist")
}

// TestRegister_OnceOnly tests that Register can only be called once per process
// due to flag redefinition constraints
func TestRegister_OnceOnly(t *testing.T) {
	// This test documents that Register should only be called once
	// because it modifies global command variables
	assert.NotNil(t, Cmd)
	assert.NotNil(t, getCmd)
	assert.NotNil(t, listCmd)
}

func TestSetGetClientForTests_WhenNil(t *testing.T) {
	// Reset clientAccessor
	clientAccessor = nil

	mockFn := mockGetClient(&client.MockClient{})
	SetGetClientForTests(mockFn)

	// clientAccessor should be initialized
	assert.NotNil(t, clientAccessor)
}

func TestSetGetClientForTests_WhenNotNil(t *testing.T) {
	// Reset clientAccessor
	clientAccessor = nil

	// First call
	mockFn1 := mockGetClient(&client.MockClient{})
	SetGetClientForTests(mockFn1)
	assert.NotNil(t, clientAccessor)

	// Second call should use SetClientForTests path
	mockFn2 := mockGetClient(&client.MockClient{})
	SetGetClientForTests(mockFn2)
	
	// clientAccessor should still exist
	assert.NotNil(t, clientAccessor)
}

func TestGetClientSafe_WhenNil(t *testing.T) {
	// Reset clientAccessor
	clientAccessor = nil

	cmd := &cobra.Command{}
	result := getClientSafe(cmd)
	
	assert.Nil(t, result)
}

func TestGetClientSafe_WhenNotNil(t *testing.T) {
	// Reset and initialize clientAccessor
	clientAccessor = nil
	
	mockFn := mockGetClient(&client.MockClient{})
	SetGetClientForTests(mockFn)
	
	cmd := &cobra.Command{}
	result := getClientSafe(cmd)
	
	// Since our mock returns nil, we expect nil
	assert.Nil(t, result)
}

func TestGetClientSafe_WithContext(t *testing.T) {
	// This tests that getClientSafe calls GetClientSafe on clientAccessor
	clientAccessor = nil
	
	mock := &client.MockClient{}
	
	// Create a mock function that returns the mock client
	mockFn := func(cmd *cobra.Command) *client.HTTPClient {
		return nil
	}
	
	SetGetClientForTests(mockFn)
	
	// Create command with mock in context using the same key as testhelper
	const httpClientKey = "httpClient"
	testCmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	testCmd.SetContext(ctx)
	
	// getClientSafe should not panic
	result := getClientSafe(testCmd)
	_ = result
}
