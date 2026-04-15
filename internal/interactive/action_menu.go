package interactive

import "context"

// ActionOption is a single option in a post-action menu.
type ActionOption struct {
	// Label is the display text (e.g. "↻ Rollback this migration").
	Label string
	// Key is a short identifier returned when selected (e.g. "rollback").
	Key string
	// Disabled makes the option visible but not selectable.
	Disabled bool
	// Hint is shown next to disabled options (e.g. "already rolled back").
	Hint string
}

// ActionMenu presents a post-action menu and returns the selected option key.
// Disabled options trigger a re-prompt loop. Ctrl+C returns ErrExit.
func ActionMenu(ctx context.Context, p Prompter, prompt string, options []ActionOption) (string, error) {
	if len(options) == 0 {
		return "", ErrExit
	}

	labels := make([]string, len(options))
	for i, o := range options {
		if o.Disabled && o.Hint != "" {
			labels[i] = o.Label + " (" + o.Hint + ")"
		} else {
			labels[i] = o.Label
		}
	}

	for {
		idx, _, err := p.Select(prompt, labels)
		if err != nil {
			if IsInterrupt(err) {
				return "", ErrExit
			}
			return "", err
		}

		opt := options[idx]
		if opt.Disabled {
			// Re-prompt — user selected a disabled option.
			continue
		}

		return opt.Key, nil
	}
}
