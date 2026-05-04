package snap

import (
	"fmt"
	"time"
)

// LabelMode represents the type of snapshot creation
type LabelMode string

const (
	ModeBatch       LabelMode = "batch"       // Non-interactive, automated execution
	ModeInteractive LabelMode = "interactive" // Interactive user session
	ModeAuto        LabelMode = "auto"        // System-generated (auto-compare, etc)
)

// LabelInfo holds information for label generation
type LabelInfo struct {
	Mode    LabelMode // batch, interactive, auto
	Command string    // sync_full, sync_cases, compare, etc
	Created time.Time
}

// GenerateDefaultLabel generates a default label in format: {mode}_{command}_{timestamp}
// Example: "interactive_sync_full_20260418143015"
//
// The timestamp is formatted in UTC to stay consistent with snapshot IDs
// (see internal/snap/store.go), so labels and IDs cannot disagree on
// calendar day across timezones.
func GenerateDefaultLabel(mode LabelMode, command string) string {
	timestamp := time.Now().UTC().Format("20060102150405")
	return fmt.Sprintf("%s_%s_%s", mode, command, timestamp)
}

// ValidateLabel checks if label is valid (alphanumeric + underscore + dash)
func ValidateLabel(label string) error {
	if label == "" {
		return fmt.Errorf("label cannot be empty")
	}

	if len(label) > 100 {
		return fmt.Errorf("label too long (max 100 characters)")
	}

	for _, ch := range label {
		if !isValidLabelChar(ch) {
			return fmt.Errorf("label contains invalid character: %c (only alphanumeric, underscore, dash allowed)", ch)
		}
	}

	return nil
}

// IsProtectedLabel checks if label has a reserved prefix
func IsProtectedLabel(label string, reservedPrefixes []string) bool {
	for _, prefix := range reservedPrefixes {
		if len(label) >= len(prefix) && label[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func isValidLabelChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' ||
		ch == '-'
}

// NormalizeCommand normalizes a command string for use in labels
// Removes special characters and converts to lowercase
func NormalizeCommand(cmd string) string {
	result := ""
	for _, ch := range cmd {
		if isValidLabelChar(ch) {
			result += string(ch)
		}
	}
	return result
}
