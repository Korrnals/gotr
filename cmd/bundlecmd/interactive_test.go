package bundlecmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/spf13/cobra"
)

func newCmdWithCtx(ctx context.Context) *cobra.Command {
	c := &cobra.Command{}
	c.SetContext(ctx)
	return c
}

func TestResolveExportSnapTarget_NonInteractive_NoArgs_Error(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter())
	cmd := newCmdWithCtx(ctx)
	_, err := resolveExportSnapTarget(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected 'required' error, got %v", err)
	}
}

func TestResolveExportReportTarget_Args_ReturnsDirect(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter())
	cmd := newCmdWithCtx(ctx)
	got, err := resolveExportReportTarget(cmd, []string{"x.md"}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != "x.md" {
		t.Errorf("got %q, want x.md", got)
	}
}

func TestPromptExportSnap_UsesManifest(t *testing.T) {
	setHome(t)
	store, err := snap.NewStore()
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := snap.LoadManifest(store)
	if err != nil {
		t.Fatal(err)
	}
	meta := &snap.Meta{
		ID:         "sync/20260424T120000_full_p1_to_p2",
		Label:      "full_p1_to_p2",
		Category:   snap.CatSync,
		Operation:  snap.OpSyncFull,
		EntityType: "cases",
		Status:     snap.StatusAvailable,
		Timestamp:  time.Now(),
	}
	if err := manifest.Add(meta); err != nil {
		t.Fatal(err)
	}

	mock := interactive.NewMockPrompter().WithSelectResponses(
		interactive.SelectResponse{Index: 0, Raw: true},
	)
	ctx := interactive.WithPrompter(context.Background(), mock)
	cmd := newCmdWithCtx(ctx)

	got, err := promptExportSnap(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if got != meta.ID {
		t.Errorf("got %q, want %q", got, meta.ID)
	}
}

func TestPromptExportSnap_EmptyManifest(t *testing.T) {
	setHome(t)
	mock := interactive.NewMockPrompter()
	ctx := interactive.WithPrompter(context.Background(), mock)
	cmd := newCmdWithCtx(ctx)
	_, err := promptExportSnap(cmd)
	if err != ErrNoInteractiveTarget {
		t.Errorf("got %v, want ErrNoInteractiveTarget", err)
	}
}

func TestPromptImportBundle_PicksFile(t *testing.T) {
	home := setHome(t)
	dir := filepath.Join(home, ".gotr", "exports", "snaps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "snap_x.tar.gz")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	mock := interactive.NewMockPrompter().WithSelectResponses(
		interactive.SelectResponse{Index: 0, Raw: true},
	)
	ctx := interactive.WithPrompter(context.Background(), mock)
	cmd := newCmdWithCtx(ctx)

	got, err := promptImportBundle(cmd, dir, "x", ".tar.gz", ".tgz")
	if err != nil {
		t.Fatal(err)
	}
	if got != p {
		t.Errorf("got %q, want %q", got, p)
	}
}

func TestPromptImportBundle_MissingDir(t *testing.T) {
	mock := interactive.NewMockPrompter()
	ctx := interactive.WithPrompter(context.Background(), mock)
	cmd := newCmdWithCtx(ctx)
	_, err := promptImportBundle(cmd, filepath.Join(t.TempDir(), "nope"), "x", ".tar.gz")
	if err != ErrNoInteractiveTarget {
		t.Errorf("got %v, want ErrNoInteractiveTarget", err)
	}
}

func TestListBundleFiles_HandlesTarGz(t *testing.T) {
	dir := t.TempDir()
	good1 := filepath.Join(dir, "a.tar.gz")
	good2 := filepath.Join(dir, "b.tgz")
	skip := filepath.Join(dir, "c.zip")
	for _, p := range []string{good1, good2, skip} {
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	got := listBundleFiles(dir, ".tar.gz", ".tgz")
	seen := map[string]bool{}
	for _, g := range got {
		seen[g] = true
	}
	if !seen[good1] || !seen[good2] {
		t.Errorf("missing good entries: %v", got)
	}
	if seen[skip] {
		t.Errorf("unexpected %q in %v", skip, got)
	}
}
