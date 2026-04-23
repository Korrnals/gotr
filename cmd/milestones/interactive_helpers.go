package milestones

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

func resolveMilestoneIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	projectID, err := resolveProjectIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	milestones, err := cli.GetMilestones(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get milestones for project %d: %w", projectID, err)
	}
	if len(milestones) == 0 {
		return 0, fmt.Errorf("no milestones found in project %d", projectID)
	}

	return selectMilestoneID(ctx, milestones)
}

func selectMilestoneID(ctx context.Context, milestones []data.Milestone) (int64, error) {
	p := interactive.PrompterFromContext(ctx)

	cols := []interactive.Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Name"},
	}
	rows := make([][]string, len(milestones))
	for i, m := range milestones {
		rows[i] = []string{fmt.Sprintf("%d", m.ID), m.Name}
	}
	header, options := interactive.AlignedLabels(cols, rows)

	idx, err := interactive.Browse(ctx, p, interactive.BrowseConfig{
		Prompt: "Select milestone:",
		Header: header,
		Items:  options,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to select milestone: %w", err)
	}

	return milestones[idx].ID, nil
}
