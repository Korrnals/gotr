package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

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

Примеры:
  gotr add project --name "New Project" --announcement "Desc"
  gotr add suite 1 --name "Smoke Tests"
  gotr add case 100 --title "Login test" --template-id 1
  gotr add run 1 --name "Nightly Run" --suite-id 100
  gotr add result 12345 --status-id 1 --comment "Passed"`,
	RunE: runAdd,
}

func init() {
	// Общие флаги для создания
	addCmd.Flags().StringP("name", "n", "", "Название ресурса")
	addCmd.Flags().StringP("description", "d", "", "Описание/announcement")
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
	addCmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("необходимо указать endpoint: project, suite, section, case, run, result, result-for-case, shared-step, milestone, plan, entry")
	}

	endpoint := args[0]
	
	// Получаем клиент
	cli := GetClientInterface(cmd)

	// Определяем ID из аргументов
	var id int64
	if len(args) > 1 {
		parsedID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("неверный ID: %v", err)
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
			return fmt.Errorf("ошибка чтения JSON-файла: %v", err)
		}
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
	default:
		return fmt.Errorf("неподдерживаемый endpoint: %s", endpoint)
	}
}

func addProject(cli client.ClientInterface, cmd *cobra.Command, jsonData []byte) error {
	var req data.AddProjectRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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
		return fmt.Errorf("ошибка создания проекта: %v", err)
	}

	return outputResult(cmd, project)
}

func addSuite(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	var req data.AddSuiteRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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
		return fmt.Errorf("ошибка создания сьюта: %v", err)
	}

	return outputResult(cmd, suite)
}

func addSection(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	var req data.AddSectionRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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
		return fmt.Errorf("ошибка создания секции: %v", err)
	}

	return outputResult(cmd, section)
}

func addCase(cli client.ClientInterface, cmd *cobra.Command, sectionID int64, jsonData []byte) error {
	var req data.AddCaseRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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
		return fmt.Errorf("ошибка создания кейса: %v", err)
	}

	return outputResult(cmd, caseResp)
}

func addRun(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	var req data.AddRunRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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
		return fmt.Errorf("ошибка создания рана: %v", err)
	}

	return outputResult(cmd, run)
}

func addResult(cli client.ClientInterface, cmd *cobra.Command, testID int64, jsonData []byte) error {
	var req data.AddResultRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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
		return fmt.Errorf("ошибка добавления результата: %v", err)
	}

	return outputResult(cmd, result)
}

func addResultForCase(cli client.ClientInterface, cmd *cobra.Command, runID, caseID int64, jsonData []byte) error {
	var req data.AddResultRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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
		return fmt.Errorf("ошибка добавления результата: %v", err)
	}

	return outputResult(cmd, result)
}

func addSharedStep(cli client.ClientInterface, cmd *cobra.Command, projectID int64, jsonData []byte) error {
	var req data.AddSharedStepRequest
	
	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %v", err)
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
		return fmt.Errorf("ошибка создания shared step: %v", err)
	}

	return outputResult(cmd, step)
}

func outputResult(cmd *cobra.Command, data interface{}) error {
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
