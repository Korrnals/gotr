// Package client provides HTTP client access utilities for CLI commands.
package client

import (
	"github.com/spf13/cobra"
)

// GetClientFunc is a function type for getting the HTTP client.
type GetClientFunc func(cmd *cobra.Command) *HTTPClient

// Accessor provides access to the HTTP client for commands.
type Accessor struct {
	getClient GetClientFunc
}

// NewAccessor creates a new accessor with the given client retrieval function.
func NewAccessor(fn GetClientFunc) *Accessor {
	return &Accessor{getClient: fn}
}

// GetClientSafe safely retrieves the client with nil check.
func (ca *Accessor) GetClientSafe(cmd *cobra.Command) *HTTPClient {
	if ca.getClient == nil {
		return nil
	}
	return ca.getClient(cmd)
}

// SetClientForTests sets the client retrieval function for tests.
func (ca *Accessor) SetClientForTests(fn GetClientFunc) {
	ca.getClient = fn
}

// GetClientSafeGlobal safely calls getClient with nil check (global function).
func GetClientSafeGlobal(cmd *cobra.Command, getClient GetClientFunc) *HTTPClient {
	if getClient == nil {
		return nil
	}
	return getClient(cmd)
}
