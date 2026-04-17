package interactive

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
)

func TestNavigateMenu_BackReturnsGoBack(t *testing.T) {
	// Select "← Back" (index 0 with raw=true)
	mock := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 0, Raw: true})
	ctx := WithPrompter(context.Background(), mock)

	root := &cobra.Command{Use: "root"}
	root.SetContext(ctx)

	err := NavigateMenu(ctx, root, "Navigate:", []NavTarget{NavCompareAll})
	if !IsGoBack(err) {
		t.Errorf("expected ErrGoBack, got %v", err)
	}
}

func TestNavigateMenu_SelectTarget(t *testing.T) {
	// Build a minimal Cobra tree with a runnable "compare all".
	root := &cobra.Command{Use: "root"}
	compare := &cobra.Command{Use: "compare"}
	ran := false
	allCmd := &cobra.Command{
		Use: "all",
		RunE: func(cmd *cobra.Command, args []string) error {
			ran = true
			return nil
		},
	}
	compare.AddCommand(allCmd)
	root.AddCommand(compare)

	// Select first data item (index 0) — Browse adds "← Back" at top, so raw index = 1.
	mock := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 0})
	ctx := WithPrompter(context.Background(), mock)
	root.SetContext(ctx)

	err := NavigateMenu(ctx, root, "Navigate:", []NavTarget{NavCompareAll})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Error("expected compare all to run")
	}
}
