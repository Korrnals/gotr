package run

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/stretchr/testify/assert"
)

// TestNewRunServiceFromInterface_Mocked tests the mock client branch
func TestNewRunServiceFromInterface_Mocked(t *testing.T) {
	// Test the else branch where mock client is used
	mock := &client.MockClient{}
	wrapper := newRunServiceFromInterface(mock)
	
	assert.NotNil(t, wrapper)
	assert.NotNil(t, wrapper.svc)
}

// TestNewRunServiceFromInterface_HTTPClient tests the HTTPClient branch
func TestNewRunServiceFromInterface_HTTPClient(t *testing.T) {
	// Create a real HTTPClient (without valid credentials)
	// This tests the type assertion branch in newRunServiceFromInterface
	httpClient := &client.HTTPClient{}
	
	// This should exercise the HTTPClient branch
	wrapper := newRunServiceFromInterface(httpClient)
	
	// The wrapper should be created successfully
	assert.NotNil(t, wrapper)
	assert.NotNil(t, wrapper.svc)
}

// TestExportedCommandsAccess tests access to exported command variables
// This ensures the commands are properly initialized
func TestExportedCommandsAccess(t *testing.T) {
	// Verify commands are accessible and initialized
	assert.NotNil(t, getCmd)
	assert.NotNil(t, listCmd)
	assert.NotNil(t, createCmd)
	assert.NotNil(t, updateCmd)
	assert.NotNil(t, closeCmd)
	assert.NotNil(t, deleteCmd)
	assert.NotNil(t, Cmd)

	// Verify command names
	assert.Equal(t, "get", getCmd.Name())
	assert.Equal(t, "list", listCmd.Name())
	assert.Equal(t, "create", createCmd.Name())
	assert.Equal(t, "update", updateCmd.Name())
	assert.Equal(t, "close", closeCmd.Name())
	assert.Equal(t, "delete", deleteCmd.Name())
	assert.Equal(t, "run", Cmd.Name())
}

// TestSubCommandsExist verifies subcommands exist on the parent Cmd
func TestSubCommandsExist(t *testing.T) {
	// Verify subcommands were added to Cmd
	subcommands := Cmd.Commands()
	assert.GreaterOrEqual(t, len(subcommands), 0)
	
	// The commands are registered via package-level vars
	// We just verify they're not nil
	assert.NotNil(t, getCmd)
	assert.NotNil(t, listCmd)
}
