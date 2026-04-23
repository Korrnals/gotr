package plans

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
)

func resolveProjectIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	return interactive.SelectProject(ctx, p, cli, "")
}

func resolvePlanIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	projectID, err := resolveProjectIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	plans, err := cli.GetPlans(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get plans for project %d: %w", projectID, err)
	}
	if len(plans) == 0 {
		return 0, fmt.Errorf("no plans found in project %d", projectID)
	}

	p := interactive.PrompterFromContext(ctx)
	return interactive.SelectPlan(ctx, p, plans, "")
}

func resolvePlanEntryIDInteractive(ctx context.Context, cli client.ClientInterface, planID int64) (string, error) {
	plan, err := cli.GetPlan(ctx, planID)
	if err != nil {
		return "", fmt.Errorf("failed to get plan %d: %w", planID, err)
	}
	if len(plan.Entries) == 0 {
		return "", fmt.Errorf("no entries found in plan %d", planID)
	}

	p := interactive.PrompterFromContext(ctx)
	return interactive.SelectPlanEntry(ctx, p, plan.Entries, "")
}
