package client

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAccessorGetClientSafe(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	accessor := NewAccessor(nil)
	if got := accessor.GetClientSafe(cmd); got != nil {
		t.Fatalf("expected nil client when getClient is nil")
	}

	var called bool
	accessor.SetClientForTests(func(cmd *cobra.Command) *HTTPClient {
		called = true
		return &HTTPClient{}
	})

	if got := accessor.GetClientSafe(cmd); got == nil {
		t.Fatalf("expected non-nil client when getClient is set")
	}
	if !called {
		t.Fatalf("expected configured getClient to be called")
	}
}

func TestGetClientSafeGlobal(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	if got := GetClientSafeGlobal(cmd, nil); got != nil {
		t.Fatalf("expected nil from GetClientSafeGlobal with nil function")
	}

	var called bool
	got := GetClientSafeGlobal(cmd, func(cmd *cobra.Command) *HTTPClient {
		called = true
		return &HTTPClient{}
	})
	if got == nil {
		t.Fatalf("expected non-nil client from GetClientSafeGlobal")
	}
	if !called {
		t.Fatalf("expected global getClient function to be called")
	}
}
