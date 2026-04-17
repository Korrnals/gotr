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

	cols := []Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Name"},
	}
	rows := make([][]string, len(projects))
	for i, proj := range projects {
		rows[i] = []string{fmt.Sprintf("%d", proj.ID), proj.Name}
	}
	header, options := AlignedLabels(cols, rows)

	back := len(allowBack) > 0 && allowBack[0]
	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Header:    header,
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

	cols := []Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Name"},
	}
	rows := make([][]string, len(suites))
	for i, s := range suites {
		rows[i] = []string{fmt.Sprintf("%d", s.ID), s.Name}
	}
	header, options := AlignedLabels(cols, rows)

	back := len(allowBack) > 0 && allowBack[0]
	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Header:    header,
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

	cols := []Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Status", MinWidth: 10},
		{Header: "Name"},
	}
	rows := make([][]string, len(runs))
	for i, run := range runs {
		status := "active"
		if run.IsCompleted {
			status = "completed"
		}
		rows[i] = []string{fmt.Sprintf("%d", run.ID), status, run.Name}
	}
	header, options := AlignedLabels(cols, rows)

	back := len(allowBack) > 0 && allowBack[0]
	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Header:    header,
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

	cols := []Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Name"},
	}
	rows := make([][]string, len(sections))
	for i, sec := range sections {
		rows[i] = []string{fmt.Sprintf("%d", sec.ID), sec.Name}
	}
	header, options := AlignedLabels(cols, rows)

	back := len(allowBack) > 0 && allowBack[0]
	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Header:    header,
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

// SelectCase selects a case using unified prompter with Browse navigation.
// Pass allowBack=true to show "← Back" in addition to "✕ Exit".
func SelectCase(ctx context.Context, p Prompter, cases data.GetCasesResponse, prompt string, allowBack ...bool) (int64, error) {
	if len(cases) == 0 {
		return 0, fmt.Errorf("no cases found")
	}

	if prompt == "" {
		prompt = "Select case:"
	}

	cols := []Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Title"},
	}
	rows := make([][]string, len(cases))
	for i, c := range cases {
		rows[i] = []string{fmt.Sprintf("%d", c.ID), c.Title}
	}
	header, options := AlignedLabels(cols, rows)

	back := len(allowBack) > 0 && allowBack[0]
	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Header:    header,
		Items:     options,
		AllowBack: back,
	})
	if err != nil {
		if IsGoBack(err) || IsExit(err) {
			return 0, err
		}
		return 0, fmt.Errorf("failed to select case: %w", err)
	}

	return cases[idx].ID, nil
}

// SelectSharedStep selects a shared step using unified prompter with Browse navigation.
// Pass allowBack=true to show "← Back" in addition to "✕ Exit".
func SelectSharedStep(ctx context.Context, p Prompter, steps data.GetSharedStepsResponse, prompt string, allowBack ...bool) (int64, error) {
	if len(steps) == 0 {
		return 0, fmt.Errorf("no shared steps found")
	}

	if prompt == "" {
		prompt = "Select shared step:"
	}

	cols := []Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Title"},
	}
	rows := make([][]string, len(steps))
	for i, s := range steps {
		rows[i] = []string{fmt.Sprintf("%d", s.ID), s.Title}
	}
	header, options := AlignedLabels(cols, rows)

	back := len(allowBack) > 0 && allowBack[0]
	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Header:    header,
		Items:     options,
		AllowBack: back,
	})
	if err != nil {
		if IsGoBack(err) || IsExit(err) {
			return 0, err
		}
		return 0, fmt.Errorf("failed to select shared step: %w", err)
	}

	return steps[idx].ID, nil
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
