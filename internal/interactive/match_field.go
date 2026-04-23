// internal/interactive/match_field.go
package interactive

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// MatchFieldKind describes which entity the match field is being chosen for.
// Different kinds ship different sets of available compare fields.
type MatchFieldKind string

const (
	MatchFieldCases        MatchFieldKind = "cases"
	MatchFieldSections     MatchFieldKind = "sections"
	MatchFieldSuites       MatchFieldKind = "suites"
	MatchFieldSharedSteps  MatchFieldKind = "shared_steps"
)

// matchFieldOptions returns the human-readable list of compare fields per
// entity kind. The first string in each tuple is the label shown to the user,
// the second is the value that must be fed to the migration layer (case-
// insensitive struct field name).
func matchFieldOptions(kind MatchFieldKind) [][2]string {
	switch kind {
	case MatchFieldCases:
		return [][2]string{
			{"title (default)", "Title"},
			{"refs", "Refs"},
			{"custom_preconds", "CustomPreconds"},
		}
	case MatchFieldSharedSteps:
		return [][2]string{
			{"title + steps hash (default)", "Title"},
		}
	case MatchFieldSections, MatchFieldSuites:
		return [][2]string{
			{"name (default)", "Name"},
		}
	default:
		return [][2]string{{"title (default)", "Title"}}
	}
}

// SelectMatchField asks the user to pick a compare field. In non-interactive
// mode (NonInteractivePrompter) or when only one option is available it
// returns defaultField unchanged. The returned value is safe to pass into the
// migration layer's compareField (matched case-insensitively against struct
// field names via reflection).
func SelectMatchField(ctx context.Context, p Prompter, kind MatchFieldKind, defaultField string) (string, error) {
	if p == nil {
		return defaultField, nil
	}
	if IsNonInteractive(ctx) {
		return defaultField, nil
	}
	// Only prompt when stdin is actually a terminal. This keeps automated
	// CLI usage (pipes, CI, test harnesses with a mocked Prompter but no
	// real TTY) from silently hanging or racing on EOF — in that case we
	// fall back to the provided default, which matches the --compare-field
	// flag value the user would have set explicitly.
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return defaultField, nil
	}
	opts := matchFieldOptions(kind)
	if len(opts) <= 1 {
		return defaultField, nil
	}

	labels := make([]string, 0, len(opts))
	for _, o := range opts {
		labels = append(labels, o[0])
	}

	idx, _, err := p.Select(fmt.Sprintf("Select match field for %s:", kind), labels)
	if err != nil {
		return "", fmt.Errorf("SelectMatchField: %w", err)
	}
	if idx < 0 || idx >= len(opts) {
		return defaultField, nil
	}
	return opts[idx][1], nil
}

// NormalizeMatchField returns the canonical (case-insensitive) form of the
// user-supplied compare field. Empty string falls back to the kind-specific
// default. Unknown values are returned as-is so downstream reflection can
// still resolve them against arbitrary struct fields.
func NormalizeMatchField(kind MatchFieldKind, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		switch kind {
		case MatchFieldSections, MatchFieldSuites:
			return "Name"
		default:
			return "Title"
		}
	}
	// Accept common aliases: the CLI flag default is lowercase "title".
	switch strings.ToLower(raw) {
	case "title":
		return "Title"
	case "name":
		return "Name"
	case "refs":
		return "Refs"
	case "custom_preconds":
		return "CustomPreconds"
	}
	return raw
}
