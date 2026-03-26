package cmd

import (
	"context"
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

// addCmd — команда для создания ресурсов через POST-запросы
var addCmd = &cobra.Command{
	Use:   "add <endpoint> [id]",
	Short: "Создать новый ресурс (POST-запрос)",
	Long: `Создаёт новый объект в TestRail через POST API.

Поддерживаемые эндпоинты:
  project              Создать проект
  suite <project_id>   Создать сьют in project
  section <project_id> Создать секцию in project/сьюте
  case <section_id>    Создать тест-кейс в секции
  run <project_id>     Создать тест-ран
  result <test_id>     Добавить результат теста
  result-for-case <run_id> <case_id>  Добавить результат для кейса
  shared-step <project_id>  Создать shared step
  milestone <project_id>    Создать milestone
  plan <project_id>         Создать test plan
  entry <plan_id>           Добавить entry в plan
  attachment case <case_id> <file>    Добавить вложение к кейсу
  attachment plan <plan_id> <file>    Добавить вложение к плану
  attachment plan-entry <plan_id> <entry_id> <file>  Добавить вложение к entry
  attachment result <result_id> <file>  Добавить вложение к результату
  attachment run <run_id> <file>      Добавить вложение к рану

Примеры:
  gotr add project --name "New Project" --announcement "Desc"
  gotr add suite 1 --name "Smoke Tests"
  gotr add case 100 --title "Login test" --template-id 1
  gotr add run 1 --name "Nightly Run" --suite-id 100
  gotr add result 12345 --status-id 1 --comment "Passed"
  gotr add attachment case 12345 ./screenshot.png
  gotr add attachment plan 100 ./report.pdf
  gotr add attachment result 98765 ./log.txt

Интерактивный режим (wizard):
  gotr add project -i
  gotr add suite 1 -i
  gotr add case 100 -i

Dry-run mode:
  gotr add project --name "Test" --dry-run  # Show what would be created`,
	RunE: runAdd,
}

func init() {
	// Общие флаги для создания
	addCmd.Flags().StringP("name", "n", "", "Название ресурса")
	addCmd.Flags().String("description", "", "Описание/announcement")
	addCmd.Flags().String("announcement", "", "Announcement (для проекта)")
	addCmd.Flags().Bool("show-announcement", false, "Показывать announcement")
	addCmd.Flags().Int64("suite-id", 0, "ID сьюта")
	addCmd.Flags().Int64("section-id", 0, "ID секции")
	addCmd.Flags().Int64("milestone-id", 0, "ID milestone")
	addCmd.Flags().Int64("template-id", 0, "ID шаблона (для case)")
	addCmd.Flags().Int64("type-id", 0, "ID типа (для case)")
	addCmd.Flags().Int64("priority-id", 0, "ID приоритета (для case)")
	addCmd.Flags().String("title", "", "Заголовок (для case)")
	addCmd.Flags().String("refs", "", "Ссылки (references)")
	addCmd.Flags().String("comment", "", "Комментарий (для result)")
	addCmd.Flags().Int64("status-id", 0, "ID статуса (для result)")
	addCmd.Flags().String("elapsed", "", "Время выполнения (для result)")
	addCmd.Flags().String("defects", "", "Дефекты (для result)")
	addCmd.Flags().Int64("assignedto-id", 0, "ID назначенного пользователя")
	addCmd.Flags().String("case-ids", "", "ID кейсов через запятую (для run)")
	addCmd.Flags().Bool("include-all", true, "Включить все кейсы (для run)")
	addCmd.Flags().String("json-file", "", "Путь к JSON-файлу с данными")
	output.AddFlag(addCmd)
	addCmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")
	addCmd.Flags().BoolP("interactive", "i", false, "Интерактивный режим (wizard)")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("endpoint required: project, suite, section, case, run, result, result-for-case, shared-step, milestone, plan, entry, attachment")
	}

	endpoint := args[0]
	ctx := cmd.Context()

	// Получаем клиент
	cli := GetClientInterface(cmd)

	// Определяем ID из аргументов (не для attachment - у него своя структура аргументов)
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

	// Читаем JSON из файла если указан
	jsonFile, _ := cmd.Flags().GetString("json-file")
	var jsonData []byte
	if jsonFile != "" {
		var err error
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("JSON file read error: %w", err)
		}
	}

	// Проверяем dry-run режим (не для attachment - у него свой dry-run внутри)
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun && endpoint != "attachment" {
		dr := output.NewDryRunPrinter("add " + endpoint)
		return runAddDryRun(cmd, dr, endpoint, id, jsonData)
	}

	// Проверяем интерактивный режим
	isInteractive, _ := cmd.Flags().GetBool("interactive")
	if isInteractive || shouldAutoRunAddInteractive(cmd, endpoint, id, jsonFile != "") {
		return runAddInteractive(cli, cmd, endpoint, id)
	}

	// Маршрутизация по endpoint
	switch endpoint {
	case "project":
		return addProject(cli, cmd, jsonData)
	case "suite":
		if id == 0 {
			return fmt.Errorf("project_id required: gotr add suite <project_id>")
		}
		return addSuite(cli, cmd, id, jsonData)
	case "section":
		if id == 0 {
			return fmt.Errorf("project_id required: gotr add section <project_id>")
		}
		return addSection(cli, cmd, id, jsonData)
	case "case":
		if id == 0 {
			return fmt.Errorf("section_id required: gotr add case <section_id>")
		}
		return addCase(cli, cmd, id, jsonData)
	case "run":
		if id == 0 {
			return fmt.Errorf("project_id required: gotr add run <project_id>")
		}
		return addRun(cli, cmd, id, jsonData)
	case "result":
		if id == 0 {
			return fmt.Errorf("test_id required: gotr add result <test_id>")
		}
		return addResult(cli, cmd, id, jsonData)
	case "result-for-case":
		if id == 0 || len(args) < 3 {
			return fmt.Errorf("run_id and case_id required: gotr add result-for-case <run_id> <case_id>")
		}
		caseID, err := flags.ValidateRequiredID(args, 2, "case_id")
		if err != nil {
			return err
		}
		return addResultForCase(cli, cmd, id, caseID, jsonData)
	case "shared-step":
		if id == 0 {
			return fmt.Errorf("project_id required: gotr add shared-step <project_id>")
		}
		return addSharedStep(cli, cmd, id, jsonData)
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

// runAddInteractive запускает интерактивный wizard для создания ресурса
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

	// Предпросмотр
	ui.Preview(os.Stdout, "Create Project", []ui.PreviewField{
		{Label: "Name", Value: answers.Name},
		{Label: "Announcement", Value: answers.Announcement},
		{Label: "Show announce", Value: answers.ShowAnnouncement},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Подтвердить создание?")
	if err != nil || !confirmed {
		ui.Cancelled(os.Stdout)
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

	// Предпросмотр
	ui.Preview(os.Stdout, "Create Suite", []ui.PreviewField{
		{Label: "Name", Value: answers.Name},
		{Label: "Description", Value: answers.Description},
		{Label: "Project ID", Value: projectID},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Подтвердить создание?")
	if err != nil || !confirmed {
		ui.Cancelled(os.Stdout)
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

	// Предпросмотр
	ui.Preview(os.Stdout, "Create Case", []ui.PreviewField{
		{Label: "Title", Value: answers.Title},
		{Label: "Section ID", Value: sectionID},
		{Label: "Type ID", Value: answers.TypeID},
		{Label: "Priority ID", Value: answers.PriorityID},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Подтвердить создание?")
	if err != nil || !confirmed {
		ui.Cancelled(os.Stdout)
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

	// Предпросмотр
	ui.Preview(os.Stdout, "Create Run", []ui.PreviewField{
		{Label: "Name", Value: answers.Name},
		{Label: "Description", Value: answers.Description},
		{Label: "Suite ID", Value: answers.SuiteID},
		{Label: "Include all", Value: answers.IncludeAll},
		{Label: "Project ID", Value: projectID},
	})

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Подтвердить создание?")
	if err != nil || !confirmed {
		ui.Cancelled(os.Stdout)
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

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Подтвердить создание?")
	if err != nil || !confirmed {
		ui.Cancelled(os.Stdout)
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

	confirmed, err := interactive.AskConfirmWithPrompter(p, "Подтвердить создание?")
	if err != nil || !confirmed {
		ui.Cancelled(os.Stdout)
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

// runAddDryRun выполняет dry-run для add команды
func runAddDryRun(cmd *cobra.Command, dr *output.DryRunPrinter, endpoint string, id int64, jsonData []byte) error {
	// Читаем флаги
	name, _ := cmd.Flags().GetString("name")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	announcement, _ := cmd.Flags().GetString("announcement")
	showAnn, _ := cmd.Flags().GetBool("show-announcement")
	suiteID, _ := cmd.Flags().GetInt64("suite-id")
	sectionID, _ := cmd.Flags().GetInt64("section-id")
	milestoneID, _ := cmd.Flags().GetInt64("milestone-id")
	templateID, _ := cmd.Flags().GetInt64("template-id")
	typeID, _ := cmd.Flags().GetInt64("type-id")
	priorityID, _ := cmd.Flags().GetInt64("priority-id")
	refs, _ := cmd.Flags().GetString("refs")
	comment, _ := cmd.Flags().GetString("comment")
	statusID, _ := cmd.Flags().GetInt64("status-id")
	elapsed, _ := cmd.Flags().GetString("elapsed")
	defects, _ := cmd.Flags().GetString("defects")
	assignedTo, _ := cmd.Flags().GetInt64("assignedto-id")
	caseIDsStr, _ := cmd.Flags().GetString("case-ids")
	includeAll, _ := cmd.Flags().GetBool("include-all")

	var method, url string
	var body interface{}

	switch endpoint {
	case "project":
		if len(jsonData) > 0 {
			var req data.AddProjectRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			body = data.AddProjectRequest{
				Name:             name,
				Announcement:     announcement,
				ShowAnnouncement: showAnn,
			}
		}
		method = "POST"
		url = "/index.php?/api/v2/add_project/"
		dr.PrintOperation("Create Project", method, url, body)

	case "suite":
		if id == 0 {
			return fmt.Errorf("project_id required: gotr add suite <project_id> --dry-run")
		}
		if len(jsonData) > 0 {
			var req data.AddSuiteRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			body = data.AddSuiteRequest{
				Name:        name,
				Description: description,
			}
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/add_suite/%d", id)
		dr.PrintOperation(fmt.Sprintf("Create Suite in Project %d", id), method, url, body)

	case "section":
		if id == 0 {
			return fmt.Errorf("project_id required: gotr add section <project_id> --dry-run")
		}
		if len(jsonData) > 0 {
			var req data.AddSectionRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			body = data.AddSectionRequest{
				Name:        name,
				Description: description,
				SuiteID:     suiteID,
				ParentID:    sectionID,
			}
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/add_section/%d", id)
		dr.PrintOperation(fmt.Sprintf("Create Section in Project %d", id), method, url, body)

	case "case":
		if id == 0 {
			return fmt.Errorf("section_id required: gotr add case <section_id> --dry-run")
		}
		if len(jsonData) > 0 {
			var req data.AddCaseRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			body = data.AddCaseRequest{
				Title:      title,
				TemplateID: templateID,
				TypeID:     typeID,
				PriorityID: priorityID,
				Refs:       refs,
			}
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/add_case/%d", id)
		dr.PrintOperation(fmt.Sprintf("Create Case in Section %d", id), method, url, body)

	case "run":
		if id == 0 {
			return fmt.Errorf("project_id required: gotr add run <project_id> --dry-run")
		}
		if len(jsonData) > 0 {
			var req data.AddRunRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			caseIDs := parseCaseIDs(caseIDsStr)
			body = data.AddRunRequest{
				Name:        name,
				Description: description,
				SuiteID:     suiteID,
				MilestoneID: milestoneID,
				AssignedTo:  assignedTo,
				IncludeAll:  includeAll,
				CaseIDs:     caseIDs,
			}
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/add_run/%d", id)
		dr.PrintOperation(fmt.Sprintf("Create Run in Project %d", id), method, url, body)

	case "result":
		if id == 0 {
			return fmt.Errorf("test_id required: gotr add result <test_id> --dry-run")
		}
		if len(jsonData) > 0 {
			var req data.AddResultRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			body = data.AddResultRequest{
				StatusID:   statusID,
				Comment:    comment,
				Elapsed:    elapsed,
				Defects:    defects,
				AssignedTo: assignedTo,
			}
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/add_result/%d", id)
		dr.PrintOperation(fmt.Sprintf("Add Result for Test %d", id), method, url, body)

	case "shared-step":
		if id == 0 {
			return fmt.Errorf("project_id required: gotr add shared-step <project_id> --dry-run")
		}
		if len(jsonData) > 0 {
			var req data.AddSharedStepRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			body = data.AddSharedStepRequest{
				Title: title,
			}
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/add_shared_step/%d", id)
		dr.PrintOperation(fmt.Sprintf("Create Shared Step in Project %d", id), method, url, body)

	case "attachment":
		// Для attachment dry-run обрабатывается отдельно в runAddAttachment
		// Этот case не должен вызываться напрямую
		return fmt.Errorf("use --dry-run with a specific attachment subcommand")

	default:
		return fmt.Errorf("unsupported endpoint for dry-run: %s", endpoint)
	}

	return nil
}

func addProject(cli client.ClientInterface, cmd *cobra.Command, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.AddProjectRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		req.Name = name
		req.Announcement, _ = cmd.Flags().GetString("announcement")
		req.ShowAnnouncement, _ = cmd.Flags().GetBool("show-announcement")
	}

	project, err := cli.AddProject(ctx, &req)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return output.OutputResult(cmd, project, "result")
}

func addSuite(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.AddSuiteRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		req.Name = name
		req.Description, _ = cmd.Flags().GetString("description")
	}

	suite, err := cli.AddSuite(ctx, projectID, &req)
	if err != nil {
		return fmt.Errorf("failed to create suite: %w", err)
	}

	return output.OutputResult(cmd, suite, "result")
}

func addSection(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.AddSectionRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		req.Name = name
		req.Description, _ = cmd.Flags().GetString("description")
		req.SuiteID, _ = cmd.Flags().GetInt64("suite-id")
		req.ParentID, _ = cmd.Flags().GetInt64("section-id")
	}

	section, err := cli.AddSection(ctx, projectID, &req)
	if err != nil {
		return fmt.Errorf("failed to create section: %w", err)
	}

	return output.OutputResult(cmd, section, "result")
}

func addCase(cli client.ClientInterface, cmd *cobra.Command, sectionID int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.AddCaseRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		title, _ := cmd.Flags().GetString("title")
		if title == "" {
			return fmt.Errorf("--title is required")
		}
		req.Title = title
		req.TemplateID, _ = cmd.Flags().GetInt64("template-id")
		req.TypeID, _ = cmd.Flags().GetInt64("type-id")
		req.PriorityID, _ = cmd.Flags().GetInt64("priority-id")
		req.Refs, _ = cmd.Flags().GetString("refs")
	}

	caseResp, err := cli.AddCase(ctx, sectionID, &req)
	if err != nil {
		return fmt.Errorf("failed to create case: %w", err)
	}

	return output.OutputResult(cmd, caseResp, "result")
}

func addRun(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.AddRunRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		req.Name = name
		req.Description, _ = cmd.Flags().GetString("description")
		req.SuiteID, _ = cmd.Flags().GetInt64("suite-id")
		req.MilestoneID, _ = cmd.Flags().GetInt64("milestone-id")
		req.AssignedTo, _ = cmd.Flags().GetInt64("assignedto-id")
		req.IncludeAll, _ = cmd.Flags().GetBool("include-all")

		caseIDsStr, _ := cmd.Flags().GetString("case-ids")
		if caseIDsStr != "" {
			req.CaseIDs = parseCaseIDs(caseIDsStr)
		}
	}

	run, err := cli.AddRun(ctx, projectID, &req)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	return output.OutputResult(cmd, run, "result")
}

func addResult(cli client.ClientInterface, cmd *cobra.Command, testID int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.AddResultRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		statusID, _ := cmd.Flags().GetInt64("status-id")
		if statusID == 0 {
			return fmt.Errorf("--status-id is required")
		}
		req.StatusID = statusID
		req.Comment, _ = cmd.Flags().GetString("comment")
		req.Elapsed, _ = cmd.Flags().GetString("elapsed")
		req.Defects, _ = cmd.Flags().GetString("defects")
		req.AssignedTo, _ = cmd.Flags().GetInt64("assignedto-id")
		req.Version, _ = cmd.Flags().GetString("version")
	}

	result, err := cli.AddResult(ctx, testID, &req)
	if err != nil {
		return fmt.Errorf("failed to add result: %w", err)
	}

	return output.OutputResult(cmd, result, "result")
}

func addResultForCase(cli client.ClientInterface, cmd *cobra.Command, runID, caseID int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.AddResultRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		statusID, _ := cmd.Flags().GetInt64("status-id")
		if statusID == 0 {
			return fmt.Errorf("--status-id is required")
		}
		req.StatusID = statusID
		req.Comment, _ = cmd.Flags().GetString("comment")
		req.Elapsed, _ = cmd.Flags().GetString("elapsed")
		req.Defects, _ = cmd.Flags().GetString("defects")
		req.AssignedTo, _ = cmd.Flags().GetInt64("assignedto-id")
	}

	result, err := cli.AddResultForCase(ctx, runID, caseID, &req)
	if err != nil {
		return fmt.Errorf("failed to add result: %w", err)
	}

	return output.OutputResult(cmd, result, "result")
}

func addSharedStep(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	ctx := cmd.Context()
	var req data.AddSharedStepRequest

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		title, _ := cmd.Flags().GetString("title")
		if title == "" {
			return fmt.Errorf("--title is required")
		}
		req.Title = title
	}

	step, err := cli.AddSharedStep(ctx, projectID, &req)
	if err != nil {
		return fmt.Errorf("failed to create shared step: %w", err)
	}

	return output.OutputResult(cmd, step, "result")
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

// runAddAttachment обрабатывает добавление вложений (DEPRECATED: use 'gotr attachments add' instead)
func runAddAttachment(cli client.ClientInterface, cmd *cobra.Command, args []string) error {
	fmt.Fprintln(os.Stderr, "⚠️  WARNING: 'gotr add attachment' is deprecated. Use 'gotr attachments add' instead.")

	if len(args) < 2 {
		return fmt.Errorf("attachment type required: case, plan, plan-entry, result, run")
	}

	attachmentType := args[1]

	switch attachmentType {
	case "case":
		if len(args) < 4 {
			return fmt.Errorf("usage: gotr add attachment case <case_id> <file_path>")
		}
		caseID, err := flags.ValidateRequiredID(args, 2, "case_id")
		if err != nil {
			return err
		}
		filePath := args[3]
		return addAttachmentToCase(cli, cmd, caseID, filePath)

	case "plan":
		if len(args) < 4 {
			return fmt.Errorf("usage: gotr add attachment plan <plan_id> <file_path>")
		}
		planID, err := flags.ValidateRequiredID(args, 2, "plan_id")
		if err != nil {
			return err
		}
		filePath := args[3]
		return addAttachmentToPlan(cli, cmd, planID, filePath)

	case "plan-entry":
		if len(args) < 5 {
			return fmt.Errorf("usage: gotr add attachment plan-entry <plan_id> <entry_id> <file_path>")
		}
		planID, err := flags.ValidateRequiredID(args, 2, "plan_id")
		if err != nil {
			return err
		}
		entryID := args[3]
		filePath := args[4]
		return addAttachmentToPlanEntry(cli, cmd, planID, entryID, filePath)

	case "result":
		if len(args) < 4 {
			return fmt.Errorf("usage: gotr add attachment result <result_id> <file_path>")
		}
		resultID, err := flags.ValidateRequiredID(args, 2, "result_id")
		if err != nil {
			return err
		}
		filePath := args[3]
		return addAttachmentToResult(cli, cmd, resultID, filePath)

	case "run":
		if len(args) < 4 {
			return fmt.Errorf("usage: gotr add attachment run <run_id> <file_path>")
		}
		runID, err := flags.ValidateRequiredID(args, 2, "run_id")
		if err != nil {
			return err
		}
		filePath := args[3]
		return addAttachmentToRun(cli, cmd, runID, filePath)

	default:
		return fmt.Errorf("unsupported attachment type: %s. Available: case, plan, plan-entry, result, run", attachmentType)
	}
}

func addAttachmentToCase(cli client.ClientInterface, cmd *cobra.Command, caseID int64, filePath string) error {
	ctx := cmd.Context()
	// Проверяем dry-run режим
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment case")
		dr.PrintSimple("Add Attachment to Case", fmt.Sprintf("Case ID: %d, File: %s", caseID, filePath))
		return nil
	}

	// Проверяем существование файла
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
