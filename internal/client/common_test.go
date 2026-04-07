// internal/client/common_test.go
// Общая инфраструктура для тестов HTTP clientа
package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockServer создаёт тестовый HTTP сервер
type mockServer struct {
	*httptest.Server
}

// newMockServer создаёт сервер с заданным handler
func newMockServer(handler http.HandlerFunc) *mockServer {
	return &mockServer{
		Server: httptest.NewServer(handler),
	}
}

// mockClient создаёт client с тестовым сервером
func mockClient(t *testing.T, handler http.HandlerFunc) (*HTTPClient, *mockServer) {
	server := newMockServer(handler)
	client, err := NewClient(server.URL, "test@test.com", "testpass", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client, server
}

// ptr возвращает указатель на значение (хелпер для тестов)
func ptr[T any](v T) *T {
	return &v
}
