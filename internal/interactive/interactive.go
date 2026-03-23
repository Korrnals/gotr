package interactive

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
)

// SelectProject selects a project using unified prompter.
func SelectProject(ctx context.Context, p Prompter, httpClient client.ClientInterface, prompt string) (int64, error) {
	projects, err := httpClient.GetProjects(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get projects list: %w", err)
	}

	if len(projects) == 0 {
		return 0, fmt.Errorf("no projects found")
	}

	if prompt == "" {
		prompt = "Select project:"
	}

	options := make([]string, 0, len(projects))
	for i, p := range projects {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, p.ID, p.Name))
	}

	idx, _, err := p.Select(prompt, options)
	if err != nil {
		return 0, fmt.Errorf("failed to select project: %w", err)
	}

	return projects[idx].ID, nil
}

// SelectSuite selects a suite using unified prompter.
func SelectSuite(ctx context.Context, p Prompter, suites data.GetSuitesResponse, prompt string) (int64, error) {
	_ = ctx
	if len(suites) == 0 {
		return 0, fmt.Errorf("no suites found")
	}

	if prompt == "" {
		prompt = "Select suite:"
	}

	options := make([]string, 0, len(suites))
	for i, suite := range suites {
		line := fmt.Sprintf("[%d] ID: %d | %s", i+1, suite.ID, suite.Name)
		options = append(options, line)
	}

	idx, _, err := p.Select(prompt, options)
	if err != nil {
		return 0, fmt.Errorf("failed to select suite: %w", err)
	}

	return suites[idx].ID, nil
}

// SelectRun selects a run using unified prompter.
func SelectRun(ctx context.Context, p Prompter, runs data.GetRunsResponse, prompt string) (int64, error) {
	_ = ctx
	if len(runs) == 0 {
		return 0, fmt.Errorf("no runs found")
	}

	if prompt == "" {
		prompt = "Select run:"
	}

	options := make([]string, 0, len(runs))
	for i, run := range runs {
		status := "active"
		if run.IsCompleted {
			status = "completed"
		}
		line := fmt.Sprintf("[%d] (%s) ID: %d | %s", i+1, status, run.ID, run.Name)
		options = append(options, line)
	}

	idx, _, err := p.Select(prompt, options)
	if err != nil {
		return 0, fmt.Errorf("failed to select run: %w", err)
	}

	return runs[idx].ID, nil
}

// SelectProjectInteractively is a compatibility wrapper.
func SelectProjectInteractively(ctx context.Context, httpClient client.ClientInterface) (int64, error) {
	return SelectProject(ctx, PrompterFromContext(ctx), httpClient, "")
}

// SelectSuiteInteractively is a compatibility wrapper.
func SelectSuiteInteractively(suites data.GetSuitesResponse) (int64, error) {
	return SelectSuite(context.Background(), NewTerminalPrompter(), suites, "")
}

// SelectRunInteractively is a compatibility wrapper.
func SelectRunInteractively(runs data.GetRunsResponse) (int64, error) {
	return SelectRun(context.Background(), NewTerminalPrompter(), runs, "")
}

// ConfirmAction asks for action confirmation (compatibility wrapper).
func ConfirmAction(message string) bool {
	ok, err := NewTerminalPrompter().Confirm(message, false)
	if err != nil {
		return false
	}
	return ok
}
