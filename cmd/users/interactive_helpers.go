package users

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
)

// resolveUserIDInteractive prompts the user to select a user by ID.
func resolveUserIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	users, err := cli.GetUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get users: %w", err)
	}
	if len(users) == 0 {
		return 0, fmt.Errorf("no users found")
	}

	cols := []interactive.Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Name"},
		{Header: "Email"},
	}
	rows := make([][]string, len(users))
	for i, user := range users {
		rows[i] = []string{fmt.Sprintf("%d", user.ID), user.Name, user.Email}
	}
	header, items := interactive.AlignedLabels(cols, rows)

	idx, err := interactive.Browse(ctx, p, interactive.BrowseConfig{
		Prompt: "Select user:",
		Header: header,
		Items:  items,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to select user: %w", err)
	}
	return users[idx].ID, nil
}

// resolveEmailInteractive prompts the user to select a user and returns their email.
func resolveEmailInteractive(ctx context.Context, cli client.ClientInterface) (string, error) {
	p := interactive.PrompterFromContext(ctx)
	users, err := cli.GetUsers(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get users: %w", err)
	}
	if len(users) == 0 {
		return "", fmt.Errorf("no users found")
	}

	cols := []interactive.Column{
		{Header: "Email"},
		{Header: "Name"},
	}
	rows := make([][]string, len(users))
	for i, user := range users {
		rows[i] = []string{user.Email, user.Name}
	}
	header, items := interactive.AlignedLabels(cols, rows)

	idx, err := interactive.Browse(ctx, p, interactive.BrowseConfig{
		Prompt: "Select user:",
		Header: header,
		Items:  items,
	})
	if err != nil {
		return "", fmt.Errorf("failed to select user: %w", err)
	}
	return users[idx].Email, nil
}

// requireInteractiveUserArg returns an error if interactive mode is unavailable.
func requireInteractiveUserArg(ctx context.Context, usage string) error {
	if !interactive.HasPrompterInContext(ctx) {
		return fmt.Errorf("required argument is missing in non-interactive mode: %s", usage)
	}
	if interactive.IsNonInteractive(ctx) {
		return fmt.Errorf("required argument is missing in non-interactive mode: %s", usage)
	}
	return nil
}

// userDisplayName returns the user's name or falls back to email.
func userDisplayName(user data.User) string {
	if user.Name != "" {
		return user.Name
	}
	return user.Email
}