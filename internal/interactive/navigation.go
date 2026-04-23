package interactive

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Navigation sentinel errors for multi-level interactive flows.
var (
	// ErrGoBack signals the user chose "← Back" to return to the previous level.
	ErrGoBack = errors.New("go back")

	// ErrExit signals the user chose "✕ Exit" to leave the interactive flow.
	ErrExit = errors.New("exit")
)

// Navigation option labels (consistent across all commands).
const (
	OptBack = "← Back"
	OptExit = "✕ Exit"
)

// IsGoBack returns true if the error is a back-navigation sentinel.
func IsGoBack(err error) bool {
	return errors.Is(err, ErrGoBack)
}

// IsExit returns true if the error is an exit sentinel.
func IsExit(err error) bool {
	return errors.Is(err, ErrExit)
}

// FindSubcommand locates a subcommand by path via Cobra tree traversal.
// Returns an error if the command is not found.
func FindSubcommand(root *cobra.Command, path ...string) (*cobra.Command, error) {
	target, _, err := root.Find(path)
	if err != nil || target == nil || target.Name() == root.Name() {
		return nil, fmt.Errorf("could not find 'gotr %s' command", strings.Join(path, " "))
	}
	return target, nil
}

// RunSubcommand finds a subcommand by path, propagates context, and runs it.
// GoBack/Exit sentinel errors are absorbed (return nil).
// Real execution errors are returned as-is.
func RunSubcommand(ctx context.Context, root *cobra.Command, path ...string) error {
	target, err := FindSubcommand(root, path...)
	if err != nil {
		return fmt.Errorf("RunSubcommand: %w", err)
	}
	target.SetContext(ctx)
	if err := target.RunE(target, nil); err != nil {
		if IsGoBack(err) || IsExit(err) {
			return nil
		}
		return fmt.Errorf("RunSubcommand: %w", err)
	}
	return nil
}
