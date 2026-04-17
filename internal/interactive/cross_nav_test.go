package interactive

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
)

func TestCrossNavOptions(t *testing.T) {
	opts := CrossNavOptions()
	if len(opts) != 3 {
		t.Fatalf("CrossNavOptions: want 3 options, got %d", len(opts))
	}
	wantKeys := []string{"nav:compare", "nav:sync", "nav:snap"}
	for i, want := range wantKeys {
		if opts[i].Key != want {
			t.Errorf("opts[%d].Key = %q, want %q", i, opts[i].Key, want)
		}
	}
}

func TestIsCrossNav(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"nav:compare", true},
		{"nav:sync", true},
		{"nav:snap", true},
		{"nav:unknown", true},
		{"exit", false},
		{"back", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsCrossNav(tt.key); got != tt.want {
			t.Errorf("IsCrossNav(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestHandleCrossNav_NonNavKey(t *testing.T) {
	ctx := context.Background()
	cmd := &cobra.Command{Use: "test"}
	if HandleCrossNav(ctx, cmd, "exit") {
		t.Error("HandleCrossNav should return false for non-nav key")
	}
}

func TestHandleCrossNav_UnknownTarget(t *testing.T) {
	ctx := context.Background()
	root := &cobra.Command{Use: "gotr"}
	child := &cobra.Command{Use: "test"}
	root.AddCommand(child)
	child.SetContext(ctx)

	// Unknown nav target — should return true (handled) but not panic.
	if !HandleCrossNav(ctx, child, "nav:unknown") {
		t.Error("HandleCrossNav should return true for nav: prefixed key")
	}
}
