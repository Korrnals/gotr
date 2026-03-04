package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// addCmd — команда для создания ресурсов через POST-запросы
var addCmd = &cobra.Command{
	Use:   "add <endpoint> [id]",
	Short: "Создать новый ресурс (POST-запрос)",
	Long: `Создаёт новый объект в TestRail через POST API.

Поддерживаемые эндпоинты:
  project              Создать проект
  suite <project_id>   Создать сьют в проекте
  section <project_id> Создать секцию в проекте/сьюте
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
		return fmt.Errorf("необходимо указать endpoint: project, suite, section, case, run, result, result-for-case, shared-step, milestone, plan, entry, attachment")
	}

	endpoint := args[0]
	
	// Получаем клиент
	cli := GetClientInterface(cmd)

	// Определяем ID из аргументов (не для attachment - у него своя структура аргументов)
	var id int64
	if len(args) > 1 && endpoint != "attachment" {
		parsedID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("неверный ID: %w", err)
		}
		id = parsedID
	}

	// Читаем JSON из файла если указан
	jsonFile, _ := cmd.Flags().GetString("json-file")
	var jsonData []byte
	if jsonFile != "" {
		var err error
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("ошибка чтения JSON-файла: %w", err)
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
	if isInteractive {
		return runAddInteractive(cli, cmd, endpoint, id)
	}

	// Маршрутизация по endpoint
	switch endpoint {
	case "project":
		return addProject(cli, cmd, jsonData)
	case "suite":
		if id == 0 {
			return fmt.Errorf("необходимо указать project_id: gotr add suite <project_id>")
		}
		return addSuite(cli, cmd, id, jsonData)
	case "section":
		if id == 0 {
			return fmt.Errorf("необходимо указать project_id: gotr add section <project_id>")
		}
		return addSection(cli, cmd, id, jsonData)
	case "case":
		if id == 0 {
			return fmt.Errorf("необходимо указать section_id: gotr add case <section_id>")
		}
		return addCase(cli, cmd, id, jsonData)
	case "run":
		if id == 0 {
			return fmt.Errorf("необходимо указать project_id: gotr add run <project_id>")
		}
		return addRun(cli, cmd, id, jsonData)
	case "result":
		if id == 0 {
			return fmt.Errorf("необходимо указать test_id: gotr add result <test_id>")
		}
		return addResult(cli, cmd, id, jsonData)
	case "result-for-case":
		if id == 0 || len(args) < 3 {
			return fmt.Errorf("необходимо указать run_id и case_id: gotr add result-for-case <run_id> <case_id>")
		}
		caseID, _ := strconv.ParseInt(args[2], 10, 64)
		return addResultForCase(cli, cmd, id, caseID, jsonData)
	case "shared-step":
		if id == 0 {
			return fmt.Errorf("необходимо указать project_id: gotr add shared-step <project_id>")
		}
		return addSharedStep(cli, cmd, id, jsonData)
	case "attachment":
		return runAddAttachment(cli, cmd, args)
	default:
		return fmt.Errorf("неподдерживаемый endpoint: %s", endpoint)
	}
}

// runAddInteractive запускает интерактивный wizard для создания ресурса
func runAddInteractive(cli client.ClientInterface, cmd *cobra.Command, endpoint string, parentID int64) error {
	switch endpoint {
	case "project":
		return addProjectInteractive(cli, cmd)
	case "suite":
		if parentID == 0 {
			return fmt.Errorf("необходимо указать project_id: gotr add suite <project_id> --interactive")
		}
		return addSuiteInteractive(cli, cmd, parentID)
	case "case":
		if parentID == 0 {
			return fmt.Errorf("необходимо указать section_id: gotr add case <section_id> --interactive")
		}
		return addCaseInteractive(cli, cmd, parentID)
	case "run":
		if parentID == 0 {
			return fmt.Errorf("необходимо указать project_id: gotr add run <project_id> --interactive")
		}
		return addRunInteractive(cli, cmd, parentID)
	default:
		return fmt.Errorf("интерактивный режим не поддерживается для endpoint: %s", endpoint)
	}
}

func addProjectInteractive(cli client.ClientInterface, cmd *cobra.Command) error {
	answers, err := interactive.AskProject(false)
	if err != nil {
		return fmt.Errorf("ошибка ввода: %w", err)
	}

	// Предпросмотр
	fmt.Println("\n────────────────────────────────────────────────────────────")
	fmt.Println("📋 ПРЕДПРОСМОТР: Create Project")
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("Название:        %s\n", answers.Name)
	fmt.Printf("Announcement:    %s\n", answers.Announcement)
	fmt.Printf("Show announce:   %v\n", answers.ShowAnnouncement)
	fmt.Println("────────────────────────────────────────────────────────────")

	confirmed, err := interactive.AskConfirm("Подтвердить создание?")
	if err != nil || !confirmed {
		fmt.Println("\n❌ Отменено")
		return nil
	}

	req := &data.AddProjectRequest{
		Name:             answers.Name,
		Announcement:     answers.Announcement,
		ShowAnnouncement: answers.ShowAnnouncement,
	}

	project, err := cli.AddProject(req)
	if err != nil {
		return fmt.Errorf("ошибка создания проекта: %w", err)
	}

	fmt.Printf("\n✅ Проект создан (ID: %d)\n", project.ID)
	return output.OutputResult(cmd, project, "result")
}

func addSuiteInteractive(cli client.ClientInterface, cmd *cobra.Command, projectID int64) error {
	answers, err := interactive.AskSuite(false)
	if err != nil {
		return fmt.Errorf("ошибка ввода: %w", err)
	}

	// Предпросмотр
	fmt.Println("\n────────────────────────────────────────────────────────────")
	fmt.Println("📋 ПРЕДПРОСМОТР: Create Suite")
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("Название:        %s\n", answers.Name)
	fmt.Printf("Описание:        %s\n", answers.Description)
	fmt.Printf("Project ID:      %d\n", projectID)
	fmt.Println("────────────────────────────────────────────────────────────")

	confirmed, err := interactive.AskConfirm("Подтвердить создание?")
	if err != nil || !confirmed {
		fmt.Println("\n❌ Отменено")
		return nil
	}

	req := &data.AddSuiteRequest{
		Name:        answers.Name,
		Description: answers.Description,
	}

	suite, err := cli.AddSuite(projectID, req)
	if err != nil {
		return fmt.Errorf("ошибка создания сьюта: %w", err)
	}

	fmt.Printf("\n✅ Сьют создан (ID: %d)\n", suite.ID)
	return output.OutputResult(cmd, suite, "result")
}

func addCaseInteractive(cli client.ClientInterface, cmd *cobra.Command, sectionID int64) error {
	answers, err := interactive.AskCase(false)
	if err != nil {
		return fmt.Errorf("ошибка ввода: %w", err)
	}

	// Предпросмотр
	fmt.Println("\n────────────────────────────────────────────────────────────")
	fmt.Println("📋 ПРЕДПРОСМОТР: Create Case")
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("Заголовок:       %s\n", answers.Title)
	fmt.Printf("Section ID:      %d\n", sectionID)
	fmt.Printf("Type ID:         %d\n", answers.TypeID)
	fmt.Printf("Priority ID:     %d\n", answers.PriorityID)
	fmt.Println("────────────────────────────────────────────────────────────")

	confirmed, err := interactive.AskConfirm("Подтвердить создание?")
	if err != nil || !confirmed {
		fmt.Println("\n❌ Отменено")
		return nil
	}

	req := &data.AddCaseRequest{
		Title:      answers.Title,
		SectionID:  sectionID,
		TypeID:     answers.TypeID,
		PriorityID: answers.PriorityID,
		Refs:       answers.Refs,
	}

	caseResp, err := cli.AddCase(sectionID, req)
	if err != nil {
		return fmt.Errorf("ошибка создания кейса: %w", err)
	}

	fmt.Printf("\n✅ Кейс создан (ID: %d)\n", caseResp.ID)
	return output.OutputResult(cmd, caseResp, "result")
}

func addRunInteractive(cli client.ClientInterface, cmd *cobra.Command, projectID int64) error {
	answers, err := interactive.AskRun(false)
	if err != nil {
		return fmt.Errorf("ошибка ввода: %w", err)
	}

	// Предпросмотр
	fmt.Println("\n────────────────────────────────────────────────────────────")
	fmt.Println("📋 ПРЕДПРОСМОТР: Create Run")
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("Название:        %s\n", answers.Name)
	fmt.Printf("Описание:        %s\n", answers.Description)
	fmt.Printf("Suite ID:        %d\n", answers.SuiteID)
	fmt.Printf("Include all:     %v\n", answers.IncludeAll)
	fmt.Printf("Project ID:      %d\n", projectID)
	fmt.Println("────────────────────────────────────────────────────────────")

	confirmed, err := interactive.AskConfirm("Подтвердить создание?")
	if err != nil || !confirmed {
		fmt.Println("\n❌ Отменено")
		return nil
	}

	req := &data.AddRunRequest{
		Name:        answers.Name,
		Description: answers.Description,
		SuiteID:     answers.SuiteID,
		IncludeAll:  answers.IncludeAll,
	}

	run, err := cli.AddRun(projectID, req)
	if err != nil {
		return fmt.Errorf("ошибка создания рана: %w", err)
	}

	fmt.Printf("\n✅ Run создан (ID: %d)\n", run.ID)
	return output.OutputResult(cmd, run, "result")
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
			return fmt.Errorf("необходимо указать project_id: gotr add suite <project_id> --dry-run")
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
			return fmt.Errorf("необходимо указать project_id: gotr add section <project_id> --dry-run")
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
			return fmt.Errorf("необходимо указать section_id: gotr add case <section_id> --dry-run")
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
			return fmt.Errorf("необходимо указать project_id: gotr add run <project_id> --dry-run")
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
			return fmt.Errorf("необходимо указать test_id: gotr add result <test_id> --dry-run")
		}
		if len(jsonData) > 0 {
			var req data.AddResultRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			body = data.AddResultRequest{
				StatusID: statusID,
				Comment:  comment,
				Elapsed:  elapsed,
				Defects:  defects,
				AssignedTo: assignedTo,
			}
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/add_result/%d", id)
		dr.PrintOperation(fmt.Sprintf("Add Result for Test %d", id), method, url, body)

	case "shared-step":
		if id == 0 {
			return fmt.Errorf("необходимо указать project_id: gotr add shared-step <project_id> --dry-run")
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
		return fmt.Errorf("используйте --dry-run с конкретной подкомандой attachment")

	default:
		return fmt.Errorf("неподдерживаемый endpoint для dry-run: %s", endpoint)
	}

	return nil
}

func addProject(cli client.ClientInterface, cmd *cobra.Command, jsonData []byte) error {
	var req data.AddProjectRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("необходимо указать --name")
		}
		req.Name = name
		req.Announcement, _ = cmd.Flags().GetString("announcement")
		req.ShowAnnouncement, _ = cmd.Flags().GetBool("show-announcement")
	}

	project, err := cli.AddProject(&req)
	if err != nil {
		return fmt.Errorf("ошибка создания проекта: %w", err)
	}

	return output.OutputResult(cmd, project, "result")
}

func addSuite(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	var req data.AddSuiteRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("необходимо указать --name")
		}
		req.Name = name
		req.Description, _ = cmd.Flags().GetString("description")
	}

	suite, err := cli.AddSuite(projectID, &req)
	if err != nil {
		return fmt.Errorf("ошибка создания сьюта: %w", err)
	}

	return output.OutputResult(cmd, suite, "result")
}

func addSection(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	var req data.AddSectionRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("необходимо указать --name")
		}
		req.Name = name
		req.Description, _ = cmd.Flags().GetString("description")
		req.SuiteID, _ = cmd.Flags().GetInt64("suite-id")
		req.ParentID, _ = cmd.Flags().GetInt64("section-id")
	}

	section, err := cli.AddSection(projectID, &req)
	if err != nil {
		return fmt.Errorf("ошибка создания секции: %w", err)
	}

	return output.OutputResult(cmd, section, "result")
}

func addCase(cli client.ClientInterface, cmd *cobra.Command, sectionID int64, jsonData []byte) error {
	var req data.AddCaseRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %w", err)
		}
	} else {
		title, _ := cmd.Flags().GetString("title")
		if title == "" {
			return fmt.Errorf("необходимо указать --title")
		}
		req.Title = title
		req.TemplateID, _ = cmd.Flags().GetInt64("template-id")
		req.TypeID, _ = cmd.Flags().GetInt64("type-id")
		req.PriorityID, _ = cmd.Flags().GetInt64("priority-id")
		req.Refs, _ = cmd.Flags().GetString("refs")
	}

	caseResp, err := cli.AddCase(sectionID, &req)
	if err != nil {
		return fmt.Errorf("ошибка создания кейса: %w", err)
	}

	return output.OutputResult(cmd, caseResp, "result")
}

func addRun(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	var req data.AddRunRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %w", err)
		}
	} else {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("необходимо указать --name")
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

	run, err := cli.AddRun(projectID, &req)
	if err != nil {
		return fmt.Errorf("ошибка создания рана: %w", err)
	}

	return output.OutputResult(cmd, run, "result")
}

func addResult(cli client.ClientInterface, cmd *cobra.Command, testID int64, jsonData []byte) error {
	var req data.AddResultRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %w", err)
		}
	} else {
		statusID, _ := cmd.Flags().GetInt64("status-id")
		if statusID == 0 {
			return fmt.Errorf("необходимо указать --status-id")
		}
		req.StatusID = statusID
		req.Comment, _ = cmd.Flags().GetString("comment")
		req.Elapsed, _ = cmd.Flags().GetString("elapsed")
		req.Defects, _ = cmd.Flags().GetString("defects")
		req.AssignedTo, _ = cmd.Flags().GetInt64("assignedto-id")
		req.Version, _ = cmd.Flags().GetString("version")
	}

	result, err := cli.AddResult(testID, &req)
	if err != nil {
		return fmt.Errorf("ошибка добавления результата: %w", err)
	}

	return output.OutputResult(cmd, result, "result")
}

func addResultForCase(cli client.ClientInterface, cmd *cobra.Command, runID, caseID int64, jsonData []byte) error {
	var req data.AddResultRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %w", err)
		}
	} else {
		statusID, _ := cmd.Flags().GetInt64("status-id")
		if statusID == 0 {
			return fmt.Errorf("необходимо указать --status-id")
		}
		req.StatusID = statusID
		req.Comment, _ = cmd.Flags().GetString("comment")
		req.Elapsed, _ = cmd.Flags().GetString("elapsed")
		req.Defects, _ = cmd.Flags().GetString("defects")
		req.AssignedTo, _ = cmd.Flags().GetInt64("assignedto-id")
	}

	result, err := cli.AddResultForCase(runID, caseID, &req)
	if err != nil {
		return fmt.Errorf("ошибка добавления результата: %w", err)
	}

	return output.OutputResult(cmd, result, "result")
}

func addSharedStep(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	var req data.AddSharedStepRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %w", err)
		}
	} else {
		title, _ := cmd.Flags().GetString("title")
		if title == "" {
			return fmt.Errorf("необходимо указать --title")
		}
		req.Title = title
	}

	step, err := cli.AddSharedStep(projectID, &req)
	if err != nil {
		return fmt.Errorf("ошибка создания shared step: %w", err)
	}

	return output.OutputResult(cmd, step, "result")
}


func parseCaseIDs(s string) []int64 {
	var ids []int64
	for _, part := range splitAndTrim(s, ",") {
		id, err := strconv.ParseInt(part, 10, 64)
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
		return fmt.Errorf("необходимо указать тип вложения: case, plan, plan-entry, result, run")
	}

	attachmentType := args[1]

	switch attachmentType {
	case "case":
		if len(args) < 4 {
			return fmt.Errorf("использование: gotr add attachment case <case_id> <file_path>")
		}
		caseID, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("неверный case_id: %w", err)
		}
		filePath := args[3]
		return addAttachmentToCase(cli, cmd, caseID, filePath)

	case "plan":
		if len(args) < 4 {
			return fmt.Errorf("использование: gotr add attachment plan <plan_id> <file_path>")
		}
		planID, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("неверный plan_id: %w", err)
		}
		filePath := args[3]
		return addAttachmentToPlan(cli, cmd, planID, filePath)

	case "plan-entry":
		if len(args) < 5 {
			return fmt.Errorf("использование: gotr add attachment plan-entry <plan_id> <entry_id> <file_path>")
		}
		planID, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("неверный plan_id: %w", err)
		}
		entryID := args[3]
		filePath := args[4]
		return addAttachmentToPlanEntry(cli, cmd, planID, entryID, filePath)

	case "result":
		if len(args) < 4 {
			return fmt.Errorf("использование: gotr add attachment result <result_id> <file_path>")
		}
		resultID, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("неверный result_id: %w", err)
		}
		filePath := args[3]
		return addAttachmentToResult(cli, cmd, resultID, filePath)

	case "run":
		if len(args) < 4 {
			return fmt.Errorf("использование: gotr add attachment run <run_id> <file_path>")
		}
		runID, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("неверный run_id: %w", err)
		}
		filePath := args[3]
		return addAttachmentToRun(cli, cmd, runID, filePath)

	default:
		return fmt.Errorf("неподдерживаемый тип вложения: %s. Доступные: case, plan, plan-entry, result, run", attachmentType)
	}
}

func addAttachmentToCase(cli client.ClientInterface, cmd *cobra.Command, caseID int64, filePath string) error {
	// Проверяем dry-run режим
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment case")
		dr.PrintSimple("Add Attachment to Case", fmt.Sprintf("Case ID: %d, File: %s", caseID, filePath))
		return nil
	}

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("файл не найден: %s", filePath)
	}

	resp, err := cli.AddAttachmentToCase(caseID, filePath)
	if err != nil {
		return fmt.Errorf("ошибка добавления вложения к кейсу: %w", err)
	}

	fmt.Printf("✅ Вложение добавлено (ID: %d)\n", resp.AttachmentID)
	fmt.Printf("   URL: %s\n", resp.URL)
	return output.OutputResult(cmd, resp, "result")
}

func addAttachmentToPlan(cli client.ClientInterface, cmd *cobra.Command, planID int64, filePath string) error {
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment plan")
		dr.PrintSimple("Add Attachment to Plan", fmt.Sprintf("Plan ID: %d, File: %s", planID, filePath))
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("файл не найден: %s", filePath)
	}

	resp, err := cli.AddAttachmentToPlan(planID, filePath)
	if err != nil {
		return fmt.Errorf("ошибка добавления вложения к плану: %w", err)
	}

	fmt.Printf("✅ Вложение добавлено (ID: %d)\n", resp.AttachmentID)
	fmt.Printf("   URL: %s\n", resp.URL)
	return output.OutputResult(cmd, resp, "result")
}

func addAttachmentToPlanEntry(cli client.ClientInterface, cmd *cobra.Command, planID int64, entryID, filePath string) error {
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment plan-entry")
		dr.PrintSimple("Add Attachment to Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s, File: %s", planID, entryID, filePath))
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("файл не найден: %s", filePath)
	}

	resp, err := cli.AddAttachmentToPlanEntry(planID, entryID, filePath)
	if err != nil {
		return fmt.Errorf("ошибка добавления вложения к plan entry: %w", err)
	}

	fmt.Printf("✅ Вложение добавлено (ID: %d)\n", resp.AttachmentID)
	fmt.Printf("   URL: %s\n", resp.URL)
	return output.OutputResult(cmd, resp, "result")
}

func addAttachmentToResult(cli client.ClientInterface, cmd *cobra.Command, resultID int64, filePath string) error {
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment result")
		dr.PrintSimple("Add Attachment to Result", fmt.Sprintf("Result ID: %d, File: %s", resultID, filePath))
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("файл не найден: %s", filePath)
	}

	resp, err := cli.AddAttachmentToResult(resultID, filePath)
	if err != nil {
		return fmt.Errorf("ошибка добавления вложения к результату: %w", err)
	}

	fmt.Printf("✅ Вложение добавлено (ID: %d)\n", resp.AttachmentID)
	fmt.Printf("   URL: %s\n", resp.URL)
	return output.OutputResult(cmd, resp, "result")
}

func addAttachmentToRun(cli client.ClientInterface, cmd *cobra.Command, runID int64, filePath string) error {
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := output.NewDryRunPrinter("add attachment run")
		dr.PrintSimple("Add Attachment to Run", fmt.Sprintf("Run ID: %d, File: %s", runID, filePath))
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("файл не найден: %s", filePath)
	}

	resp, err := cli.AddAttachmentToRun(runID, filePath)
	if err != nil {
		return fmt.Errorf("ошибка добавления вложения к рану: %w", err)
	}

	fmt.Printf("✅ Вложение добавлено (ID: %d)\n", resp.AttachmentID)
	fmt.Printf("   URL: %s\n", resp.URL)
	return output.OutputResult(cmd, resp, "result")
}
