// internal/client/cancel_test.go
// Тест проверяет, что отменённый контекст прерывает HTTP запрос
package client

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

// TestGetRuns_Cancellation проверяет критерий Stage 7.0:
// отменённый контекст должен возвращать context.Canceled из клиентских методов.
func TestGetRuns_Cancellation(t *testing.T) {
	// Сервер, который отвечает успешно (но клиент не должен до него дойти)
	server := newMockServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"runs":[],"offset":0,"limit":250,"size":0}`))
	}))
	defer server.Close()

	cli, err := NewClient(server.URL, "test@test.com", "testpass", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Уже отменённый контекст — http.Do вернёт context.Canceled до отправки запроса
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = cli.GetRuns(ctx, 1)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

// TestHTTPClient_CancelledContext проверяет что DoRequest уважает отменённый контекст.
func TestHTTPClient_CancelledContext(t *testing.T) {
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

