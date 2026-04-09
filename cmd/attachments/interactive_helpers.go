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

	return selectCaseID(ctx, cases)
}

func selectCaseID(ctx context.Context, cases data.GetCasesResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(cases))
	for i, kase := range cases {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, kase.ID, kase.Title))
	}

	idx, _, err := p.Select("Select case:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select case: %w", err)
	}

	return cases[idx].ID, nil
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

	return selectAttachmentID(ctx, attachments)
}

func selectAttachmentID(ctx context.Context, attachments data.GetAttachmentsResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(attachments))
	for i, a := range attachments {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, a.ID, a.Name))
	}

	idx, _, err := p.Select("Select attachment:", options)
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

	return selectTestID(ctx, tests)
}

func selectTestID(ctx context.Context, tests []data.Test) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(tests))
	for i, test := range tests {
		options = append(options, fmt.Sprintf("[%d] ID: %d | case: %d | %s", i+1, test.ID, test.CaseID, test.Title))
	}

	idx, _, err := p.Select("Select test:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select test: %w", err)
	}

	return tests[idx].ID, nil
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

	return selectPlanID(ctx, plans)
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

	return selectPlanEntryID(ctx, plan.Entries)
}

func selectPlanID(ctx context.Context, plans data.GetPlansResponse) (int64, error) {
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

func selectPlanEntryID(ctx context.Context, entries []data.PlanEntry) (string, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(entries))
	for i, entry := range entries {
		options = append(options, fmt.Sprintf("[%d] ID: %s | %s", i+1, entry.ID, entry.Name))
	}

	idx, _, err := p.Select("Select plan entry:", options)
	if err != nil {
		return "", fmt.Errorf("failed to select plan entry: %w", err)
	}

	return entries[idx].ID, nil
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

	return selectResultID(ctx, results)
}

func selectResultID(ctx context.Context, results data.GetResultsResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(results))
	for i, result := range results {
		options = append(options, fmt.Sprintf("[%d] ID: %d | status: %d | test: %d", i+1, result.ID, result.StatusID, result.TestID))
	}

	idx, _, err := p.Select("Select result:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select result: %w", err)
	}

	return results[idx].ID, nil
}
