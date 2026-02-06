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
  gotr update shared-step 50 --title "Updated Step"`,
	RunE: runUpdate,
}

func init() {
	// Общие флаги для обновления
	updateCmd.Flags().StringP("name", "n", "", "Название ресурса")
	updateCmd.Flags().StringP("description", "d", "", "Описание")
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
