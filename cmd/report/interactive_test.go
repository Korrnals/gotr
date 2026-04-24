package report

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

func newCmdWithCtx(ctx context.Context) *cobra.Command {
	c := &cobra.Command{}
	c.SetContext(ctx)
	return c
}

func TestResolveShowTarget_NonInteractive_NoArgs_Error(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter())
	cmd := newCmdWithCtx(ctx)
	_, err := resolveShowTarget(cmd, nil, t.TempDir())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got %q", err.Error())
	}
}

func TestResolveShowTarget_Args_ReturnsDirect(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter())
	cmd := newCmdWithCtx(ctx)
	got, err := resolveShowTarget(cmd, []string{"foo.md"}, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "foo.md" {
		t.Errorf("got %q, want %q", got, "foo.md")
	}
}

func TestPromptForReport_ReturnsSelected(t *testing.T) {
	reportsDir := setHome(t)
	// categorized entry
	rel := filepath.Join("migrations", "default", "2026-04", "r.md")
	full := filepath.Join(reportsDir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	mock := interactive.NewMockPrompter().WithSelectResponses(
		interactive.SelectResponse{Index: 1, Raw: true}, // skip "latest" (index 0)
	)
	ctx := interactive.WithPrompter(context.Background(), mock)
	cmd := newCmdWithCtx(ctx)

	got, err := promptForReport(cmd, reportsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Rel path uses slash separators via RecursiveListReports -> filepath.ToSlash.
	want := "migrations/default/2026-04/r.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPromptForReport_EmptyDir(t *testing.T) {
	reportsDir := setHome(t)
	mock := interactive.NewMockPrompter()
	ctx := interactive.WithPrompter(context.Background(), mock)
	cmd := newCmdWithCtx(ctx)

	_, err := promptForReport(cmd, reportsDir)
	if err == nil {
		t.Fatal("expected error for empty dir")
	}
	if err != ErrNoInteractiveTarget {
		t.Errorf("got %v, want ErrNoInteractiveTarget", err)
	}
}
