package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/cmd/common/wizard"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// updateCmd — команда для обновления ресурсов через POST-запросы
var updateCmd = &cobra.Command{
	Use:   "update <endpoint> <id>",
	Short: "Обновить существующий ресурс (POST-запрос)",
	Long: `Обновляет существующий объект в TestRail через POST API.

Поддерживаемые эндпоинты:
  project <id>       Обновить проект
  suite <id>         Обновить сьют
  section <id>       Обновить секцию
  case <id>          Обновить тест-кейс
  run <id>           Обновить тест-ран
  shared-step <id>   Обновить shared step
  milestone <id>     Обновить milestone
  plan <id>          Обновить test plan

Примеры:
  gotr update project 1 --name "Updated Project"
  gotr update suite 100 --name "Updated Suite"
  gotr update case 12345 --title "Updated Title" --priority-id 2
  gotr update run 1000 --name "Updated Run Name"
  gotr update shared-step 50 --title "Updated Step"

Интерактивный режим (wizard):
  gotr update project 1 -i
  gotr update suite 100 -i
  gotr update case 12345 -i

Dry-run mode:
  gotr update project 1 --name "Test" --dry-run  # Show what would be updated`,
	RunE: runUpdate,
}

func init() {
	// Общие флаги для обновления
	updateCmd.Flags().StringP("name", "n", "", "Название ресурса")
	updateCmd.Flags().String("description", "", "Описание")
	updateCmd.Flags().String("announcement", "", "Announcement (для проекта)")
	updateCmd.Flags().Bool("show-announcement", false, "Показывать announcement")
	updateCmd.Flags().Bool("is-completed", false, "Отметить как завершённый")
	updateCmd.Flags().String("title", "", "Заголовок (для case)")
	updateCmd.Flags().Int64("type-id", 0, "ID типа (для case)")
	updateCmd.Flags().Int64("priority-id", 0, "ID приоритета (для case)")
	updateCmd.Flags().String("refs", "", "Ссылки (references)")
	updateCmd.Flags().Int64("suite-id", 0, "ID сьюта")
	updateCmd.Flags().Int64("milestone-id", 0, "ID milestone")
	updateCmd.Flags().Int64("assignedto-id", 0, "ID назначенного пользователя")
	updateCmd.Flags().String("case-ids", "", "ID кейсов через запятую (для run)")
	updateCmd.Flags().Bool("include-all", false, "Включить все кейсы (для run)")
	updateCmd.Flags().String("json-file", "", "Путь к JSON-файлу с данными")
	updateCmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл")
	updateCmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")
	updateCmd.Flags().BoolP("interactive", "i", false, "Интерактивный режим (wizard)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("необходимо указать endpoint и id: gotr update <endpoint> <id>")
	}

	endpoint := args[0]
	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("неверный ID: %v", err)
	}

	// Получаем клиент
	cli := GetClientInterface(cmd)

	// Читаем JSON из файла если указан
	jsonFile, _ := cmd.Flags().GetString("json-file")
	var jsonData []byte
	if jsonFile != "" {
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("ошибка чтения JSON-файла: %v", err)
		}
	}

	// Проверяем dry-run режим
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := dryrun.New("update " + endpoint)
		return runUpdateDryRun(cmd, dr, endpoint, id, jsonData)
	}

	// Проверяем интерактивный режим
	isInteractive, _ := cmd.Flags().GetBool("interactive")
	if isInteractive {
		return runUpdateInteractive(cli, cmd, endpoint, id)
	}

	// Маршрутизация по endpoint
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
	default:
		return fmt.Errorf("неподдерживаемый endpoint: %s", endpoint)
	}
}

// runUpdateInteractive запускает интерактивный wizard для обновления ресурса
func runUpdateInteractive(cli client.ClientInterface, cmd *cobra.Command, endpoint string, id int64) error {
	switch endpoint {
	case "project":
		return updateProjectInteractive(cli, cmd, id)
	case "suite":
		return updateSuiteInteractive(cli, cmd, id)
	case "case":
		return updateCaseInteractive(cli, cmd, id)
	default:
		return fmt.Errorf("интерактивный режим не поддерживается для endpoint: %s", endpoint)
	}
}

func updateProjectInteractive(cli client.ClientInterface, cmd *cobra.Command, id int64) error {
	answers, err := wizard.AskProject(true)
	if err != nil {
		return fmt.Errorf("ошибка ввода: %v", err)
	}

	// Предпросмотр
	fmt.Println("\n────────────────────────────────────────────────────────────")
	fmt.Println("📋 ПРЕДПРОСМОТР: Update Project")
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("Project ID:      %d\n", id)
	fmt.Printf("Название:        %s\n", answers.Name)
	fmt.Printf("Announcement:    %s\n", answers.Announcement)
	fmt.Printf("Show announce:   %v\n", answers.ShowAnnouncement)
	fmt.Printf("Is completed:    %v\n", answers.IsCompleted)
	fmt.Println("────────────────────────────────────────────────────────────")

	confirmed, err := wizard.AskConfirm("Подтвердить обновление?")
	if err != nil || !confirmed {
		fmt.Println("\n❌ Отменено")
		return nil
	}

	req := &data.UpdateProjectRequest{
		Name:             answers.Name,
		Announcement:     answers.Announcement,
		ShowAnnouncement: answers.ShowAnnouncement,
		IsCompleted:      answers.IsCompleted,
	}

	project, err := cli.UpdateProject(id, req)
	if err != nil {
		return fmt.Errorf("ошибка обновления проекта: %v", err)
	}

	fmt.Printf("\n✅ Проект обновлён (ID: %d)\n", project.ID)
	return outputUpdateResult(cmd, project)
}

func updateSuiteInteractive(cli client.ClientInterface, cmd *cobra.Command, id int64) error {
	answers, err := wizard.AskSuite(true)
	if err != nil {
		return fmt.Errorf("ошибка ввода: %v", err)
	}

	// Предпросмотр
	fmt.Println("\n────────────────────────────────────────────────────────────")
	fmt.Println("📋 ПРЕДПРОСМОТР: Update Suite")
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("Suite ID:        %d\n", id)
	fmt.Printf("Название:        %s\n", answers.Name)
	fmt.Printf("Описание:        %s\n", answers.Description)
	fmt.Printf("Is completed:    %v\n", answers.IsCompleted)
	fmt.Println("────────────────────────────────────────────────────────────")

	confirmed, err := wizard.AskConfirm("Подтвердить обновление?")
	if err != nil || !confirmed {
		fmt.Println("\n❌ Отменено")
		return nil
	}

	req := &data.UpdateSuiteRequest{
		Name:        answers.Name,
		Description: answers.Description,
		IsCompleted: answers.IsCompleted,
	}

	suite, err := cli.UpdateSuite(id, req)
	if err != nil {
		return fmt.Errorf("ошибка обновления сьюта: %v", err)
	}

	fmt.Printf("\n✅ Сьют обновлён (ID: %d)\n", suite.ID)
	return outputUpdateResult(cmd, suite)
}

func updateCaseInteractive(cli client.ClientInterface, cmd *cobra.Command, id int64) error {
	answers, err := wizard.AskCase(true)
	if err != nil {
		return fmt.Errorf("ошибка ввода: %v", err)
	}

	// Предпросмотр
	fmt.Println("\n────────────────────────────────────────────────────────────")
	fmt.Println("📋 ПРЕДПРОСМОТР: Update Case")
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("Case ID:         %d\n", id)
	fmt.Printf("Заголовок:       %s\n", answers.Title)
	fmt.Printf("Type ID:         %d\n", answers.TypeID)
	fmt.Printf("Priority ID:     %d\n", answers.PriorityID)
	fmt.Println("────────────────────────────────────────────────────────────")

	confirmed, err := wizard.AskConfirm("Подтвердить обновление?")
	if err != nil || !confirmed {
		fmt.Println("\n❌ Отменено")
		return nil
	}

	req := &data.UpdateCaseRequest{
		Title:      &answers.Title,
		TypeID:     &answers.TypeID,
		PriorityID: &answers.PriorityID,
		Refs:       &answers.Refs,
	}

	caseResp, err := cli.UpdateCase(id, req)
	if err != nil {
		return fmt.Errorf("ошибка обновления кейса: %v", err)
	}

	fmt.Printf("\n✅ Кейс обновлён (ID: %d)\n", caseResp.ID)
	return outputUpdateResult(cmd, caseResp)
}

// runUpdateDryRun выполняет dry-run для update команды
func runUpdateDryRun(cmd *cobra.Command, dr *dryrun.Printer, endpoint string, id int64, jsonData []byte) error {
	// Читаем флаги
	name, _ := cmd.Flags().GetString("name")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	announcement, _ := cmd.Flags().GetString("announcement")
	showAnn, _ := cmd.Flags().GetBool("show-announcement")
	isCompleted, _ := cmd.Flags().GetBool("is-completed")
	milestoneID, _ := cmd.Flags().GetInt64("milestone-id")
	assignedToID, _ := cmd.Flags().GetInt64("assignedto-id")
	includeAll, _ := cmd.Flags().GetBool("include-all")
	typeID, _ := cmd.Flags().GetInt64("type-id")
	priorityID, _ := cmd.Flags().GetInt64("priority-id")
	refs, _ := cmd.Flags().GetString("refs")
	caseIDsStr, _ := cmd.Flags().GetString("case-ids")

	var method, url string
	var body interface{}

	switch endpoint {
	case "project":
		if len(jsonData) > 0 {
			var req data.UpdateProjectRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			req := data.UpdateProjectRequest{
				ShowAnnouncement: showAnn,
				IsCompleted:      isCompleted,
			}
			if name != "" {
				req.Name = name
			}
			if announcement != "" {
				req.Announcement = announcement
			}
			body = req
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/update_project/%d", id)
		dr.PrintOperation(fmt.Sprintf("Update Project %d", id), method, url, body)

	case "suite":
		if len(jsonData) > 0 {
			var req data.UpdateSuiteRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			req := data.UpdateSuiteRequest{
				IsCompleted: isCompleted,
			}
			if name != "" {
				req.Name = name
			}
			if description != "" {
				req.Description = description
			}
			body = req
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/update_suite/%d", id)
		dr.PrintOperation(fmt.Sprintf("Update Suite %d", id), method, url, body)

	case "section":
		if len(jsonData) > 0 {
			var req data.UpdateSectionRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			req := data.UpdateSectionRequest{}
			if name != "" {
				req.Name = name
			}
			if description != "" {
				req.Description = description
			}
			body = req
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/update_section/%d", id)
		dr.PrintOperation(fmt.Sprintf("Update Section %d", id), method, url, body)

	case "case":
		if len(jsonData) > 0 {
			var req data.UpdateCaseRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
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
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/update_case/%d", id)
		dr.PrintOperation(fmt.Sprintf("Update Case %d", id), method, url, body)

	case "run":
		if len(jsonData) > 0 {
			var req data.UpdateRunRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			req := data.UpdateRunRequest{
				IncludeAll: &includeAll,
			}
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
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/update_run/%d", id)
		dr.PrintOperation(fmt.Sprintf("Update Run %d", id), method, url, body)

	case "shared-step":
		if len(jsonData) > 0 {
			var req data.UpdateSharedStepRequest
			json.Unmarshal(jsonData, &req)
			body = req
		} else {
			req := data.UpdateSharedStepRequest{}
			if title != "" {
				req.Title = title
			}
			body = req
		}
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/update_shared_step/%d", id)
		dr.PrintOperation(fmt.Sprintf("Update Shared Step %d", id), method, url, body)

	default:
		return fmt.Errorf("неподдерживаемый endpoint для dry-run: %s", endpoint)
	}

	return nil
}

func updateProject(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	var req data.UpdateProjectRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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

	project, err := cli.UpdateProject(id, &req)
	if err != nil {
		return fmt.Errorf("ошибка обновления проекта: %v", err)
	}

	return outputUpdateResult(cmd, project)
}

func updateSuite(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	var req data.UpdateSuiteRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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

	suite, err := cli.UpdateSuite(id, &req)
	if err != nil {
		return fmt.Errorf("ошибка обновления сьюта: %v", err)
	}

	return outputUpdateResult(cmd, suite)
}

func updateSection(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	var req data.UpdateSectionRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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

	section, err := cli.UpdateSection(id, &req)
	if err != nil {
		return fmt.Errorf("ошибка обновления секции: %v", err)
	}

	return outputUpdateResult(cmd, section)
}

func updateCase(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	var req data.UpdateCaseRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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

	caseResp, err := cli.UpdateCase(id, &req)
	if err != nil {
		return fmt.Errorf("ошибка обновления кейса: %v", err)
	}

	return outputUpdateResult(cmd, caseResp)
}

func updateRun(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	var req data.UpdateRunRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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

	run, err := cli.UpdateRun(id, &req)
	if err != nil {
		return fmt.Errorf("ошибка обновления рана: %v", err)
	}

	return outputUpdateResult(cmd, run)
}

func updateSharedStep(cli client.ClientInterface, cmd *cobra.Command, id int64, jsonData []byte) error {
	var req data.UpdateSharedStepRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
		}
	} else {
		title, _ := cmd.Flags().GetString("title")
		if title != "" {
			req.Title = title
		}
	}

	step, err := cli.UpdateSharedStep(id, &req)
	if err != nil {
		return fmt.Errorf("ошибка обновления shared step: %v", err)
	}

	return outputUpdateResult(cmd, step)
}

func outputUpdateResult(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	
	if output != "" {
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(output, jsonBytes, 0644)
	}
	
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonBytes))
	return nil
}
