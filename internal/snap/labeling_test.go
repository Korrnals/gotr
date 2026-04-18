package snap

import (
	"strings"
	"testing"
)

func TestGenerateDefaultLabel(t *testing.T) {
	label := GenerateDefaultLabel(ModeInteractive, "sync_full")

	if !strings.HasPrefix(label, "interactive_sync_full_") {
		t.Errorf("expected label to start with 'interactive_sync_full_', got %s", label)
	}

	if len(label) != len("interactive_sync_full_20060102150405") {
		t.Errorf("unexpected label length: %d", len(label))
	}
}

func TestGenerateDefaultLabelModes(t *testing.T) {
	tests := []struct {
		mode     LabelMode
		expected string
	}{
		{ModeBatch, "batch_"},
		{ModeInteractive, "interactive_"},
		{ModeAuto, "auto_"},
	}

	for _, tt := range tests {
		label := GenerateDefaultLabel(tt.mode, "test_cmd")
		if !strings.HasPrefix(label, tt.expected) {
			t.Errorf("mode %s: expected prefix %s, got %s", tt.mode, tt.expected, label)
		}
	}
}

func TestValidateLabel(t *testing.T) {
	tests := []struct {
		label     string
		shouldErr bool
	}{
		{"valid_label_123", false},
		{"valid-label-456", false},
		{"ValidLabel", false},
		{"", true}, // empty
		{strings.Repeat("a", 101), true}, // too long
		{"invalid@label", true},           // invalid char
		{"invalid#label", true},           // invalid char
		{"invalid label", true},           // space
	}

	for _, tt := range tests {
		err := ValidateLabel(tt.label)
		if (err != nil) != tt.shouldErr {
			t.Errorf("label %q: expected error=%v, got error=%v", tt.label, tt.shouldErr, err)
		}
	}
}

func TestIsProtectedLabel(t *testing.T) {
	reserved := []string{"pinned_", "archived_", "system_"}

	tests := []struct {
		label      string
		isProtected bool
	}{
		{"pinned_backup", true},
		{"archived_old", true},
		{"system_auto", true},
		{"user_sync", false},
		{"manual_backup", false},
		{"pinned", false},
	}

	for _, tt := range tests {
		result := IsProtectedLabel(tt.label, reserved)
		if result != tt.isProtected {
			t.Errorf("label %q: expected protected=%v, got %v", tt.label, tt.isProtected, result)
		}
	}
}

func TestNormalizeCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sync full", "syncfull"},
		{"sync-full", "sync-full"},
		{"sync_full", "sync_full"},
		{"Sync-Full", "Sync-Full"},
		{"sync/full/test", "syncfulltest"},
		{"compare@v2", "comparev2"},
	}

	for _, tt := range tests {
		result := NormalizeCommand(tt.input)
		if result != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}
