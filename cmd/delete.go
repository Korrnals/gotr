package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/snap"
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

For milestones use: gotr milestones delete
For plans use: gotr plans delete

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
	deleteCmd.Flags().Int64("project-id", 0, "Project ID (required for section cascade snapshot)")
	deleteCmd.Flags().Int64("suite-id", 0, "Suite ID (required for section cascade snapshot)")
	snap.RegisterFlags(deleteCmd)
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
	var deleteErr error
	switch endpoint {
	case "project":
		snap.HookMutation(ctx, snap.Mutation{Cmd: cmd, Op: snap.OpDelete, EntityType: "project",
			EntityIDs: []int64{id}, Tier: snap.Tier2,
			FetchFn: func(ctx context.Context) (interface{}, error) { return cli.GetProject(ctx, id) }})
		deleteErr = cli.DeleteProject(ctx, id)
	case "suite":
		snap.HookMutation(ctx, snap.Mutation{Cmd: cmd, Op: snap.OpDelete, EntityType: "suite",
			EntityIDs: []int64{id}, Tier: snap.Tier2,
			FetchFn: func(ctx context.Context) (interface{}, error) { return cli.GetSuite(ctx, id) }})
		deleteErr = cli.DeleteSuite(ctx, id)
	case "section":
		deleteErr = deleteSectionWithSnap(cmd, cli, ctx, id)
	case "case":
		deleteErr = deleteCaseWithSnap(cmd, cli, ctx, id)
	case "run":
		snap.HookMutation(ctx, snap.Mutation{Cmd: cmd, Op: snap.OpDelete, EntityType: "run",
			EntityIDs: []int64{id}, Tier: snap.Tier2,
			FetchFn: func(ctx context.Context) (interface{}, error) { return cli.GetRun(ctx, id) }})
		deleteErr = cli.DeleteRun(ctx, id)
	case "shared-step":
		snap.HookMutation(ctx, snap.Mutation{Cmd: cmd, Op: snap.OpDelete, EntityType: "shared_step",
			EntityIDs: []int64{id}, Tier: snap.Tier2,
			FetchFn: func(ctx context.Context) (interface{}, error) { return cli.GetSharedStep(ctx, id) }})
		deleteErr = cli.DeleteSharedStep(ctx, id, 0)
	default:
		return fmt.Errorf("unsupported endpoint: %s", endpoint)
	}
	if deleteErr != nil {
		return deleteErr
	}

	interactive.MutationPostAction(ctx, cmd)
	return nil
}

// deleteSectionWithSnap creates a cascade snapshot before deleting a section.
func deleteSectionWithSnap(cmd *cobra.Command, cli client.ClientInterface, ctx context.Context, sectionID int64) error {
	hook := snap.NewHook(cmd)

	projectID, _ := cmd.Flags().GetInt64("project-id")
	suiteID, _ := cmd.Flags().GetInt64("suite-id")

	hook.Before(ctx, snap.BuildMeta(
		snap.OpDelete, "section", []int64{sectionID},
		snap.Tier2, projectID, suiteID, snap.ResolveName(cmd), os.Args[1:],
		snap.CurrentServerURL(),
	), func(ctx context.Context) (interface{}, error) {
		section, err := cli.GetSection(ctx, sectionID)
		if err != nil {
			return nil, err
		}
		if section == nil {
			return nil, fmt.Errorf("section %d not found", sectionID)
		}

		cascade := snap.CascadeData{Section: *section}

		// Fetch child cases if projectID is available.
		pid := projectID
		sid := suiteID
		if sid == 0 {
			sid = section.SuiteID
		}
		if pid > 0 && sid > 0 {
			cases, err := cli.GetCases(ctx, pid, sid, sectionID)
			if err == nil {
				cascade.Cases = cases
			}
		}
		return cascade, nil
	})

	return cli.DeleteSection(ctx, sectionID)
}

// deleteCaseWithSnap creates a snapshot before deleting a case.
func deleteCaseWithSnap(cmd *cobra.Command, cli client.ClientInterface, ctx context.Context, caseID int64) error {
	hook := snap.NewHook(cmd)
	hook.Before(ctx, snap.BuildMeta(
		snap.OpDelete, "case", []int64{caseID},
		snap.Tier2, 0, 0, snap.ResolveName(cmd), os.Args[1:],
		snap.CurrentServerURL(),
	), func(ctx context.Context) (interface{}, error) {
		c, err := cli.GetCase(ctx, caseID)
		if err != nil {
			return nil, err
		}
		if c == nil {
			return nil, fmt.Errorf("case %d not found", caseID)
		}
		return c, nil
	})

	return cli.DeleteCase(ctx, caseID)
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
	for {
		projectID, err := interactive.SelectProject(ctx, p, cli, "Select project for suite deletion:")
		if err != nil {
			return 0, err
		}
		suites, err := cli.GetSuites(ctx, projectID)
		if err != nil {
			return 0, fmt.Errorf("failed to get suites for project %d: %w", projectID, err)
		}
		suiteID, err := interactive.SelectSuite(ctx, p, suites, "", true)
		if err != nil {
			if interactive.IsGoBack(err) {
				continue
			}
			return 0, err
		}
		return suiteID, nil
	}
}

func resolveDeleteSection(ctx context.Context, p interactive.Prompter, cli client.ClientInterface) (int64, error) {
	var projectID, suiteID int64
	step := 0 // 0=project, 1=suite, 2=section
	for {
		switch step {
		case 0:
			var err error
			projectID, err = interactive.SelectProject(ctx, p, cli, "Select project for section deletion:")
			if err != nil {
				return 0, err
			}
			step = 1
		case 1:
			var err error
			suiteID, err = interactive.SelectSuiteForProject(ctx, p, cli, projectID, "Select suite for section deletion:", true)
			if err != nil {
				if interactive.IsGoBack(err) {
					step = 0
					continue
				}
				return 0, err
			}
			step = 2
		case 2:
			sections, err := cli.GetSections(ctx, projectID, suiteID)
			if err != nil {
				return 0, fmt.Errorf("failed to get sections for project %d suite %d: %w", projectID, suiteID, err)
			}
			sectionID, err := interactive.SelectSection(ctx, p, sections, "", true)
			if err != nil {
				if interactive.IsGoBack(err) {
					step = 1
					continue
				}
				return 0, err
			}
			return sectionID, nil
		}
	}
}

func resolveDeleteCase(ctx context.Context, p interactive.Prompter, cli client.ClientInterface) (int64, error) {
	var projectID, suiteID int64
	step := 0 // 0=project, 1=suite, 2=case
	for {
		switch step {
		case 0:
			var err error
			projectID, err = interactive.SelectProject(ctx, p, cli, "Select project for case deletion:")
			if err != nil {
				return 0, err
			}
			step = 1
		case 1:
			var err error
			suiteID, err = interactive.SelectSuiteForProject(ctx, p, cli, projectID, "Select suite for case deletion:", true)
			if err != nil {
				if interactive.IsGoBack(err) {
					step = 0
					continue
				}
				return 0, err
			}
			step = 2
		case 2:
			cases, err := cli.GetCases(ctx, projectID, suiteID, 0)
			if err != nil {
				return 0, fmt.Errorf("failed to get cases for project %d suite %d: %w", projectID, suiteID, err)
			}
			caseID, err := interactive.SelectCase(ctx, p, cases, "", true)
			if err != nil {
				if interactive.IsGoBack(err) {
					step = 1
					continue
				}
				return 0, err
			}
			return caseID, nil
		}
	}
}

func resolveDeleteRun(ctx context.Context, p interactive.Prompter, cli client.ClientInterface) (int64, error) {
	for {
		projectID, err := interactive.SelectProject(ctx, p, cli, "Select project for run deletion:")
		if err != nil {
			return 0, err
		}
		runs, err := cli.GetRuns(ctx, projectID)
		if err != nil {
			return 0, fmt.Errorf("failed to get runs for project %d: %w", projectID, err)
		}
		runID, err := interactive.SelectRun(ctx, p, runs, "", true)
		if err != nil {
			if interactive.IsGoBack(err) {
				continue
			}
			return 0, err
		}
		return runID, nil
	}
}

func resolveDeleteSharedStep(ctx context.Context, p interactive.Prompter, cli client.ClientInterface) (int64, error) {
	for {
		projectID, err := interactive.SelectProject(ctx, p, cli, "Select project for shared-step deletion:")
		if err != nil {
			return 0, err
		}
		steps, err := cli.GetSharedSteps(ctx, projectID)
		if err != nil {
			return 0, fmt.Errorf("failed to get shared steps for project %d: %w", projectID, err)
		}
		stepID, err := interactive.SelectSharedStep(ctx, p, steps, "", true)
		if err != nil {
			if interactive.IsGoBack(err) {
				continue
			}
			return 0, err
		}
		return stepID, nil
	}
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
