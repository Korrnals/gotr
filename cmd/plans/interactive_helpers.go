package plans

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
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

	return selectPlanID(ctx, plans)
}

func selectPlanID(ctx context.Context, plans []data.Plan) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(plans))
	for i, plan := range plans {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, plan.ID, plan.Name))
	}

	idx, _, err := p.Select("Select plan:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select plan: %w", err)
	}

	return plans[idx].ID, nil
}
