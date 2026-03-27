package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
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
