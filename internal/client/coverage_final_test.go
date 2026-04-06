// internal/client/coverage_final_test.go
// Targeted tests for remaining uncovered lines.
package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
)

// errReadCloser is an io.ReadCloser whose Read always returns an error.
type errReadCloser struct{ err error }

func (e *errReadCloser) Read(_ []byte) (int, error) { return 0, e.err }
func (e *errReadCloser) Close() error               { return nil }

// nopPaginatedReporter satisfies concurrency.PaginatedProgressReporter.
type nopPaginatedReporter struct{ calls atomic.Int32 }

func (n *nopPaginatedReporter) OnItemComplete()     { n.calls.Add(1) }
func (n *nopPaginatedReporter) OnBatchReceived(int) {}
func (n *nopPaginatedReporter) OnError()            {}
func (n *nopPaginatedReporter) OnPageFetched()      {}

// ---------- GetCasesWithProgress: io.ReadAll error ----------

func TestGetCasesWithProgress_ReadBodyError(t *testing.T) {
client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
})
defer server.Close()

client.client.Transport = &staticRoundTripper{resp: &http.Response{
StatusCode: http.StatusOK,
Status:     "200 OK",
Body:       &errReadCloser{err: errors.New("network read failure")},
Header:     make(http.Header),
}}

_, err := client.GetCasesWithProgress(context.Background(), 1, 0, 0, nil)
if err == nil {
t.Fatal("expected error from read failure, got nil")
}
}

// ---------- FetchPageCtx: io.ReadAll error ----------
// ---------- FetchPageCtx: client.Get transport error ----------

func TestFetchPageCtx_GetTransportError(t *testing.T) {
	c, err := NewClient("http://example.com", "test@test.com", "testpass", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	c.client.Transport = &staticRoundTripper{err: errors.New("connection refused")}

	f := &casesFetcher{client: c}
	_, _, err = f.FetchPageCtx(context.Background(), concurrency.PageRequest{
		SuiteTask: concurrency.SuiteTask{ProjectID: 1, SuiteID: 2},
		Offset:    0,
		Limit:     250,
	})
	if err == nil {
		t.Fatal("expected transport error, got nil")
	}
}


func TestFetchPageCtx_ReadBodyError(t *testing.T) {
client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
})
defer server.Close()

client.client.Transport = &staticRoundTripper{resp: &http.Response{
StatusCode: http.StatusOK,
Status:     "200 OK",
Body:       &errReadCloser{err: errors.New("body read error")},
Header:     make(http.Header),
}}

f := &casesFetcher{client: client}
_, _, err := f.FetchPageCtx(context.Background(), concurrency.PageRequest{
		SuiteTask: concurrency.SuiteTask{ProjectID: 1, SuiteID: 1},
		Offset:    0,
		Limit:     250,
})
if err == nil {
t.Fatal("expected error from body read failure, got nil")
}
}

// ---------- GetCasesParallelCtx: err != nil && len(result.Cases) == 0 ----------

func TestGetCasesParallelCtx_AllFail_ReturnsError(t *testing.T) {
	// Pre-cancel the context so controller.Execute returns a context error with no cases.
	// This exercises the `if err != nil && len(result.Cases) == 0` branch.
	client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel so controller returns ctx.Err() from suiteWorker

	cfg := &concurrency.ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxRetries:          0,
		RequestsPerMinute:   0,
	}
	_, result, err := client.GetCasesParallelCtx(ctx, 1, []int64{10}, cfg)
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
	if result == nil {
		t.Fatal("expected non-nil ExecutionResult even on error")
	}
}

// ---------- GetCasesParallel: partial failure ----------

func TestGetCasesParallel_PartialFailure(t *testing.T) {
client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
})
defer server.Close()

client.client.Transport = &staticRoundTripper{err: errors.New("transport fail")}

_, err := client.GetCasesParallel(context.Background(), 1, []int64{1, 2}, 2, nil)
if err == nil {
t.Fatal("expected partial failure error")
}
}

// ---------- GetSuitesParallel: partial failure ----------

func TestGetSuitesParallel_PartialFailure(t *testing.T) {
client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
})
defer server.Close()

client.client.Transport = &staticRoundTripper{err: errors.New("transport fail")}

_, err := client.GetSuitesParallel(context.Background(), []int64{1, 2}, 2, nil)
if err == nil {
t.Fatal("expected partial failure error")
}
}

// ---------- GetSectionsParallelCtx: config.Reporter != nil ----------

func TestGetSectionsParallelCtx_WithReporter(t *testing.T) {
handler := func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
"sections": []data.Section{{ID: 1, Name: "s1"}},
"_links":   map[string]interface{}{},
})
}
client, server := mockClient(t, handler)
defer server.Close()

reporter := &nopPaginatedReporter{}
cfg := &concurrency.ControllerConfig{
MaxConcurrentSuites: 1,
Reporter:            reporter,
}

sections, err := client.GetSectionsParallelCtx(context.Background(), 1, []int64{1}, cfg)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
_ = sections
}

// ---------- GetSectionsParallelCtx: rate limiter WaitCtx cancelled ----------

func TestGetSectionsParallelCtx_RateLimiterCancelledCtx(t *testing.T) {
handler := func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
"sections": []data.Section{},
"_links":   map[string]interface{}{},
})
}
client, server := mockClient(t, handler)
defer server.Close()

ctx, cancel := context.WithCancel(context.Background())
cancel()

cfg := &concurrency.ControllerConfig{
RequestsPerMinute:   1,
MaxConcurrentSuites: 1,
}

_, err := client.GetSectionsParallelCtx(ctx, 1, []int64{1}, cfg)
_ = err
}

// ---------- GetSectionsParallelCtx: retry loop with ctx cancel during delay ----------

func TestGetSectionsParallelCtx_RetryDelay_CtxCancel(t *testing.T) {
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
}))
defer srv.Close()

c, err := NewClient(srv.URL, "test@test.com", "testpass", false)
if err != nil {
t.Fatalf("failed to create client: %v", err)
}

ctx, cancel := context.WithCancel(context.Background())
go func() {
time.Sleep(30 * time.Millisecond)
cancel()
}()

cfg := &concurrency.ControllerConfig{
MaxRetries:          3,
MaxConcurrentSuites: 1,
RequestsPerMinute:   0,
}

_, err = c.GetSectionsParallelCtx(ctx, 1, []int64{1}, cfg)
if err == nil {
t.Fatal("expected error from context cancellation or retry failure")
}
}

// ---------- uploadAttachment: non-200 from DoRequest ----------

func TestUploadAttachment_NonOKResponse(t *testing.T) {
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
}))
defer srv.Close()

c, err := NewClient(srv.URL, "test@test.com", "testpass", false)
if err != nil {
t.Fatalf("failed to create client: %v", err)
}

f, err := os.CreateTemp(t.TempDir(), "attachment-*.txt")
if err != nil {
t.Fatalf("failed to create temp file: %v", err)
}
f.WriteString("test content")
f.Close()

_, uploadErr := c.uploadAttachment(context.Background(), "add_attachment_to_case/1", f.Name())
if uploadErr == nil {
t.Fatal("expected error from API non-200 response")
}
}
