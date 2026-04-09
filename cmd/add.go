package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/crud"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// addCmd creates resources via POST requests.
var addCmd = &cobra.Command{
	Use:   "add <endpoint> [id]",
	Short: "Create a new resource (POST request)",
	Long: `Creates a new object in TestRail via the POST API.

Supported endpoints:
  project              Create a project
  suite <project_id>   Create a suite in a project
  section <project_id> Create a section in a project/suite
  case <section_id>    Create a test case in a section
  run <project_id>     Create a test run
  result <test_id>     Add a test result
  result-for-case <run_id> <case_id>  Add a result for a case
  shared-step <project_id>  Create a shared step
  milestone <project_id>    Create a milestone
  plan <project_id>         Create a test plan
  entry <plan_id>           Add an entry to a plan
  attachment case <case_id> <file>    Attach a file to a case
  attachment plan <plan_id> <file>    Attach a file to a plan
  attachment plan-entry <plan_id> <entry_id> <file>  Attach a file to an entry
  attachment result <result_id> <file>  Attach a file to a result
  attachment run <run_id> <file>      Attach a file to a run

Examples:
  gotr add project --name "New Project" --announcement "Desc"
  gotr add suite 1 --name "Smoke Tests"
  gotr add case 100 --title "Login test" --template-id 1
  gotr add run 1 --name "Nightly Run" --suite-id 100
  gotr add result 12345 --status-id 1 --comment "Passed"
  gotr add attachment case 12345 ./screenshot.png
  gotr add attachment plan 100 ./report.pdf
  gotr add attachment result 98765 ./log.txt

Interactive mode (wizard):
  gotr add project -i
  gotr add suite 1 -i
  gotr add case 100 -i

Dry-run mode:
  gotr add project --name "Test" --dry-run  # Show what would be created`,
	RunE: runAdd,
}

func init() {
	// Common flags for resource creation
	addCmd.Flags().StringP("name", "n", "", "Resource name")
	addCmd.Flags().String("description", "", "Description/announcement")
	addCmd.Flags().String("announcement", "", "Announcement (for project)")
	addCmd.Flags().Bool("show-announcement", false, "Show announcement")
	addCmd.Flags().Int64("suite-id", 0, "Suite ID")
	addCmd.Flags().Int64("section-id", 0, "Section ID")
	addCmd.Flags().Int64("milestone-id", 0, "Milestone ID")
	addCmd.Flags().Int64("template-id", 0, "Template ID (for case)")
	addCmd.Flags().Int64("type-id", 0, "Type ID (for case)")
	addCmd.Flags().Int64("priority-id", 0, "Priority ID (for case)")
	addCmd.Flags().String("title", "", "Title (for case)")
	addCmd.Flags().String("refs", "", "References")
	addCmd.Flags().String("comment", "", "Comment (for result)")
	addCmd.Flags().Int64("status-id", 0, "Status ID (for result)")
	addCmd.Flags().String("elapsed", "", "Elapsed time (for result)")
	addCmd.Flags().String("defects", "", "Defects (for result)")
	addCmd.Flags().Int64("assignedto-id", 0, "Assigned user ID")
	addCmd.Flags().String("case-ids", "", "Comma-separated case IDs (for run)")
	addCmd.Flags().Bool("include-all", true, "Include all cases (for run)")
	addCmd.Flags().String("json-file", "", "Path to JSON data file")
	output.AddFlag(addCmd)
	addCmd.Flags().Bool("dry-run", false, "Show what would be executed without making changes")
	addCmd.Flags().BoolP("interactive", "i", false, "Interactive mode (wizard)")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("endpoint required: project, suite, section, case, run, result, result-for-case, shared-step, milestone, plan, entry, attachment")
	}

	endpoint := args[0]
	ctx := cmd.Context()

	// Get the client
	cli := GetClient(cmd)

	// Parse ID from args (attachment has its own argument structure)
	var id int64
	if len(args) > 1 && endpoint != "attachment" {
		parsedID, err := flags.ValidateRequiredID(args, 1, "ID")
		if err != nil {
			return err
		}
		id = parsedID
	}

	resolvedID, err := resolveAddParentID(ctx, interactive.PrompterFromContext(ctx), cli, endpoint, id)
	if err != nil {
		return err
	}
	id = resolvedID

	// Read JSON from file if specified
	jsonFile, _ := cmd.Flags().GetString("json-file")
	var jsonData []byte
	if jsonFile != "" {
		var err error
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("JSON file read error: %w", err)
		}
	}

	// Check dry-run mode (attachment has its own dry-run handling)
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun && endpoint != "attachment" {
		dr := output.NewDryRunPrinter("add " + endpoint)
		return runAddDryRun(cmd, dr, endpoint, id, jsonData)
	}

	// Check interactive mode
	isInteractive, _ := cmd.Flags().GetBool("interactive")
	if isInteractive || shouldAutoRunAddInteractive(cmd, endpoint, id, jsonFile != "") {
		return runAddInteractive(cli, cmd, endpoint, id)
	}

	return dispatchAdd(cli, cmd, endpoint, id, args, jsonData)
}

// addEndpointSpec describes a simple add endpoint with optional ID requirement.
type addEndpointSpec struct {
	idLabel string // empty for endpoints that don't require an ID (e.g. project)
	fn      func(client.ClientInterface, *cobra.Command, int64, []byte) error
}

var addEndpoints = map[string]addEndpointSpec{
	"project": {"", func(cli client.ClientInterface, cmd *cobra.Command, _ int64, jsonData []byte) error {
		return addProject(cli, cmd, jsonData)
	}},
	"suite":        {"project_id", addSuite},
	"section":      {"project_id", addSection},
	"case":         {"section_id", addCase},
	"run":          {"project_id", addRun},
	"result":       {"test_id", addResult},
	"shared-step":  {"project_id", addSharedStep},
}

func dispatchAdd(cli client.ClientInterface, cmd *cobra.Command, endpoint string, id int64, args []string, jsonData []byte) error {
	if spec, ok := addEndpoints[endpoint]; ok {
		if spec.idLabel != "" && id == 0 {
			return fmt.Errorf("%s required: gotr add %s <%s>", spec.idLabel, endpoint, spec.idLabel)
		}
		return spec.fn(cli, cmd, id, jsonData)
	}

	switch endpoint {
	case "result-for-case":
		if id == 0 || len(args) < 3 {
			return fmt.Errorf("run_id and case_id required: gotr add result-for-case <run_id> <case_id>")
		}
		caseID, err := flags.ValidateRequiredID(args, 2, "case_id")
		if err != nil {
			return err
		}
		return addResultForCase(cli, cmd, id, caseID, jsonData)
	case "attachment":
		return runAddAttachment(cli, cmd, args)
	default:
		return fmt.Errorf("unsupported endpoint: %s", endpoint)
	}
}

func hasChangedFlag(cmd *cobra.Command, name string) bool {
	flag := cmd.Flags().Lookup(name)
	return flag != nil && cmd.Flags().Changed(name)
}

func hasAnyChangedFlag(cmd *cobra.Command, names ...string) bool {
	for _, name := range names {
		if hasChangedFlag(cmd, name) {
			return true
		}
	}
	return false
}

func shouldAutoRunAddInteractive(cmd *cobra.Command, endpoint string, parentID int64, hasJSONFile bool) bool {
	if hasJSONFile || !interactive.HasPrompterInContext(cmd.Context()) {
		return false
	}

	switch endpoint {
	case "project":
		return !hasAnyChangedFlag(cmd, "name", "announcement", "show-announcement")
	case "suite":
		return parentID != 0 && !hasAnyChangedFlag(cmd, "name", "description")
	case "section":
		return parentID != 0 && !hasAnyChangedFlag(cmd, "name", "description", "suite-id", "section-id")
	case "case":
		return parentID != 0 && !hasAnyChangedFlag(cmd, "title", "type-id", "priority-id", "refs")
	case "run":
		return parentID != 0 && !hasAnyChangedFlag(cmd, "name", "description", "suite-id", "case-ids", "include-all", "milestone-id", "assignedto-id")
	case "shared-step":
		return parentID != 0 && !hasAnyChangedFlag(cmd, "title")
	default:
		return false
	}
}

func resolveAddParentID(ctx context.Context, p interactive.Prompter, cli client.ClientInterface, endpoint string, currentID int64) (int64, error) {
	if currentID != 0 || !interactive.HasPrompterInContext(ctx) {
		return currentID, nil
	}

	switch endpoint {
	case "suite", "section", "run", "shared-step":
		projectID, err := interactive.SelectProject(ctx, p, cli, "")
		if err != nil {
			return 0, err
		}
		return projectID, nil
	case "case":
		projectID, err := interactive.SelectProject(ctx, p, cli, "")
		if err != nil {
			return 0, err
		}

		suites, err := cli.GetSuites(ctx, projectID)
		if err != nil {
			return 0, fmt.Errorf("failed to get suites for project %d: %w", projectID, err)
		}

		var suiteID int64
		switch len(suites) {
		case 0:
			suiteID = 0
		case 1:
			suiteID = suites[0].ID
		default:
			suiteID, err = interactive.SelectSuite(ctx, p, suites, "")
			if err != nil {
				return 0, err
			}
		}

		sections, err := cli.GetSections(ctx, projectID, suiteID)
		if err != nil {
			return 0, fmt.Errorf("failed to get sections for project %d: %w", projectID, err)
		}

		if len(sections) == 1 {
			return sections[0].ID, nil
		}

		sectionID, err := interactive.SelectSection(ctx, p, sections, "")
		if err != nil {
			return 0, err
		}
		return sectionID, nil
	default:
		return currentID, nil
	}
}

// runAddInteractive starts an interactive wizard for creating a resource.
func runAddInteractive(cli client.ClientInterface, cmd *cobra.Command, endpoint string, parentID int64) error {
	switch endpoint {
	case "project":
		return addProjectInteractive(cli, cmd)
	case "suite":
		if parentID == 0 {
			return fmt.Errorf("project_id required: gotr add suite <project_id> --interactive")
		}
		return addSuiteInteractive(cli, cmd, parentID)
	case "section":
		if parentID == 0 {
			return fmt.Errorf("project_id required: gotr add section <project_id> --interactive")
		}
		return addSectionInteractive(cli, cmd, parentID)
	case "case":
		if parentID == 0 {
			return fmt.Errorf("section_id required: gotr add case <section_id> --interactive")
		}
		return addCaseInteractive(cli, cmd, parentID)
	case "run":
		if parentID == 0 {
			return fmt.Errorf("project_id required: gotr add run <project_id> --interactive")
		}
		return addRunInteractive(cli, cmd, parentID)
	case "shared-step":
		if parentID == 0 {
			return fmt.Errorf("project_id required: gotr add shared-step <project_id> --interactive")
		}
		return addSharedStepInteractive(cli, cmd, parentID)
	default:
		return fmt.Errorf("interactive mode not supported for endpoint: %s", endpoint)
	}
}

func parseOptionalID(input string) (int64, error) {
	if input == "" {
		return 0, nil
	}
	return flags.ParseID(input)
}

func addProjectInteractive(cli client.ClientInterface, cmd *cobra.Command) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)
	answers, err := interactive.AskProjectWithPrompter(p, false)
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	// Preview
	ui.Preview(os.Stdout, "Create Project", []ui.PreviewField{
		{Label: "Name", Value: answers.Name},
		{Label: "Announcement", Value: answers.Announcement},
		{Label: "Show announce", Value: answers.ShowAnnouncement},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm creation?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.AddProjectRequest{
		Name:             answers.Name,
		Announcement:     answers.Announcement,
		ShowAnnouncement: answers.ShowAnnouncement,
	}

	project, err := cli.AddProject(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Project created (ID: %d)", project.ID)
	}
	return output.OutputResult(cmd, project, "result")
}

func addSuiteInteractive(cli client.ClientInterface, cmd *cobra.Command, projectID int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)
	answers, err := interactive.AskSuiteWithPrompter(p, false)
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	// Preview
	ui.Preview(os.Stdout, "Create Suite", []ui.PreviewField{
		{Label: "Name", Value: answers.Name},
		{Label: "Description", Value: answers.Description},
		{Label: "Project ID", Value: projectID},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm creation?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.AddSuiteRequest{
		Name:        answers.Name,
		Description: answers.Description,
	}

	suite, err := cli.AddSuite(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to create suite: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Suite created (ID: %d)", suite.ID)
	}
	return output.OutputResult(cmd, suite, "result")
}

func addCaseInteractive(cli client.ClientInterface, cmd *cobra.Command, sectionID int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)
	answers, err := interactive.AskCaseWithPrompter(p, false)
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	// Preview
	ui.Preview(os.Stdout, "Create Case", []ui.PreviewField{
		{Label: "Title", Value: answers.Title},
		{Label: "Section ID", Value: sectionID},
		{Label: "Type ID", Value: answers.TypeID},
		{Label: "Priority ID", Value: answers.PriorityID},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm creation?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.AddCaseRequest{
		Title:      answers.Title,
		SectionID:  sectionID,
		TypeID:     answers.TypeID,
		PriorityID: answers.PriorityID,
		Refs:       answers.Refs,
	}

	caseResp, err := cli.AddCase(ctx, sectionID, req)
	if err != nil {
		return fmt.Errorf("failed to create case: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Case created (ID: %d)", caseResp.ID)
	}
	return output.OutputResult(cmd, caseResp, "result")
}

func addRunInteractive(cli client.ClientInterface, cmd *cobra.Command, projectID int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)
	answers, err := interactive.AskRunWithPrompter(p, false)
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	// Preview
	ui.Preview(os.Stdout, "Create Run", []ui.PreviewField{
		{Label: "Name", Value: answers.Name},
		{Label: "Description", Value: answers.Description},
		{Label: "Suite ID", Value: answers.SuiteID},
		{Label: "Include all", Value: answers.IncludeAll},
		{Label: "Project ID", Value: projectID},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm creation?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.AddRunRequest{
		Name:        answers.Name,
		Description: answers.Description,
		SuiteID:     answers.SuiteID,
		IncludeAll:  answers.IncludeAll,
	}

	run, err := cli.AddRun(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Run created (ID: %d)", run.ID)
	}
	return output.OutputResult(cmd, run, "result")
}

func addSectionInteractive(cli client.ClientInterface, cmd *cobra.Command, projectID int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)

	name, err := p.Input("Section name:", "")
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}
	if name == "" {
		return fmt.Errorf("section name is required")
	}

	description, err := p.MultilineInput("Description (optional):", "")
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	suiteIDInput, err := p.Input("Suite ID (optional):", "")
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}
	suiteID, err := parseOptionalID(suiteIDInput)
	if err != nil {
		return fmt.Errorf("invalid suite id: %w", err)
	}

	parentSectionInput, err := p.Input("Parent section ID (optional):", "")
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}
	parentSectionID, err := parseOptionalID(parentSectionInput)
	if err != nil {
		return fmt.Errorf("invalid parent section id: %w", err)
	}

	ui.Preview(os.Stdout, "Create Section", []ui.PreviewField{
		{Label: "Name", Value: name},
		{Label: "Description", Value: description},
		{Label: "Project ID", Value: projectID},
		{Label: "Suite ID", Value: suiteID},
		{Label: "Parent Section ID", Value: parentSectionID},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm creation?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.AddSectionRequest{
		Name:        name,
		Description: description,
		SuiteID:     suiteID,
		ParentID:    parentSectionID,
	}

	section, err := cli.AddSection(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to create section: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Section created (ID: %d)", section.ID)
	}
	return output.OutputResult(cmd, section, "result")
}

func addSharedStepInteractive(cli client.ClientInterface, cmd *cobra.Command, projectID int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)

	title, err := p.Input("Shared step title:", "")
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}
	if title == "" {
		return fmt.Errorf("shared step title is required")
	}

	ui.Preview(os.Stdout, "Create Shared Step", []ui.PreviewField{
		{Label: "Title", Value: title},
		{Label: "Project ID", Value: projectID},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm creation?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.AddSharedStepRequest{Title: title}
	step, err := cli.AddSharedStep(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to create shared step: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Shared step created (ID: %d)", step.ID)
	}
	return output.OutputResult(cmd, step, "result")
}

// runAddDryRun performs a dry-run for the add command.
func runAddDryRun(cmd *cobra.Command, dr *output.DryRunPrinter, endpoint string, id int64, jsonData []byte) error {
	handler, ok := addDryRunHandlers[endpoint]
	if !ok {
		if endpoint == "attachment" {
			return fmt.Errorf("use --dry-run with a specific attachment subcommand")
		}
		return fmt.Errorf("unsupported endpoint for dry-run: %s", endpoint)
	}
	return handler(cmd, dr, id, jsonData)
}

var addDryRunHandlers = map[string]func(*cobra.Command, *output.DryRunPrinter, int64, []byte) error{
	"project":     dryRunAddProject,
	"suite":       dryRunAddSuite,
	"section":     dryRunAddSection,
	"case":        dryRunAddCase,
	"run":         dryRunAddRun,
	"result":      dryRunAddResult,
	"shared-step": dryRunAddSharedStep,
}

// --- Request builders (shared between execute and dry-run) ---

func buildAddProjectReq(cmd *cobra.Command, validate bool) (*data.AddProjectRequest, error) {
	name, _ := cmd.Flags().GetString("name")
	if validate && name == "" {
		return nil, fmt.Errorf("--name is required")
	}
	req := &data.AddProjectRequest{Name: name}
	req.Announcement, _ = cmd.Flags().GetString("announcement")
	req.ShowAnnouncement, _ = cmd.Flags().GetBool("show-announcement")
	return req, nil
}

func buildAddSuiteReq(cmd *cobra.Command, validate bool) (*data.AddSuiteRequest, error) {
	name, _ := cmd.Flags().GetString("name")
	if validate && name == "" {
		return nil, fmt.Errorf("--name is required")
	}
	req := &data.AddSuiteRequest{Name: name}
	req.Description, _ = cmd.Flags().GetString("description")
	return req, nil
}

func buildAddSectionReq(cmd *cobra.Command, validate bool) (*data.AddSectionRequest, error) {
	name, _ := cmd.Flags().GetString("name")
	if validate && name == "" {
		return nil, fmt.Errorf("--name is required")
	}
	req := &data.AddSectionRequest{Name: name}
	req.Description, _ = cmd.Flags().GetString("description")
	req.SuiteID, _ = cmd.Flags().GetInt64("suite-id")
	req.ParentID, _ = cmd.Flags().GetInt64("section-id")
	return req, nil
}

func buildAddCaseReq(cmd *cobra.Command, validate bool) (*data.AddCaseRequest, error) {
	title, _ := cmd.Flags().GetString("title")
	if validate && title == "" {
		return nil, fmt.Errorf("--title is required")
	}
	req := &data.AddCaseRequest{Title: title}
	req.TemplateID, _ = cmd.Flags().GetInt64("template-id")
	req.TypeID, _ = cmd.Flags().GetInt64("type-id")
	req.PriorityID, _ = cmd.Flags().GetInt64("priority-id")
	req.Refs, _ = cmd.Flags().GetString("refs")
	return req, nil
}

func buildAddRunReq(cmd *cobra.Command, validate bool) (*data.AddRunRequest, error) {
	name, _ := cmd.Flags().GetString("name")
	if validate && name == "" {
		return nil, fmt.Errorf("--name is required")
	}
	req := &data.AddRunRequest{Name: name}
	req.Description, _ = cmd.Flags().GetString("description")
	req.SuiteID, _ = cmd.Flags().GetInt64("suite-id")
	req.MilestoneID, _ = cmd.Flags().GetInt64("milestone-id")
	req.AssignedTo, _ = cmd.Flags().GetInt64("assignedto-id")
	req.IncludeAll, _ = cmd.Flags().GetBool("include-all")
	caseIDsStr, _ := cmd.Flags().GetString("case-ids")
	if caseIDsStr != "" {
		req.CaseIDs = parseCaseIDs(caseIDsStr)
	}
	return req, nil
}

func buildAddResultReq(cmd *cobra.Command, validate bool) (*data.AddResultRequest, error) {
	statusID, _ := cmd.Flags().GetInt64("status-id")
	if validate && statusID == 0 {
		return nil, fmt.Errorf("--status-id is required")
	}
	req := &data.AddResultRequest{StatusID: statusID}
	req.Comment, _ = cmd.Flags().GetString("comment")
	req.Elapsed, _ = cmd.Flags().GetString("elapsed")
	req.Defects, _ = cmd.Flags().GetString("defects")
	req.AssignedTo, _ = cmd.Flags().GetInt64("assignedto-id")
	req.Version, _ = cmd.Flags().GetString("version")
	return req, nil
}

func buildAddSharedStepReq(cmd *cobra.Command, validate bool) (*data.AddSharedStepRequest, error) {
	title, _ := cmd.Flags().GetString("title")
	if validate && title == "" {
		return nil, fmt.Errorf("--title is required")
	}
	return &data.AddSharedStepRequest{Title: title}, nil
}

// --- Dry-run handlers (delegate to crud.DryRun) ---

func dryRunAddProject(cmd *cobra.Command, dr *output.DryRunPrinter, _ int64, jsonData []byte) error {
	return crud.DryRun(cmd, dr, jsonData, buildAddProjectReq,
		"Create Project", "POST", "/index.php?/api/v2/add_project/",
	)
}

func dryRunAddSuite(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	if id == 0 {
		return fmt.Errorf("project_id required: gotr add suite <project_id> --dry-run")
	}
	return crud.DryRun(cmd, dr, jsonData, buildAddSuiteReq,
		fmt.Sprintf("Create Suite in Project %d", id), "POST",
		fmt.Sprintf("/index.php?/api/v2/add_suite/%d", id),
	)
}

func dryRunAddSection(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	if id == 0 {
		return fmt.Errorf("project_id required: gotr add section <project_id> --dry-run")
	}
	return crud.DryRun(cmd, dr, jsonData, buildAddSectionReq,
		fmt.Sprintf("Create Section in Project %d", id), "POST",
		fmt.Sprintf("/index.php?/api/v2/add_section/%d", id),
	)
}

func dryRunAddCase(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	if id == 0 {
		return fmt.Errorf("section_id required: gotr add case <section_id> --dry-run")
	}
	return crud.DryRun(cmd, dr, jsonData, buildAddCaseReq,
		fmt.Sprintf("Create Case in Section %d", id), "POST",
		fmt.Sprintf("/index.php?/api/v2/add_case/%d", id),
	)
}

func dryRunAddRun(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	if id == 0 {
		return fmt.Errorf("project_id required: gotr add run <project_id> --dry-run")
	}
	return crud.DryRun(cmd, dr, jsonData, buildAddRunReq,
		fmt.Sprintf("Create Run in Project %d", id), "POST",
		fmt.Sprintf("/index.php?/api/v2/add_run/%d", id),
	)
}

func dryRunAddResult(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	if id == 0 {
		return fmt.Errorf("test_id required: gotr add result <test_id> --dry-run")
	}
	return crud.DryRun(cmd, dr, jsonData, buildAddResultReq,
		fmt.Sprintf("Add Result for Test %d", id), "POST",
		fmt.Sprintf("/index.php?/api/v2/add_result/%d", id),
	)
}

func dryRunAddSharedStep(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	if id == 0 {
		return fmt.Errorf("project_id required: gotr add shared-step <project_id> --dry-run")
	}
	return crud.DryRun(cmd, dr, jsonData, buildAddSharedStepReq,
		fmt.Sprintf("Create Shared Step in Project %d", id), "POST",
		fmt.Sprintf("/index.php?/api/v2/add_shared_step/%d", id),
	)
}

func addProject(cli client.ClientInterface, cmd *cobra.Command, jsonData []byte) error {
	return crud.Execute(cmd, 0, jsonData, buildAddProjectReq,
		func(ctx context.Context, _ int64, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
			return cli.AddProject(ctx, req)
		},
		"failed to create project",
	)
}

func addSuite(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	return crud.Execute(cmd, projectID, jsonData, buildAddSuiteReq,
		func(ctx context.Context, id int64, req *data.AddSuiteRequest) (*data.Suite, error) {
			return cli.AddSuite(ctx, id, req)
		},
		"failed to create suite",
	)
}

func addSection(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	return crud.Execute(cmd, projectID, jsonData, buildAddSectionReq,
		func(ctx context.Context, id int64, req *data.AddSectionRequest) (*data.Section, error) {
			return cli.AddSection(ctx, id, req)
		},
		"failed to create section",
	)
}

func addCase(cli client.ClientInterface, cmd *cobra.Command, sectionID int64, jsonData []byte) error {
	return crud.Execute(cmd, sectionID, jsonData, buildAddCaseReq,
		func(ctx context.Context, id int64, req *data.AddCaseRequest) (*data.Case, error) {
			return cli.AddCase(ctx, id, req)
		},
		"failed to create case",
	)
}

func addRun(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	return crud.Execute(cmd, projectID, jsonData, buildAddRunReq,
		func(ctx context.Context, id int64, req *data.AddRunRequest) (*data.Run, error) {
			return cli.AddRun(ctx, id, req)
		},
		"failed to create run",
	)
}

func addResult(cli client.ClientInterface, cmd *cobra.Command, testID int64, jsonData []byte) error {
	return crud.Execute(cmd, testID, jsonData, buildAddResultReq,
		func(ctx context.Context, id int64, req *data.AddResultRequest) (*data.Result, error) {
			return cli.AddResult(ctx, id, req)
		},
		"failed to add result",
	)
}

func addResultForCase(cli client.ClientInterface, cmd *cobra.Command, runID, caseID int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.AddResultRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		built, err := buildAddResultReq(cmd, true)
		if err != nil {
			return err
		}
		req = *built
	}

	result, err := cli.AddResultForCase(ctx, runID, caseID, &req)
	if err != nil {
		return fmt.Errorf("failed to add result: %w", err)
	}

	return output.OutputResult(cmd, result, "result")
}

func addSharedStep(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	return crud.Execute(cmd, projectID, jsonData, buildAddSharedStepReq,
		func(ctx context.Context, id int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
			return cli.AddSharedStep(ctx, id, req)
		},
		"failed to create shared step",
	)
}

func parseCaseIDs(s string) []int64 {
	var ids []int64
	for _, part := range splitAndTrim(s, ",") {
		id, err := flags.ParseID(part)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func splitAndTrim(s, sep string) []string {
	var parts []string
	for _, p := range splitString(s, sep) {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	var result []string
	var current string
	for _, c := range s {
		if string(c) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}

// attachmentSpec describes a simple attachment endpoint (ID + file path).
type attachmentSpec struct {
	minArgs int
	usage   string
	idLabel string
	handler func(client.ClientInterface, *cobra.Command, int64, string) error
}

var simpleAttachSpecs = map[string]attachmentSpec{
	"case":   {4, "gotr add attachment case <case_id> <file_path>", "case_id", addAttachmentToCase},
	"plan":   {4, "gotr add attachment plan <plan_id> <file_path>", "plan_id", addAttachmentToPlan},
	"result": {4, "gotr add attachment result <result_id> <file_path>", "result_id", addAttachmentToResult},
	"run":    {4, "gotr add attachment run <run_id> <file_path>", "run_id", addAttachmentToRun},
}

// runAddAttachment handles attachment creation (DEPRECATED: use 'gotr attachments add' instead).
func runAddAttachment(cli client.ClientInterface, cmd *cobra.Command, args []string) error {
	fmt.Fprintln(os.Stderr, "⚠️  WARNING: 'gotr add attachment' is deprecated. Use 'gotr attachments add' instead.")

	if len(args) < 2 {
		return fmt.Errorf("attachment type required: case, plan, plan-entry, result, run")
	}

	attachmentType := args[1]

	if spec, ok := simpleAttachSpecs[attachmentType]; ok {
		if len(args) < spec.minArgs {
			return fmt.Errorf("usage: %s", spec.usage)
		}
		id, err := flags.ValidateRequiredID(args, 2, spec.idLabel)
		if err != nil {
			return err
		}
		return spec.handler(cli, cmd, id, args[3])
	}

	switch attachmentType {
	case "plan-entry":
		if len(args) < 5 {
			return fmt.Errorf("usage: gotr add attachment plan-entry <plan_id> <entry_id> <file_path>")
		}
		planID, err := flags.ValidateRequiredID(args, 2, "plan_id")
		if err != nil {
			return err
		}
		return addAttachmentToPlanEntry(cli, cmd, planID, args[3], args[4])
	default:
		return fmt.Errorf("unsupported attachment type: %s. Available: case, plan, plan-entry, result, run", attachmentType)
	}
}

func addAttachmentToCase(cli client.ClientInterface, cmd *cobra.Command, caseID int64, filePath string) error {
	ctx := cmd.Context()
	// Check dry-run mode
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment case")
		dr.PrintSimple("Add Attachment to Case", fmt.Sprintf("Case ID: %d, File: %s", caseID, filePath))
		return nil
	}

	// Check that the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	resp, err := cli.AddAttachmentToCase(ctx, caseID, filePath)
	if err != nil {
		return fmt.Errorf("failed to add attachment to case: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Attachment added (ID: %d)", resp.AttachmentID)
		ui.Infof(os.Stdout, "URL: %s", resp.URL)
	}
	return output.OutputResult(cmd, resp, "result")
}

func addAttachmentToPlan(cli client.ClientInterface, cmd *cobra.Command, planID int64, filePath string) error {
	ctx := cmd.Context()
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment plan")
		dr.PrintSimple("Add Attachment to Plan", fmt.Sprintf("Plan ID: %d, File: %s", planID, filePath))
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	resp, err := cli.AddAttachmentToPlan(ctx, planID, filePath)
	if err != nil {
		return fmt.Errorf("failed to add attachment to plan: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Attachment added (ID: %d)", resp.AttachmentID)
		ui.Infof(os.Stdout, "URL: %s", resp.URL)
	}
	return output.OutputResult(cmd, resp, "result")
}

func addAttachmentToPlanEntry(cli client.ClientInterface, cmd *cobra.Command, planID int64, entryID, filePath string) error {
	ctx := cmd.Context()
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment plan-entry")
		dr.PrintSimple("Add Attachment to Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s, File: %s", planID, entryID, filePath))
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	resp, err := cli.AddAttachmentToPlanEntry(ctx, planID, entryID, filePath)
	if err != nil {
		return fmt.Errorf("failed to add attachment to plan entry: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Attachment added (ID: %d)", resp.AttachmentID)
		ui.Infof(os.Stdout, "URL: %s", resp.URL)
	}
	return output.OutputResult(cmd, resp, "result")
}

func addAttachmentToResult(cli client.ClientInterface, cmd *cobra.Command, resultID int64, filePath string) error {
	ctx := cmd.Context()
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment result")
		dr.PrintSimple("Add Attachment to Result", fmt.Sprintf("Result ID: %d, File: %s", resultID, filePath))
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	resp, err := cli.AddAttachmentToResult(ctx, resultID, filePath)
	if err != nil {
		return fmt.Errorf("failed to add attachment to result: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Attachment added (ID: %d)", resp.AttachmentID)
		ui.Infof(os.Stdout, "URL: %s", resp.URL)
	}
	return output.OutputResult(cmd, resp, "result")
}

func addAttachmentToRun(cli client.ClientInterface, cmd *cobra.Command, runID int64, filePath string) error {
	ctx := cmd.Context()
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment run")
		dr.PrintSimple("Add Attachment to Run", fmt.Sprintf("Run ID: %d, File: %s", runID, filePath))
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	resp, err := cli.AddAttachmentToRun(ctx, runID, filePath)
	if err != nil {
		return fmt.Errorf("failed to add attachment to run: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Attachment added (ID: %d)", resp.AttachmentID)
		ui.Infof(os.Stdout, "URL: %s", resp.URL)
	}
	return output.OutputResult(cmd, resp, "result")
}
