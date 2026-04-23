package interactive

import (
	"context"
	"fmt"
)

// SelectServer presents the configured server URL for confirmation.
// If the URL is empty, it returns an error asking the user to configure first.
// Returns the confirmed base URL or an error (including ErrGoBack/ErrExit).
func SelectServer(ctx context.Context, p Prompter, baseURL string) (string, error) {
	if baseURL == "" {
		return "", fmt.Errorf("no server configured; run 'gotr config init' first")
	}

	ok, err := p.Confirm(fmt.Sprintf("⚡ You are connecting to: %s. Continue?", baseURL), true)
	if err != nil {
		if IsGoBack(err) || IsExit(err) || IsInterrupt(err) {
			return "", err
		}
		return "", fmt.Errorf("server confirmation: %w", err)
	}
	if !ok {
		return "", ErrExit
	}

	return baseURL, nil
}
