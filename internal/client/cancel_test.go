// internal/client/cancel_test.go
// Test verifies that a canceled context aborts an HTTP request
package client

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

// TestGetRuns_Cancellation verifies Stage 7.0 criterion:
// a canceled context must return context.Canceled from client methods.
func TestGetRuns_Cancellation(t *testing.T) {
	// Server that responds successfully (but the client should not reach it)
	server := newMockServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"runs":[],"offset":0,"limit":250,"size":0}`))
	}))
	defer server.Close()

	cli, err := NewClient(server.URL, "test@test.com", "testpass", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Already canceled context — http.Do will return context.Canceled before sending request
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = cli.GetRuns(ctx, 1)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

// TestHTTPClient_CanceledContext verifies that DoRequest respects a canceled context.
func TestHTTPClient_CanceledContext(t *testing.T) {
	server := newMockServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cli, err := NewClient(server.URL, "test@test.com", "testpass", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = cli.DoRequest(ctx, http.MethodGet, "get_runs/1", nil, nil)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("DoRequest: expected context.Canceled, got: %v", err)
	}
}

