// internal/client/common_test.go
// Common infrastructure for HTTP client tests
package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockServer creates a test HTTP server
type mockServer struct {
	*httptest.Server
}

// newMockServer creates a server with the given handler
func newMockServer(handler http.HandlerFunc) *mockServer {
	return &mockServer{
		Server: httptest.NewServer(handler),
	}
}

// mockClient creates a client with a test server
func mockClient(t *testing.T, handler http.HandlerFunc) (*HTTPClient, *mockServer) {
	server := newMockServer(handler)
	client, err := NewClient(server.URL, "test@test.com", "testpass", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client, server
}

// ptr returns a pointer to the value (test helper)
func ptr[T any](v T) *T {
	return &v
}
