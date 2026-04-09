package roles

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
)

func resolveRoleIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	roles, err := cli.GetRoles(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get roles list: %w", err)
	}
	if len(roles) == 0 {
		return 0, fmt.Errorf("no roles found")
	}
	items := make([]string, len(roles))
	for i, role := range roles {
		items[i] = fmt.Sprintf("[%d] ID: %d | %s", i+1, role.ID, role.Name)
	}
	idx, _, err := p.Select("Select role:", items)
	if err != nil {
		return 0, fmt.Errorf("failed to select role: %w", err)
	}
	return roles[idx].ID, nil
}

func requireInteractiveRoleArg(ctx context.Context, usage string) error {
	if !interactive.HasPrompterInContext(ctx) {
		return fmt.Errorf("required argument is missing in non-interactive mode: %s", usage)
	}
	if interactive.IsNonInteractive(ctx) {
		return fmt.Errorf("required argument is missing in non-interactive mode: %s", usage)
	}
	return nil
}