package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// deleteCmd deletes resources via DELETE/POST requests.
var deleteCmd = &cobra.Command{
	Use:   "delete <endpoint> <id>",
	Short: "Delete a resource (DELETE/POST request)",
	Long: `Deletes an existing object in TestRail.

Supported endpoints:
  project <id>       Delete a project
  suite <id>         Delete a suite
  section <id>       Delete a section
  case <id>          Delete a test case
  run <id>           Delete a test run
  shared-step <id>   Delete a shared step
  milestone <id>     Delete a milestone
  plan <id>          Delete a test plan

Examples:
  gotr delete project 1
  gotr delete case 12345
  gotr delete run 1000

Dry-run mode:
  gotr delete case 12345 --dry-run  # Show what would be deleted`,
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().Bool("dry-run", false, "Show what would be executed without making changes")
	deleteCmd.Flags().Bool("soft", false, "Soft delete (where supported)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if len(args) == 0 && !interactive.HasPrompterInContext(ctx) {
		return fmt.Errorf("endpoint and id required: gotr delete <endpoint> <id>")
	}

	cli := GetClient(cmd)
	p := interactive.PrompterFromContext(ctx)

	endpoint := ""
	if len(args) > 0 {
		endpoint = args[0]
	} else {
		selectedEndpoint, err := selectDeleteEndpoint(p)
		if err != nil {
			return err
		}
		endpoint = selectedEndpoint
	}

	id, err := parseDeleteIDArg(args)
	if err != nil {
		return err
	}
	if id == 0 {
		id, err = resolveDeleteID(ctx, p, cli, endpoint)
		if err != nil {
			return err
		}
	}

	// Check dry-run mode
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("delete " + endpoint)
		return runDeleteDryRun(dr, endpoint, id)
	}

	// Route by endpoint
	switch endpoint {
	case "project":
		return cli.DeleteProject(ctx, id)
	case "suite":
		return cli.DeleteSuite(ctx, id)
	case "section":
		return cli.DeleteSection(ctx, id)
	case "case":
		return cli.DeleteCase(ctx, id)
	case "run":
		return cli.DeleteRun(ctx, id)
	case "shared-step":
		// Shared step has a special keep_in_cases flag
		return cli.DeleteSharedStep(ctx, id, 0)
	default:
		return fmt.Errorf("unsupported endpoint: %s", endpoint)
	}
}

func parseDeleteIDArg(args []string) (int64, error) {
	if len(args) < 2 {
		return 0, nil
	}

	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid ID: %s", args[1])
	}

	return id, nil
}

func selectDeleteEndpoint(p interactive.Prompter) (string, error) {
	options := []string{"project", "suite", "section", "case", "run", "shared-step"}
	idx, _, err := p.Select("Select endpoint to delete:", options)
	if err != nil {
		return "", fmt.Errorf("failed to select endpoint: %w", err)
	}

	return options[idx], nil
}

var deleteResolvers = map[string]func(context.Context, interactive.Prompter, client.ClientInterface) (int64, error){
	"project":     resolveDeleteProject,
	"suite":       resolveDeleteSuite,
	"section":     resolveDeleteSection,
	"case":        resolveDeleteCase,
	"run":         resolveDeleteRun,
	"shared-step": resolveDeleteSharedStep,
}

func resolveDeleteID(ctx context.Context, p interactive.Prompter, cli client.ClientInterface, endpoint string) (int64, error) {
	resolver, ok := deleteResolvers[endpoint]
	if !ok {
		return 0, fmt.Errorf("unsupported endpoint: %s", endpoint)
	}
	return resolver(ctx, p, cli)
}

func resolveDeleteProject(ctx context.Context, p interactive.Prompter, cli client.ClientInterface) (int64, error) {
	return interactive.SelectProject(ctx, p, cli, "")
}

func resolveDeleteSuite(ctx context.Context, p interactive.Prompter, cli client.ClientInterface) (int64, error) {
	projectID, err := interactive.SelectProject(ctx, p, cli, "Select project for suite deletion:")
	if err != nil {
		return 0, err
	}
	suites, err := cli.GetSuites(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get suites for project %d: %w", projectID, err)
	}
	return interactive.SelectSuite(ctx, p, suites, "")
}

func resolveDeleteSection(ctx context.Context, p interactive.Prompter, cli client.ClientInterface) (int64, error) {
	projectID, err := interactive.SelectProject(ctx, p, cli, "Select project for section deletion:")
	if err != nil {
		return 0, err
	}
	suiteID, err := interactive.SelectSuiteForProject(ctx, p, cli, projectID, "Select suite for section deletion:")
	if err != nil {
		return 0, err
	}
	sections, err := cli.GetSections(ctx, projectID, suiteID)
	if err != nil {
		return 0, fmt.Errorf("failed to get sections for project %d suite %d: %w", projectID, suiteID, err)
	}
	return interactive.SelectSection(ctx, p, sections, "")
}

func resolveDeleteCase(ctx context.Context, p interactive.Prompter, cli client.ClientInterface) (int64, error) {
	projectID, err := interactive.SelectProject(ctx, p, cli, "Select project for case deletion:")
	if err != nil {
		return 0, err
	}
	suiteID, err := interactive.SelectSuiteForProject(ctx, p, cli, projectID, "Select suite for case deletion:")
	if err != nil {
		return 0, err
	}
	cases, err := cli.GetCases(ctx, projectID, suiteID, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get cases for project %d suite %d: %w", projectID, suiteID, err)
	}
	return selectCaseID(ctx, p, cases)
}

func resolveDeleteRun(ctx context.Context, p interactive.Prompter, cli client.ClientInterface) (int64, error) {
	projectID, err := interactive.SelectProject(ctx, p, cli, "Select project for run deletion:")
	if err != nil {
		return 0, err
	}
	runs, err := cli.GetRuns(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get runs for project %d: %w", projectID, err)
	}
	return interactive.SelectRun(ctx, p, runs, "")
}

func resolveDeleteSharedStep(ctx context.Context, p interactive.Prompter, cli client.ClientInterface) (int64, error) {
	projectID, err := interactive.SelectProject(ctx, p, cli, "Select project for shared-step deletion:")
	if err != nil {
		return 0, err
	}
	steps, err := cli.GetSharedSteps(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get shared steps for project %d: %w", projectID, err)
	}
	return selectSharedStepID(p, steps)
}

func selectCaseID(ctx context.Context, p interactive.Prompter, cases data.GetCasesResponse) (int64, error) {
	_ = ctx
	if len(cases) == 0 {
		return 0, fmt.Errorf("no cases found")
	}

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

func selectSharedStepID(p interactive.Prompter, steps data.GetSharedStepsResponse) (int64, error) {
	if len(steps) == 0 {
		return 0, fmt.Errorf("no shared steps found")
	}

	options := make([]string, 0, len(steps))
	for i, step := range steps {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, step.ID, step.Title))
	}

	idx, _, err := p.Select("Select shared step:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select shared step: %w", err)
	}

	return steps[idx].ID, nil
}

// runDeleteDryRun performs a dry-run for the delete command.
func runDeleteDryRun(dr *output.DryRunPrinter, endpoint string, id int64) error {
	var method, url string

	switch endpoint {
	case "project":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_project/%d", id)
	case "suite":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_suite/%d", id)
	case "section":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_section/%d", id)
	case "case":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_case/%d", id)
	case "run":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_run/%d", id)
	case "shared-step":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_shared_step/%d", id)
	default:
		return fmt.Errorf("unsupported endpoint for dry-run: %s", endpoint)
	}

	dr.PrintOperation(
		fmt.Sprintf("Delete %s %d", endpoint, id),
		method,
		url,
		nil,
	)
	return nil
}
