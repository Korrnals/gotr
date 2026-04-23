package attachments

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
)

func resolveCaseIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}

	suites, err := cli.GetSuites(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get suites: %w", err)
	}
	if len(suites) == 0 {
		return 0, fmt.Errorf("no suites found in project %d", projectID)
	}

	suiteID, err := interactive.SelectSuite(ctx, p, suites, "")
	if err != nil {
		return 0, err
	}

	cases, err := cli.GetCases(ctx, projectID, suiteID, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get cases: %w", err)
	}
	if len(cases) == 0 {
		return 0, fmt.Errorf("no cases found in suite %d", suiteID)
	}

	return interactive.SelectCase(ctx, p, cases, "")
}

func resolveAttachmentIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	caseID, err := resolveCaseIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	attachments, err := cli.GetAttachmentsForCase(ctx, caseID)
	if err != nil {
		return 0, fmt.Errorf("failed to get attachments for case %d: %w", caseID, err)
	}
	if len(attachments) == 0 {
		return 0, fmt.Errorf("no attachments found in case %d", caseID)
	}

	p := interactive.PrompterFromContext(ctx)
	return selectAttachmentID(ctx, p, attachments)
}

func selectAttachmentID(ctx context.Context, p interactive.Prompter, attachments data.GetAttachmentsResponse) (int64, error) {
	if len(attachments) == 0 {
		return 0, fmt.Errorf("no attachments found")
	}

	cols := []interactive.Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Name"},
	}
	rows := make([][]string, len(attachments))
	for i, a := range attachments {
		rows[i] = []string{fmt.Sprintf("%d", a.ID), a.Name}
	}
	header, options := interactive.AlignedLabels(cols, rows)

	idx, err := interactive.Browse(ctx, p, interactive.BrowseConfig{
		Prompt: "Select attachment:",
		Header: header,
		Items:  options,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to select attachment: %w", err)
	}

	return attachments[idx].ID, nil
}

func resolveRunIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}

	runs, err := cli.GetRuns(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get runs for project %d: %w", projectID, err)
	}
	if len(runs) == 0 {
		return 0, fmt.Errorf("no runs found in project %d", projectID)
	}

	return interactive.SelectRun(ctx, p, runs, "")
}

func resolveTestIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	runID, err := resolveRunIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	tests, err := cli.GetTests(ctx, runID, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get tests for run %d: %w", runID, err)
	}
	if len(tests) == 0 {
		return 0, fmt.Errorf("no tests found in run %d", runID)
	}

	p := interactive.PrompterFromContext(ctx)
	return interactive.SelectTest(ctx, p, tests, "")
}

func resolvePlanIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	projectID, err := interactive.SelectProject(ctx, p, cli, "")
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

	return interactive.SelectPlan(ctx, p, plans, "")
}

func resolvePlanAndEntryIDInteractive(ctx context.Context, cli client.ClientInterface) (planID int64, entryID string, err error) {
	planID, err = resolvePlanIDInteractive(ctx, cli)
	if err != nil {
		return 0, "", err
	}

	entryID, err = resolvePlanEntryIDInteractive(ctx, cli, planID)
	if err != nil {
		return 0, "", err
	}

	return planID, entryID, nil
}

func resolvePlanEntryIDInteractive(ctx context.Context, cli client.ClientInterface, planID int64) (string, error) {
	plan, err := cli.GetPlan(ctx, planID)
	if err != nil {
		return "", fmt.Errorf("failed to get plan %d: %w", planID, err)
	}
	if len(plan.Entries) == 0 {
		return "", fmt.Errorf("no plan entries found in plan %d", planID)
	}

	p := interactive.PrompterFromContext(ctx)
	return interactive.SelectPlanEntry(ctx, p, plan.Entries, "")
}

func resolveResultIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	testID, err := resolveTestIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	results, err := cli.GetResults(ctx, testID)
	if err != nil {
		return 0, fmt.Errorf("failed to get results for test %d: %w", testID, err)
	}
	if len(results) == 0 {
		return 0, fmt.Errorf("no results found in test %d", testID)
	}

	p := interactive.PrompterFromContext(ctx)
	return selectResultID(ctx, p, results)
}

func selectResultID(ctx context.Context, p interactive.Prompter, results data.GetResultsResponse) (int64, error) {
	if len(results) == 0 {
		return 0, fmt.Errorf("no results found")
	}

	cols := []interactive.Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Status", MinWidth: 6},
		{Header: "Test", MinWidth: 6},
	}
	rows := make([][]string, len(results))
	for i, r := range results {
		rows[i] = []string{fmt.Sprintf("%d", r.ID), fmt.Sprintf("%d", r.StatusID), fmt.Sprintf("%d", r.TestID)}
	}
	header, options := interactive.AlignedLabels(cols, rows)

	idx, err := interactive.Browse(ctx, p, interactive.BrowseConfig{
		Prompt: "Select result:",
		Header: header,
		Items:  options,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to select result: %w", err)
	}

	return results[idx].ID, nil
}
