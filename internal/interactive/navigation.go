package interactive

import "errors"

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
