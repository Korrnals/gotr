package interactive

import (
	"context"
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

var surveyAskOne = survey.AskOne

// ErrNonInteractive is returned when a prompt is requested in non-interactive mode.
var ErrNonInteractive = errors.New("non-interactive mode: input required but unavailable")

// Prompter is a unified interface for all interactive input in commands.
type Prompter interface {
	Input(message, defaultVal string) (string, error)
	Confirm(message string, def bool) (bool, error)
	Select(message string, options []string) (idx int, value string, err error)
	MultilineInput(message, defaultVal string) (string, error)
}

// TerminalPrompter is survey/v2-based implementation used in normal CLI mode.
type TerminalPrompter struct{}

// NewTerminalPrompter creates a terminal prompter.
func NewTerminalPrompter() Prompter {
	return &TerminalPrompter{}
}

// Input asks for a single-line user input.
func (p *TerminalPrompter) Input(message, defaultVal string) (string, error) {
	var answer string
	prompt := &survey.Input{Message: message}
	if defaultVal != "" {
		prompt.Default = defaultVal
	}

	err := surveyAskOne(prompt, &answer)
	if err != nil {
		return "", fmt.Errorf("failed to get input: %w", err)
	}

	return answer, nil
}

// Confirm asks for yes/no confirmation.
func (p *TerminalPrompter) Confirm(message string, def bool) (bool, error) {
	var answer bool
	err := surveyAskOne(&survey.Confirm{Message: message, Default: def}, &answer)
	if err != nil {
		return false, fmt.Errorf("failed to get confirmation: %w", err)
	}

	return answer, nil
}

// Select asks user to choose one option from the provided list.
func (p *TerminalPrompter) Select(message string, options []string) (idx int, value string, err error) {
	if len(options) == 0 {
		return 0, "", fmt.Errorf("select options list is empty")
	}

	var selected string
	err = surveyAskOne(&survey.Select{Message: message, Options: options}, &selected)
	if err != nil {
		return 0, "", fmt.Errorf("failed to select option: %w", err)
	}

	for i, option := range options {
		if option == selected {
			return i, selected, nil
		}
	}

	return 0, "", fmt.Errorf("selected option is not in list")
}

// MultilineInput asks for multi-line input.
func (p *TerminalPrompter) MultilineInput(message, defaultVal string) (string, error) {
	var answer string
	prompt := &survey.Multiline{Message: message, Default: defaultVal}
	err := surveyAskOne(prompt, &answer)
	if err != nil {
		return "", fmt.Errorf("failed to get multiline input: %w", err)
	}

	return answer, nil
}

// NonInteractivePrompter always fails and is used for --non-interactive mode.
type NonInteractivePrompter struct{}

// NewNonInteractivePrompter creates a prompter that blocks all prompt operations.
func NewNonInteractivePrompter() Prompter {
	return &NonInteractivePrompter{}
}

// Input returns ErrNonInteractive in non-interactive mode.
func (p *NonInteractivePrompter) Input(message, defaultVal string) (string, error) {
	return "", ErrNonInteractive
}

// Confirm returns ErrNonInteractive in non-interactive mode.
func (p *NonInteractivePrompter) Confirm(message string, def bool) (bool, error) {
	return false, ErrNonInteractive
}

// Select returns ErrNonInteractive in non-interactive mode.
func (p *NonInteractivePrompter) Select(message string, options []string) (idx int, value string, err error) {
	return 0, "", ErrNonInteractive
}

// MultilineInput returns ErrNonInteractive in non-interactive mode.
func (p *NonInteractivePrompter) MultilineInput(message, defaultVal string) (string, error) {
	return "", ErrNonInteractive
}

type prompterContextKey struct{}

// WithPrompter attaches a prompter to context.
func WithPrompter(ctx context.Context, p Prompter) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, prompterContextKey{}, p)
}

// HasPrompterInContext reports whether context contains an explicitly injected prompter.
func HasPrompterInContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	p, ok := ctx.Value(prompterContextKey{}).(Prompter)
	return ok && p != nil
}

// PrompterFromContext returns context-attached prompter or terminal default.
func PrompterFromContext(ctx context.Context) Prompter {
	if ctx == nil {
		return NewTerminalPrompter()
	}

	p, ok := ctx.Value(prompterContextKey{}).(Prompter)
	if ok && p != nil {
		return p
	}

	return NewTerminalPrompter()
}

// IsNonInteractive reports whether current context uses non-interactive prompter.
func IsNonInteractive(ctx context.Context) bool {
	_, ok := PrompterFromContext(ctx).(*NonInteractivePrompter)
	return ok
}
