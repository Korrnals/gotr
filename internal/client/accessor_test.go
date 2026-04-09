package client

import (
	"context"
	"testing"
)

func TestAccessorGetClientSafe(t *testing.T) {
	ctx := context.Background()

	accessor := NewAccessor(nil)
	if got := accessor.GetClientSafe(ctx); got != nil {
		t.Fatalf("expected nil client when getClient is nil")
	}

	var called bool
	accessor.SetClientForTests(func(ctx context.Context) ClientInterface {
		called = true
		return &HTTPClient{}
	})

	if got := accessor.GetClientSafe(ctx); got == nil {
		t.Fatalf("expected non-nil client when getClient is set")
	}
	if !called {
		t.Fatalf("expected configured getClient to be called")
	}
}

func TestGetClientSafeGlobal(t *testing.T) {
	ctx := context.Background()

	if got := GetClientSafeGlobal(ctx, nil); got != nil {
		t.Fatalf("expected nil from GetClientSafeGlobal with nil function")
	}

	var called bool
	got := GetClientSafeGlobal(ctx, func(ctx context.Context) ClientInterface {
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
