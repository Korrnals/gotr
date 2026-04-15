package interactive

import "context"

// BrowseConfig configures a single-level browse interaction.
type BrowseConfig struct {
	// Prompt is the selection prompt text.
	Prompt string
	// Header is printed above options (e.g. column headers). Optional.
	Header string
	// Items are the selectable option labels.
	Items []string
	// AllowBack adds "← Back" at top and bottom of the list.
	AllowBack bool
}

// Browse presents a single-level selection with optional Back/Exit.
// Returns the 0-based index into cfg.Items, or ErrGoBack/ErrExit.
func Browse(ctx context.Context, p Prompter, cfg BrowseConfig) (int, error) {
	if len(cfg.Items) == 0 {
		return 0, ErrGoBack
	}

	// Build option list: [Back] + Exit + items + [Back].
	options := make([]string, 0, len(cfg.Items)+4)
	if cfg.AllowBack {
		options = append(options, OptBack)
	}
	options = append(options, OptExit)
	itemStart := len(options)
	options = append(options, cfg.Items...)
	if cfg.AllowBack && len(cfg.Items) > 5 {
		options = append(options, OptBack)
	}

	// Print header if provided.
	if cfg.Header != "" {
		// Header is informational — cannot be selected.
		// We'll show it as part of the prompt.
		cfg.Prompt = cfg.Prompt + "\n  " + cfg.Header
	}

	idx, _, err := p.Select(cfg.Prompt, options)
	if err != nil {
		if IsInterrupt(err) {
			return 0, ErrExit
		}
		return 0, err
	}

	selected := options[idx]
	switch selected {
	case OptBack:
		return 0, ErrGoBack
	case OptExit:
		return 0, ErrExit
	default:
		return idx - itemStart, nil
	}
}

// IsInterrupt checks if the error is a context cancellation (Ctrl+C).
func IsInterrupt(err error) bool {
	return err == context.Canceled
}
