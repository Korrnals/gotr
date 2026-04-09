// Package client provides HTTP client access utilities for CLI commands.
package client

import (
	"context"
)

// GetClientFunc is a function type for getting the HTTP client from a context.
// Decoupled from cobra — callers pass cmd.Context() instead of *cobra.Command.
type GetClientFunc func(ctx context.Context) ClientInterface

// Accessor provides access to the HTTP client for commands.
type Accessor struct {
	getClient GetClientFunc
}

// NewAccessor creates a new accessor with the given client retrieval function.
func NewAccessor(fn GetClientFunc) *Accessor {
	return &Accessor{getClient: fn}
}

// GetClientSafe safely retrieves the client with nil check.
func (ca *Accessor) GetClientSafe(ctx context.Context) ClientInterface {
	if ca.getClient == nil {
		return nil
	}
	return ca.getClient(ctx)
}

// SetClientForTests sets the client retrieval function for tests.
func (ca *Accessor) SetClientForTests(fn GetClientFunc) {
	ca.getClient = fn
}

// GetClientSafeGlobal safely calls getClient with nil check (global function).
func GetClientSafeGlobal(ctx context.Context, getClient GetClientFunc) ClientInterface {
	if getClient == nil {
		return nil
	}
	return getClient(ctx)
}
