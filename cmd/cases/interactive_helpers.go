package cases

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

func resolveSectionIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	projectID, suiteID, err := resolveProjectAndSuiteInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	sections, err := cli.GetSections(ctx, projectID, suiteID)
	if err != nil {
		return 0, fmt.Errorf("failed to get sections: %w", err)
	}
	if len(sections) == 0 {
		return 0, fmt.Errorf("no sections found in suite %d", suiteID)
	}

	return selectSectionID(ctx, sections)
}

func resolveSuiteIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	_, suiteID, err := resolveProjectAndSuiteInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}
	return suiteID, nil
}

func resolveProjectAndSuiteInteractive(ctx context.Context, cli client.ClientInterface) (projectID, suiteID int64, err error) {
	p := interactive.PrompterFromContext(ctx)
	projectID, err = interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, 0, err
	}

	suites, err := cli.GetSuites(ctx, projectID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get suites: %w", err)
	}
	if len(suites) == 0 {
		return 0, 0, fmt.Errorf("no suites found in project %d", projectID)
	}

	suiteID, err = interactive.SelectSuite(ctx, p, suites, "")
	if err != nil {
		return 0, 0, err
	}

	return projectID, suiteID, nil
}

func selectSectionID(ctx context.Context, sections data.GetSectionsResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(sections))
	for i, section := range sections {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, section.ID, section.Name))
	}

	idx, _, err := p.Select("Select section:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select section: %w", err)
	}

	return sections[idx].ID, nil
}
