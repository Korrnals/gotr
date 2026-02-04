// internal/client/common_test.go
// Общая инфраструктура для тестов HTTP клиента
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockServer создаёт тестовый HTTP сервер
type mockServer struct {
	*httptest.Server
	requests []mockRequest
}

// mockRequest сохраняет данные о запросе для проверки
type mockRequest struct {
	Method  string
	URL     string
	Headers http.Header
	Body    []byte
}

// newMockServer создаёт сервер с заданным handler
func newMockServer(handler http.HandlerFunc) *mockServer {
	return &mockServer{
		Server: httptest.NewServer(handler),
	}
}

// newMockServerWithRecorder создаёт сервер, который записывает все запросы
func newMockServerWithRecorder(responseCode int, responseBody interface{}) *mockServer {
	ms := &mockServer{}
	ms.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Читаем тело запроса
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		
		// Сохраняем запрос
		ms.requests = append(ms.requests, mockRequest{
			Method:  r.Method,
			URL:     r.URL.String(),
			Headers: r.Header,
			Body:    body,
		})
		
		// Отправляем ответ
		w.WriteHeader(responseCode)
		if responseBody != nil {
			json.NewEncoder(w).Encode(responseBody)
		}
	}))
	return ms
}

// mockClient создаёт клиент с тестовым сервером
func mockClient(t *testing.T, handler http.HandlerFunc) (*HTTPClient, *mockServer) {
	server := newMockServer(handler)
	client, err := NewClient(server.URL, "test@test.com", "testpass", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client, server
}

// assertRequest проверяет параметры последнего запроса
func (ms *mockServer) assertRequest(t *testing.T, expectedMethod, expectedPath string) {
	t.Helper()
	if len(ms.requests) == 0 {
		t.Fatal("no requests recorded")
	}
	last := ms.requests[len(ms.requests)-1]
	if last.Method != expectedMethod {
		t.Errorf("expected method %s, got %s", expectedMethod, last.Method)
	}
	if last.URL != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, last.URL)
	}
}

// ptr возвращает указатель на значение (хелпер для тестов)
func ptr[T any](v T) *T {
	return &v
}
