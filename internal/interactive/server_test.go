package interactive

import (
	"context"
	"testing"
)

func TestSelectServer_EmptyURL(t *testing.T) {
	p := NewMockPrompter()
	ctx := WithPrompter(context.Background(), p)

	_, err := SelectServer(ctx, p, "")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
	if err.Error() != "no server configured; run 'gotr config init' first" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSelectServer_Confirmed(t *testing.T) {
	p := NewMockPrompter().WithConfirmResponses(true)
	ctx := WithPrompter(context.Background(), p)

	url, err := SelectServer(ctx, p, "https://example.testrail.io")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://example.testrail.io" {
		t.Errorf("got %q, want %q", url, "https://example.testrail.io")
	}
}

func TestSelectServer_Rejected(t *testing.T) {
	p := NewMockPrompter().WithConfirmResponses(false)
	ctx := WithPrompter(context.Background(), p)

	_, err := SelectServer(ctx, p, "https://example.testrail.io")
	if !IsExit(err) {
		t.Errorf("expected ErrExit, got %v", err)
	}
}
