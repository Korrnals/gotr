package interactive

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
)

// SelectProject selects a project using unified prompter with Browse navigation.
// Pass allowBack=true to show "← Back" in addition to "✕ Exit".
func SelectProject(ctx context.Context, p Prompter, httpClient client.ClientInterface, prompt string, allowBack ...bool) (int64, error) {
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

	back := len(allowBack) > 0 && allowBack[0]
	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Items:     options,
		AllowBack: back,
	})
	if err != nil {
		if IsGoBack(err) || IsExit(err) {
			return 0, err
		}
		return 0, fmt.Errorf("failed to select project: %w", err)
	}

	return projects[idx].ID, nil
}

// SelectSuite selects a suite using unified prompter with Browse navigation.
// Pass allowBack=true to show "← Back" in addition to "✕ Exit".
func SelectSuite(ctx context.Context, p Prompter, suites data.GetSuitesResponse, prompt string, allowBack ...bool) (int64, error) {
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

	back := len(allowBack) > 0 && allowBack[0]
	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Items:     options,
		AllowBack: back,
	})
	if err != nil {
		if IsGoBack(err) || IsExit(err) {
			return 0, err
		}
		return 0, fmt.Errorf("failed to select suite: %w", err)
	}

	return suites[idx].ID, nil
}

// SelectRun selects a run using unified prompter with Browse navigation.
// Pass allowBack=true to show "← Back" in addition to "✕ Exit".
func SelectRun(ctx context.Context, p Prompter, runs data.GetRunsResponse, prompt string, allowBack ...bool) (int64, error) {
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

	back := len(allowBack) > 0 && allowBack[0]
	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Items:     options,
		AllowBack: back,
	})
	if err != nil {
		if IsGoBack(err) || IsExit(err) {
			return 0, err
		}
		return 0, fmt.Errorf("failed to select run: %w", err)
	}

	return runs[idx].ID, nil
}

// SelectSection selects a section using unified prompter with Browse navigation.
// Pass allowBack=true to show "← Back" in addition to "✕ Exit".
func SelectSection(ctx context.Context, p Prompter, sections data.GetSectionsResponse, prompt string, allowBack ...bool) (int64, error) {
	if len(sections) == 0 {
		return 0, fmt.Errorf("no sections found")
	}

	if prompt == "" {
		prompt = "Select section:"
	}

	options := make([]string, 0, len(sections))
	for i, section := range sections {
		line := fmt.Sprintf("[%d] ID: %d | %s", i+1, section.ID, section.Name)
		options = append(options, line)
	}

	back := len(allowBack) > 0 && allowBack[0]
	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Items:     options,
		AllowBack: back,
	})
	if err != nil {
		if IsGoBack(err) || IsExit(err) {
			return 0, err
		}
		return 0, fmt.Errorf("failed to select section: %w", err)
	}

	return sections[idx].ID, nil
}

// SelectSuiteForProject fetches suites for a project and selects one using the prompter.
// Pass allowBack=true to show "← Back" in addition to "✕ Exit".
func SelectSuiteForProject(ctx context.Context, p Prompter, cli client.ClientInterface, projectID int64, prompt string, allowBack ...bool) (int64, error) {
	suites, err := cli.GetSuites(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get suites for project %d: %w", projectID, err)
	}
	return SelectSuite(ctx, p, suites, prompt, allowBack...)
}
