package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// updateCmd updates resources via POST requests.
var updateCmd = &cobra.Command{
	Use:   "update <endpoint> <id>",
	Short: "Update an existing resource (POST request)",
	Long: `Updates an existing object in TestRail via the POST API.

Supported endpoints:
  project <id>       Update a project
  suite <id>         Update a suite
  section <id>       Update a section
  case <id>          Update a test case
  run <id>           Update a test run
  shared-step <id>   Update a shared step
  milestone <id>     Update a milestone
  plan <id>          Update a test plan
  labels <test_id>   Update test labels (deprecated: use 'gotr labels update')

Examples:
  gotr update project 1 --name "Updated Project"
  gotr update suite 100 --name "Updated Suite"
  gotr update case 12345 --title "Updated Title" --priority-id 2
  gotr update run 1000 --name "Updated Run Name"
  gotr update shared-step 50 --title "Updated Step"

Interactive mode (wizard):
  gotr update project 1 -i
  gotr update suite 100 -i
  gotr update case 12345 -i

Dry-run mode:
  gotr update project 1 --name "Test" --dry-run  # Show what would be updated`,
	RunE: runUpdate,
}

func init() {
	// Common flags for updating
	updateCmd.Flags().StringP("name", "n", "", "Resource name")
	updateCmd.Flags().String("description", "", "Description")
	updateCmd.Flags().String("announcement", "", "Announcement (for project)")
	updateCmd.Flags().Bool("show-announcement", false, "Show announcement")
	updateCmd.Flags().Bool("is-completed", false, "Mark as completed")
	updateCmd.Flags().String("title", "", "Title (for case)")
	updateCmd.Flags().Int64("type-id", 0, "Type ID (for case)")
	updateCmd.Flags().Int64("priority-id", 0, "Priority ID (for case)")
	updateCmd.Flags().String("refs", "", "References")
	updateCmd.Flags().Int64("suite-id", 0, "Suite ID")
	updateCmd.Flags().Int64("milestone-id", 0, "Milestone ID")
	updateCmd.Flags().Int64("assignedto-id", 0, "Assigned user ID")
	updateCmd.Flags().String("case-ids", "", "Comma-separated case IDs (for run)")
	updateCmd.Flags().Bool("include-all", false, "Include all cases (for run)")
	updateCmd.Flags().String("json-file", "", "Path to JSON data file")
	output.AddFlag(updateCmd)
	updateCmd.Flags().Bool("dry-run", false, "Show what would be executed without making changes")
	updateCmd.Flags().BoolP("interactive", "i", false, "Interactive mode (wizard)")

	// Flags for labels
	updateCmd.Flags().String("labels", "", "Comma-separated labels for test (for 'update labels')")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("endpoint and id required: gotr update <endpoint> <id>")
	}

	endpoint := args[0]
	id, err := flags.ValidateRequiredID(args, 1, "ID")
	if err != nil {
		return err
	}

	// Get the client
	cli := GetClientInterface(cmd)

	// Read JSON from file if specified
	jsonFile, _ := cmd.Flags().GetString("json-file")
	var jsonData []byte
	if jsonFile != "" {
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("JSON file read error: %w", err)
		}
	}

	// Check dry-run mode
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("update " + endpoint)
		return runUpdateDryRun(cmd, dr, endpoint, id, jsonData)
	}

	// Check interactive mode
	isInteractive, _ := cmd.Flags().GetBool("interactive")
	if isInteractive || shouldAutoRunUpdateInteractive(cmd, endpoint, jsonFile != "") {
		return runUpdateInteractive(cli, cmd, endpoint, id)
	}

	// Route by endpoint
	switch endpoint {
	case "project":
		return updateProject(cli, cmd, id, jsonData)
	case "suite":
		return updateSuite(cli, cmd, id, jsonData)
	case "section":
		return updateSection(cli, cmd, id, jsonData)
	case "case":
		return updateCase(cli, cmd, id, jsonData)
	case "run":
		return updateRun(cli, cmd, id, jsonData)
	case "shared-step":
		return updateSharedStep(cli, cmd, id, jsonData)
	case "labels":
		return updateLabels(cli, cmd, id)
	default:
		return fmt.Errorf("unsupported endpoint: %s", endpoint)
	}
}

func shouldAutoRunUpdateInteractive(cmd *cobra.Command, endpoint string, hasJSONFile bool) bool {
	if hasJSONFile || !interactive.HasPrompterInContext(cmd.Context()) {
		return false
	}

	switch endpoint {
	case "project":
		return !hasAnyChangedFlag(cmd, "name", "announcement", "show-announcement", "is-completed")
	case "suite":
		return !hasAnyChangedFlag(cmd, "name", "description", "is-completed")
	case "section":
		return !hasAnyChangedFlag(cmd, "name", "description")
	case "case":
		return !hasAnyChangedFlag(cmd, "title", "type-id", "priority-id", "refs")
	case "run":
		return !hasAnyChangedFlag(cmd, "name", "description", "milestone-id", "assignedto-id", "include-all", "case-ids")
	case "shared-step":
		return !hasAnyChangedFlag(cmd, "title")
	default:
		return false
	}
}

// runUpdateInteractive starts an interactive wizard for updating a resource.
func runUpdateInteractive(cli client.ClientInterface, cmd *cobra.Command, endpoint string, id int64) error {
	switch endpoint {
	case "project":
		return updateProjectInteractive(cli, cmd, id)
	case "suite":
		return updateSuiteInteractive(cli, cmd, id)
	case "section":
		return updateSectionInteractive(cli, cmd, id)
	case "case":
		return updateCaseInteractive(cli, cmd, id)
	case "run":
		return updateRunInteractive(cli, cmd, id)
	case "shared-step":
		return updateSharedStepInteractive(cli, cmd, id)
	default:
		return fmt.Errorf("interactive mode not supported for endpoint: %s", endpoint)
	}
}

func updateSectionInteractive(cli client.ClientInterface, cmd *cobra.Command, id int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)

	name, err := p.Input("Section name (optional):", "")
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}
	description, err := p.MultilineInput("Description (optional):", "")
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}
	if name == "" && description == "" {
		return fmt.Errorf("at least one field is required for update")
	}

	ui.Preview(os.Stdout, "Update Section", []ui.PreviewField{
		{Label: "Section ID", Value: id},
		{Label: "Name", Value: name},
		{Label: "Description", Value: description},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm update?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.UpdateSectionRequest{}
	if name != "" {
		req.Name = name
	}
	if description != "" {
		req.Description = description
	}

	section, err := cli.UpdateSection(ctx, id, req)
	if err != nil {
		return fmt.Errorf("failed to update section: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Section updated (ID: %d)", section.ID)
	}
	return outputUpdateResult(cmd, section)
}

func updateProjectInteractive(cli client.ClientInterface, cmd *cobra.Command, id int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)
	answers, err := interactive.AskProjectWithPrompter(p, true)
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	// Preview
	ui.Preview(os.Stdout, "Update Project", []ui.PreviewField{
		{Label: "Project ID", Value: id},
		{Label: "Name", Value: answers.Name},
		{Label: "Announcement", Value: answers.Announcement},
		{Label: "Show announce", Value: answers.ShowAnnouncement},
		{Label: "Is completed", Value: answers.IsCompleted},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm update?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.UpdateProjectRequest{
		Name:             answers.Name,
		Announcement:     answers.Announcement,
		ShowAnnouncement: answers.ShowAnnouncement,
		IsCompleted:      answers.IsCompleted,
	}

	project, err := cli.UpdateProject(ctx, id, req)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Project updated (ID: %d)", project.ID)
	}
	return outputUpdateResult(cmd, project)
}

func updateSuiteInteractive(cli client.ClientInterface, cmd *cobra.Command, id int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)
	answers, err := interactive.AskSuiteWithPrompter(p, true)
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	// Preview
	ui.Preview(os.Stdout, "Update Suite", []ui.PreviewField{
		{Label: "Suite ID", Value: id},
		{Label: "Name", Value: answers.Name},
		{Label: "Description", Value: answers.Description},
		{Label: "Is completed", Value: answers.IsCompleted},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm update?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.UpdateSuiteRequest{
		Name:        answers.Name,
		Description: answers.Description,
		IsCompleted: answers.IsCompleted,
	}

	suite, err := cli.UpdateSuite(ctx, id, req)
	if err != nil {
		return fmt.Errorf("failed to update suite: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Suite updated (ID: %d)", suite.ID)
	}
	return outputUpdateResult(cmd, suite)
}

func updateCaseInteractive(cli client.ClientInterface, cmd *cobra.Command, id int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)
	answers, err := interactive.AskCaseWithPrompter(p, true)
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	// Preview
	ui.Preview(os.Stdout, "Update Case", []ui.PreviewField{
		{Label: "Case ID", Value: id},
		{Label: "Title", Value: answers.Title},
		{Label: "Type ID", Value: answers.TypeID},
		{Label: "Priority ID", Value: answers.PriorityID},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm update?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.UpdateCaseRequest{
		Title:      &answers.Title,
		TypeID:     &answers.TypeID,
		PriorityID: &answers.PriorityID,
		Refs:       &answers.Refs,
	}

	caseResp, err := cli.UpdateCase(ctx, id, req)
	if err != nil {
		return fmt.Errorf("failed to update case: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Case updated (ID: %d)", caseResp.ID)
	}
	return outputUpdateResult(cmd, caseResp)
}

func updateRunInteractive(cli client.ClientInterface, cmd *cobra.Command, id int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)

	answers, err := interactive.AskRunWithPrompter(p, true)
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	ui.Preview(os.Stdout, "Update Run", []ui.PreviewField{
		{Label: "Run ID", Value: id},
		{Label: "Name", Value: answers.Name},
		{Label: "Description", Value: answers.Description},
		{Label: "Include all", Value: answers.IncludeAll},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm update?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.UpdateRunRequest{IncludeAll: &answers.IncludeAll}
	if answers.Name != "" {
		req.Name = &answers.Name
	}
	if answers.Description != "" {
		req.Description = &answers.Description
	}

	run, err := cli.UpdateRun(ctx, id, req)
	if err != nil {
		return fmt.Errorf("failed to update run: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Run updated (ID: %d)", run.ID)
	}
	return outputUpdateResult(cmd, run)
}

func updateSharedStepInteractive(cli client.ClientInterface, cmd *cobra.Command, id int64) error {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)

	title, err := p.Input("Shared step title:", "")
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}
	if title == "" {
		return fmt.Errorf("shared step title is required")
	}

	ui.Preview(os.Stdout, "Update Shared Step", []ui.PreviewField{
		{Label: "Shared Step ID", Value: id},
		{Label: "Title", Value: title},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Confirm update?")
	if err != nil || !confirmed {
		ui.Canceled(os.Stdout)
		return nil
	}

	req := &data.UpdateSharedStepRequest{Title: title}
	step, err := cli.UpdateSharedStep(ctx, id, req)
	if err != nil {
		return fmt.Errorf("failed to update shared step: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Shared step updated (ID: %d)", step.ID)
	}
	return outputUpdateResult(cmd, step)
}

// runUpdateDryRun performs a dry-run for the update command.
func runUpdateDryRun(cmd *cobra.Command, dr *output.DryRunPrinter, endpoint string, id int64, jsonData []byte) error {
	handler, ok := updateDryRunHandlers[endpoint]
	if !ok {
		return fmt.Errorf("unsupported endpoint for dry-run: %s", endpoint)
	}
	return handler(cmd, dr, id, jsonData)
}

var updateDryRunHandlers = map[string]func(*cobra.Command, *output.DryRunPrinter, int64, []byte) error{
	"project":     dryRunUpdateProject,
	"suite":       dryRunUpdateSuite,
	"section":     dryRunUpdateSection,
	"case":        dryRunUpdateCase,
	"run":         dryRunUpdateRun,
	"shared-step": dryRunUpdateSharedStep,
	"labels":      dryRunUpdateLabels,
}

func dryRunUpdateProject(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	var body interface{}
	if len(jsonData) > 0 {
		var req data.UpdateProjectRequest
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
		body = req
	} else {
		name, _ := cmd.Flags().GetString("name")
		announcement, _ := cmd.Flags().GetString("announcement")
		showAnn, _ := cmd.Flags().GetBool("show-announcement")
		isCompleted, _ := cmd.Flags().GetBool("is-completed")
		req := data.UpdateProjectRequest{ShowAnnouncement: showAnn, IsCompleted: isCompleted}
		if name != "" {
			req.Name = name
		}
		if announcement != "" {
			req.Announcement = announcement
		}
		body = req
	}
	dr.PrintOperation(fmt.Sprintf("Update Project %d", id), "POST", fmt.Sprintf("/index.php?/api/v2/update_project/%d", id), body)
	return nil
}

func dryRunUpdateSuite(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	var body interface{}
	if len(jsonData) > 0 {
		var req data.UpdateSuiteRequest
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
		body = req
	} else {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		isCompleted, _ := cmd.Flags().GetBool("is-completed")
		req := data.UpdateSuiteRequest{IsCompleted: isCompleted}
		if name != "" {
			req.Name = name
		}
		if description != "" {
			req.Description = description
		}
		body = req
	}
	dr.PrintOperation(fmt.Sprintf("Update Suite %d", id), "POST", fmt.Sprintf("/index.php?/api/v2/update_suite/%d", id), body)
	return nil
}

func dryRunUpdateSection(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	var body interface{}
	if len(jsonData) > 0 {
		var req data.UpdateSectionRequest
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
		body = req
	} else {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		req := data.UpdateSectionRequest{}
		if name != "" {
			req.Name = name
		}
		if description != "" {
			req.Description = description
		}
		body = req
	}
	dr.PrintOperation(fmt.Sprintf("Update Section %d", id), "POST", fmt.Sprintf("/index.php?/api/v2/update_section/%d", id), body)
	return nil
}

func dryRunUpdateCase(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	var body interface{}
	if len(jsonData) > 0 {
		var req data.UpdateCaseRequest
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
		body = req
	} else {
		title, _ := cmd.Flags().GetString("title")
		typeID, _ := cmd.Flags().GetInt64("type-id")
		priorityID, _ := cmd.Flags().GetInt64("priority-id")
		refs, _ := cmd.Flags().GetString("refs")
		req := data.UpdateCaseRequest{}
		if title != "" {
			req.Title = &title
		}
		if typeID > 0 {
			req.TypeID = &typeID
		}
		if priorityID > 0 {
			req.PriorityID = &priorityID
		}
		if refs != "" {
			req.Refs = &refs
		}
		body = req
	}
	dr.PrintOperation(fmt.Sprintf("Update Case %d", id), "POST", fmt.Sprintf("/index.php?/api/v2/update_case/%d", id), body)
	return nil
}

func dryRunUpdateRun(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	var body interface{}
	if len(jsonData) > 0 {
		var req data.UpdateRunRequest
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
		body = req
	} else {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		milestoneID, _ := cmd.Flags().GetInt64("milestone-id")
		assignedToID, _ := cmd.Flags().GetInt64("assignedto-id")
		includeAll, _ := cmd.Flags().GetBool("include-all")
		caseIDsStr, _ := cmd.Flags().GetString("case-ids")
		req := data.UpdateRunRequest{IncludeAll: &includeAll}
		if name != "" {
			req.Name = &name
		}
		if description != "" {
			req.Description = &description
		}
		if milestoneID > 0 {
			req.MilestoneID = &milestoneID
		}
		if assignedToID > 0 {
			req.AssignedTo = &assignedToID
		}
		if caseIDsStr != "" {
			caseIDs := parseCaseIDs(caseIDsStr)
			req.CaseIDs = caseIDs
		}
		body = req
	}
	dr.PrintOperation(fmt.Sprintf("Update Run %d", id), "POST", fmt.Sprintf("/index.php?/api/v2/update_run/%d", id), body)
	return nil
}

func dryRunUpdateSharedStep(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, jsonData []byte) error {
	var body interface{}
	if len(jsonData) > 0 {
		var req data.UpdateSharedStepRequest
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
		body = req
	} else {
		title, _ := cmd.Flags().GetString("title")
		req := data.UpdateSharedStepRequest{}
		if title != "" {
			req.Title = title
		}
		body = req
	}
	dr.PrintOperation(fmt.Sprintf("Update Shared Step %d", id), "POST", fmt.Sprintf("/index.php?/api/v2/update_shared_step/%d", id), body)
	return nil
}

func dryRunUpdateLabels(cmd *cobra.Command, dr *output.DryRunPrinter, id int64, _ []byte) error {
	labels, _ := cmd.Flags().GetString("labels")
	dr.PrintSimple("Update Test Labels", fmt.Sprintf("Test ID: %d, Labels: %s", id, labels))
	return nil
}

func updateProject(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.UpdateProjectRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			req.Name = name
		}
		announcement, _ := cmd.Flags().GetString("announcement")
		if announcement != "" {
			req.Announcement = announcement
		}
		req.ShowAnnouncement, _ = cmd.Flags().GetBool("show-announcement")
		req.IsCompleted, _ = cmd.Flags().GetBool("is-completed")
	}

	project, err := cli.UpdateProject(ctx, id, &req)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return outputUpdateResult(cmd, project)
}

func updateSuite(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.UpdateSuiteRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			req.Name = name
		}
		description, _ := cmd.Flags().GetString("description")
		if description != "" {
			req.Description = description
		}
		req.IsCompleted, _ = cmd.Flags().GetBool("is-completed")
	}

	suite, err := cli.UpdateSuite(ctx, id, &req)
	if err != nil {
		return fmt.Errorf("failed to update suite: %w", err)
	}

	return outputUpdateResult(cmd, suite)
}

func updateSection(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.UpdateSectionRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			req.Name = name
		}
		description, _ := cmd.Flags().GetString("description")
		if description != "" {
			req.Description = description
		}
	}

	section, err := cli.UpdateSection(ctx, id, &req)
	if err != nil {
		return fmt.Errorf("failed to update section: %w", err)
	}

	return outputUpdateResult(cmd, section)
}

func updateCase(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.UpdateCaseRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		title, _ := cmd.Flags().GetString("title")
		if title != "" {
			req.Title = &title
		}
		typeID, _ := cmd.Flags().GetInt64("type-id")
		if typeID > 0 {
			req.TypeID = &typeID
		}
		priorityID, _ := cmd.Flags().GetInt64("priority-id")
		if priorityID > 0 {
			req.PriorityID = &priorityID
		}
		refs, _ := cmd.Flags().GetString("refs")
		if refs != "" {
			req.Refs = &refs
		}
	}

	caseResp, err := cli.UpdateCase(ctx, id, &req)
	if err != nil {
		return fmt.Errorf("failed to update case: %w", err)
	}

	return outputUpdateResult(cmd, caseResp)
}

func updateRun(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.UpdateRunRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			req.Name = &name
		}
		description, _ := cmd.Flags().GetString("description")
		if description != "" {
			req.Description = &description
		}
		milestoneID, _ := cmd.Flags().GetInt64("milestone-id")
		if milestoneID > 0 {
			req.MilestoneID = &milestoneID
		}
		assignedToID, _ := cmd.Flags().GetInt64("assignedto-id")
		if assignedToID > 0 {
			req.AssignedTo = &assignedToID
		}
		includeAll, _ := cmd.Flags().GetBool("include-all")
		req.IncludeAll = &includeAll

		caseIDsStr, _ := cmd.Flags().GetString("case-ids")
		if caseIDsStr != "" {
			req.CaseIDs = parseCaseIDs(caseIDsStr)
		}
	}

	run, err := cli.UpdateRun(ctx, id, &req)
	if err != nil {
		return fmt.Errorf("failed to update run: %w", err)
	}

	return outputUpdateResult(cmd, run)
}

func updateSharedStep(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.UpdateSharedStepRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		title, _ := cmd.Flags().GetString("title")
		if title != "" {
			req.Title = title
		}
	}

	step, err := cli.UpdateSharedStep(ctx, id, &req)
	if err != nil {
		return fmt.Errorf("failed to update shared step: %w", err)
	}

	return outputUpdateResult(cmd, step)
}

func outputUpdateResult(cmd *cobra.Command, v interface{}) error {
	_, err := output.Output(cmd, v, "result", "json")
	return err
}

// updateLabels updates test labels (DEPRECATED: use 'gotr labels update' instead).
func updateLabels(cli client.ClientInterface, cmd *cobra.Command, testID int64) error {
	ctx := cmd.Context()
	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		fmt.Fprintln(os.Stderr, "⚠️  WARNING: 'gotr update labels' is deprecated. Use 'gotr labels update test' instead.")
	}

	labelsFlag, _ := cmd.Flags().GetString("labels")
	if labelsFlag == "" {
		return fmt.Errorf("--labels is required")
	}

	// Parse labels
	labels := parseLabels(labelsFlag)
	if len(labels) == 0 {
		return fmt.Errorf("labels not specified")
	}

	// Check dry-run
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("update labels")
		dr.PrintSimple("Update Test Labels", fmt.Sprintf("Test ID: %d, Labels: %v", testID, labels))
		return nil
	}

	if err := cli.UpdateTestLabels(ctx, testID, labels); err != nil {
		return fmt.Errorf("failed to update labels: %w", err)
	}

	if quiet, _ := cmd.Flags().GetBool("quiet"); !quiet {
		ui.Successf(os.Stdout, "Labels updated for test %d: %v", testID, labels)
	}
	return nil
}

// parseLabels parses a comma-separated string of labels.
func parseLabels(s string) []string {
	var labels []string
	for _, part := range splitAndTrim(s, ",") {
		if part != "" {
			labels = append(labels, part)
		}
	}
	return labels
}
