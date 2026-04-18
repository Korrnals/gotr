package snap

import (
	"testing"
)

func TestPromptResult(t *testing.T) {
	// Test that PromptResult can be created with various actions
	tests := []struct {
		action string
		pinned bool
	}{
		{"use_default", false},
		{"custom", false},
		{"pin", true},
		{"skip", false},
	}

	for _, tt := range tests {
		result := PromptResult{
			Label:  "test_label",
			Action: tt.action,
			Pinned: tt.pinned,
		}

		if result.Label != "test_label" {
			t.Errorf("expected label test_label, got %s", result.Label)
		}
		if result.Action != tt.action {
			t.Errorf("expected action %s, got %s", tt.action, result.Action)
		}
		if result.Pinned != tt.pinned {
			t.Errorf("expected pinned %v, got %v", tt.pinned, result.Pinned)
		}
	}
}

func TestInteractivePromptOptionsValidation(t *testing.T) {
	// Test that options can be created and used
	opts := InteractivePromptOptions{
		DefaultLabel: "interactive_sync_full_20260418143015",
		AllowPin:     true,
		AllowSkip:    true,
	}

	if opts.DefaultLabel == "" {
		t.Error("default label should not be empty")
	}

	if !opts.AllowPin {
		t.Error("AllowPin should be true")
	}

	if !opts.AllowSkip {
		t.Error("AllowSkip should be true")
	}
}
