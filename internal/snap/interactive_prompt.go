package snap

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// InteractivePromptOptions controls interactive prompting behavior
type InteractivePromptOptions struct {
	DefaultLabel string
	AllowPin     bool
	AllowSkip    bool
}

// PromptResult contains the result of an interactive label prompt
type PromptResult struct {
	Label     string
	Action    string // "use_default", "custom", "pin", "skip"
	Pinned    bool
	Canceled  bool
}

// PromptForLabel shows an interactive menu for label selection after snapshot creation
// Returns PromptResult with user's choice
//nolint:gocyclo // Menu handling intentionally keeps all options in one place.
func PromptForLabel(ctx context.Context, opts InteractivePromptOptions) (PromptResult, error) {
	scanner := bufio.NewScanner(os.Stdin)

	// Check for non-interactive mode
	if !isInteractive() {
		return PromptResult{
			Label:  opts.DefaultLabel,
			Action: "use_default",
		}, nil
	}

	result := PromptResult{}

	fmt.Println("\n✅ Snapshot created successfully")
	fmt.Println()
	fmt.Println("Snapshot label options:")
	fmt.Printf("  [1] Use default:    %s\n", opts.DefaultLabel)
	fmt.Printf("  [2] Custom label:   _______________________________\n")
	if opts.AllowPin {
		fmt.Printf("  [3] Pin for storage (won't auto-delete)\n")
	}
	if opts.AllowSkip {
		fmt.Printf("  [4] Skip snapshot\n")
	}
	fmt.Printf("  [Q] Quit\n")
	fmt.Print("\nChoose [1-4/Q]: ")

	if !scanner.Scan() {
		result.Canceled = true
		return result, scanner.Err()
	}

	choice := strings.TrimSpace(strings.ToLower(scanner.Text()))

	switch choice {
	case "1":
		result.Label = opts.DefaultLabel
		result.Action = "use_default"

	case "2":
		fmt.Print("Enter custom label (alphanumeric, underscore, dash): ")
		if !scanner.Scan() {
			result.Canceled = true
			return result, scanner.Err()
		}

		customLabel := strings.TrimSpace(scanner.Text())
		if err := ValidateLabel(customLabel); err != nil {
			fmt.Printf("❌ Invalid label: %v\n", err)
			fmt.Printf("Using default instead: %s\n", opts.DefaultLabel)
			result.Label = opts.DefaultLabel
			result.Action = "use_default"
		} else {
			result.Label = customLabel
			result.Action = "custom"
		}

	case "3":
		if !opts.AllowPin {
			fmt.Println("Pinning not available for this operation")
			result.Label = opts.DefaultLabel
			result.Action = "use_default"
		} else {
			fmt.Print("Enter storage label prefix (default: pinned_): ")
			if !scanner.Scan() {
				result.Canceled = true
				return result, scanner.Err()
			}

			prefix := strings.TrimSpace(scanner.Text())
			if prefix == "" {
				prefix = "pinned_"
			}

			if !strings.HasSuffix(prefix, "_") {
				prefix += "_"
			}

			customLabel := fmt.Sprintf("%s%s", prefix, opts.DefaultLabel)
			if err := ValidateLabel(customLabel); err != nil {
				fmt.Printf("❌ Invalid label: %v\n", err)
				result.Label = opts.DefaultLabel
				result.Action = "use_default"
			} else {
				result.Label = customLabel
				result.Action = "pin"
				result.Pinned = true
			}
		}

	case "4":
		if !opts.AllowSkip {
			fmt.Println("Skipping not available for this operation")
			result.Label = opts.DefaultLabel
			result.Action = "use_default"
		} else {
			result.Action = "skip"
		}

	case "q":
		result.Canceled = true

	default:
		fmt.Println("Invalid choice, using default")
		result.Label = opts.DefaultLabel
		result.Action = "use_default"
	}

	return result, nil
}

// isInteractive checks if stdin is a terminal (interactive mode)
func isInteractive() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
