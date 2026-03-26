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

func resolvePlanEntryIDInteractive(ctx context.Context, cli client.ClientInterface, planID int64) (string, error) {
	plan, err := cli.GetPlan(ctx, planID)
	if err != nil {
		return "", fmt.Errorf("failed to get plan %d: %w", planID, err)
	}
	if len(plan.Entries) == 0 {
		return "", fmt.Errorf("no entries found in plan %d", planID)
	}

	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(plan.Entries))
	for i, entry := range plan.Entries {
		options = append(options, fmt.Sprintf("[%d] ID: %s | %s", i+1, entry.ID, entry.Name))
	}

	idx, _, err := p.Select("Select plan entry:", options)
	if err != nil {
		return "", fmt.Errorf("failed to select plan entry: %w", err)
	}

	return plan.Entries[idx].ID, nil
}
