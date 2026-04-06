package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type trackingReadCloser struct {
	io.Reader
	closed *bool
}

func (trc *trackingReadCloser) Close() error {
	*trc.closed = true
	return nil
}

func TestReadJSONResponse_ClosesBodyOnErrorStatus(t *testing.T) {
	closed := false
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Status:     "500 Internal Server Error",
		Body: &trackingReadCloser{
			Reader: strings.NewReader(`{"error":"boom"}`),
			closed: &closed,
		},
	}

	client := &HTTPClient{}
	var target map[string]any

	err := client.ReadJSONResponse(context.Background(), resp, &target)
	if err == nil {
		t.Fatal("expected API error, got nil")
	}
	if !closed {
		t.Fatal("expected response body to be closed on non-200 status")
	}
}

func TestReadJSONResponse_NilResponse(t *testing.T) {
	client := &HTTPClient{}
	var target map[string]any

	err := client.ReadJSONResponse(context.Background(), nil, &target)
	if err == nil {
		t.Fatal("expected nil response error")
	}
	if !strings.Contains(err.Error(), "nil response") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadJSONResponse_NilBody(t *testing.T) {
	client := &HTTPClient{}
	resp := &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: nil}
	var target map[string]any

	err := client.ReadJSONResponse(context.Background(), resp, &target)
	if err == nil {
		t.Fatal("expected nil response body error")
	}
	if !strings.Contains(err.Error(), "nil response body") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadJSONResponse_ClosesBodyOnSuccessStatus(t *testing.T) {
	closed := false
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body: &trackingReadCloser{
			Reader: strings.NewReader(`{"id":123}`),
			closed: &closed,
		},
	}

	client := &HTTPClient{}
	target := map[string]any{}

	err := client.ReadJSONResponse(context.Background(), resp, &target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !closed {
		t.Fatal("expected response body to be closed on success status")
	}
}

func TestReadJSONResponse_NonOKReadError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Status:     "500 Internal Server Error",
		Body:       requestReadErrorCloser{},
	}

	client := &HTTPClient{}
	var target map[string]any

	err := client.ReadJSONResponse(context.Background(), resp, &target)
	if err == nil {
		t.Fatal("expected API error when reading error body fails")
	}
	if !strings.Contains(err.Error(), "failed to read error body") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}

	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = original

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error: %v", err)
	}
	return string(out)
}

func TestReadResponse(t *testing.T) {
	client := &HTTPClient{}

	t.Run("json body", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"X-Test": []string{"1"}},
			Body:       io.NopCloser(strings.NewReader(`{"id":123,"name":"ok"}`)),
		}

		data, err := client.ReadResponse(context.Background(), resp, 42*time.Millisecond, "json")
		if err != nil {
			t.Fatalf("ReadResponse() error = %v", err)
		}
		if data.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d, want %d", data.StatusCode, http.StatusOK)
		}
		if data.Duration != 42*time.Millisecond {
			t.Fatalf("Duration = %v, want %v", data.Duration, 42*time.Millisecond)
		}
		if _, ok := data.Body.(map[string]interface{}); !ok {
			t.Fatalf("expected parsed JSON object in Body, got %T", data.Body)
		}
	})

	t.Run("non-json body fallback", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     "400 Bad Request",
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("plain text body")),
		}

		data, err := client.ReadResponse(context.Background(), resp, time.Second, "json")
		if err != nil {
			t.Fatalf("ReadResponse() error = %v", err)
		}
		body, ok := data.Body.(string)
		if !ok {
			t.Fatalf("expected string fallback in Body, got %T", data.Body)
		}
		if body != "plain text body" {
			t.Fatalf("Body = %q, want %q", body, "plain text body")
		}
	})
}

func TestPrintResponseFromData(t *testing.T) {
	client := &HTTPClient{}
	data := ResponseData{
		Status:     "200 OK",
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       map[string]interface{}{"id": 1, "name": "demo"},
		Timestamp:  time.Unix(0, 0),
		Duration:   10 * time.Millisecond,
	}

	t.Run("json", func(t *testing.T) {
		out := captureStdout(t, func() {
			client.PrintResponseFromData(context.Background(), data, "json")
		})
		if !strings.Contains(out, "\"name\": \"demo\"") {
			t.Fatalf("json output missing body fields: %s", out)
		}
	})

	t.Run("json-full", func(t *testing.T) {
		out := captureStdout(t, func() {
			client.PrintResponseFromData(context.Background(), data, "json-full")
		})
		if !strings.Contains(out, "\"status\": \"200 OK\"") {
			t.Fatalf("json-full output missing status field: %s", out)
		}
	})

	t.Run("default table", func(t *testing.T) {
		out := captureStdout(t, func() {
			client.PrintResponseFromData(context.Background(), data, "table")
		})
		if !strings.Contains(out, "Status: 200 OK (200)") {
			t.Fatalf("table output missing status line: %s", out)
		}
		if !strings.Contains(out, "Body:") {
			t.Fatalf("table output missing body section: %s", out)
		}
	})
}

func TestSaveResponseToFile(t *testing.T) {
	client := &HTTPClient{}
	data := ResponseData{
		Status:     "200 OK",
		StatusCode: 200,
		Headers:    http.Header{"X-Test": []string{"1"}},
		Body:       map[string]interface{}{"k": "v"},
		Timestamp:  time.Unix(0, 0),
		Duration:   time.Second,
	}

	t.Run("json", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "resp-json.json")
		if err := client.SaveResponseToFile(context.Background(), data, file, "json"); err != nil {
			t.Fatalf("SaveResponseToFile(json) error = %v", err)
		}
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		var body map[string]interface{}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if body["k"] != "v" {
			t.Fatalf("unexpected saved body: %+v", body)
		}
	})

	t.Run("json-full and default", func(t *testing.T) {
		fileFull := filepath.Join(t.TempDir(), "resp-full.json")
		if err := client.SaveResponseToFile(context.Background(), data, fileFull, "json-full"); err != nil {
			t.Fatalf("SaveResponseToFile(json-full) error = %v", err)
		}

		raw, err := os.ReadFile(fileFull)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		var full map[string]interface{}
		if err := json.Unmarshal(raw, &full); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if full["status"] != "200 OK" {
			t.Fatalf("unexpected full payload: %+v", full)
		}

		fileDefault := filepath.Join(t.TempDir(), "resp-default.json")
		if err := client.SaveResponseToFile(context.Background(), data, fileDefault, "table"); err != nil {
			t.Fatalf("SaveResponseToFile(default) error = %v", err)
		}
		if _, err := os.Stat(fileDefault); err != nil {
			t.Fatalf("expected default file to be written: %v", err)
		}
	})
}

type requestReadErrorCloser struct{}

func (requestReadErrorCloser) Read([]byte) (int, error) { return 0, errors.New("read boom") }
func (requestReadErrorCloser) Close() error              { return nil }

func TestReadResponse_ReadError(t *testing.T) {
	client := &HTTPClient{}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       requestReadErrorCloser{},
		Header:     http.Header{},
	}

	_, err := client.ReadResponse(context.Background(), resp, time.Millisecond, "json")
	if err == nil {
		t.Fatal("expected ReadResponse error")
	}
}

func TestSaveResponseToFile_WriteError(t *testing.T) {
	client := &HTTPClient{}
	data := ResponseData{Body: map[string]any{"k": "v"}}

	err := client.SaveResponseToFile(context.Background(), data, t.TempDir(), "json")
	if err == nil {
		t.Fatal("expected SaveResponseToFile error")
	}
}
