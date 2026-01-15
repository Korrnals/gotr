package cmd

import (
	"encoding/json"
	"fmt"
	embed "gotr/embedded"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getCmd — основная команда для GET-запросов
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "GET-запросы к TestRail API",
	Long: `Выполняет GET-запросы к TestRail API.

Подкоманды:
	case               - получить один кейс по ID кейса
	cases              - получить кейсы проекта (требует ID проекта и ID сюиты)
	case-types         - получить список типов кейсов
	case-fields        - получить список полей кейсов
	case-history       - получить историю изменений кейса по ID кейса

	project            - получить один проект по ID проекта
	projects           - получить все проекты

	sharedstep         - получить один shared step по ID шага
	sharedsteps        - получить shared steps проекта (требует ID проекта)
	sharedstep-history - получить историю изменений shared step по ID шага

	suite              - получить одну тест-сюиту по ID сюиты
	suites             - получить тест-сюиты проекта (требует ID проекта)

Примеры:
	gotr get project 30
	gotr get projects

	gotr get case 12345
	gotr get cases 30 --suite-id 20069

	gotr get suite 20069
	gotr get suites 30
	
	gotr get sharedstep 45678
	gotr get sharedsteps 30
`,
}

// handleOutput — общая логика вывода, сохранения и jq
func handleOutput(cmd *cobra.Command, data any, start time.Time) error {
	quiet, _ := cmd.Flags().GetBool("quiet")
	outputFormat, _ := cmd.Flags().GetString("type")
	saveFile, _ := cmd.Flags().GetString("output")
	jqEnabled, _ := cmd.Flags().GetBool("jq")
	jqFilter, _ := cmd.Flags().GetString("jq-filter")
	bodyOnly, _ := cmd.Flags().GetBool("body-only")

	// Если jq не указан явно — берём из конфига
	if !jqEnabled {
		jqEnabled = viper.GetBool("jq_format")
	}

	if saveFile != "" {
		var toSave []byte
		var err error
		if bodyOnly {
			toSave, err = json.MarshalIndent(data, "", "  ")
		} else {
			full := struct {
				Status     string        `json:"status"`
				StatusCode int           `json:"status_code"`
				Duration   time.Duration `json:"duration"`
				Timestamp  time.Time     `json:"timestamp"`
				Data       any           `json:"data"`
			}{
				Status:     "200 OK",
				StatusCode: 200,
				Duration:   time.Since(start),
				Timestamp:  time.Now(),
				Data:       data,
			}
			toSave, err = json.MarshalIndent(full, "", "  ")
		}
		if err != nil {
			return fmt.Errorf("ошибка маршалинга: %w", err)
		}
		if err := os.WriteFile(saveFile, toSave, 0644); err != nil {
			return err
		}
		if !quiet {
			fmt.Printf("Ответ сохранён в %s\n", saveFile)
		}
	}

	if jqEnabled || jqFilter != "" {
		toSave, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("ошибка маршалинга для jq: %w", err)
		}
		if err := embed.RunEmbeddedJQ(toSave, jqFilter); err != nil {
			return err
		}
		return nil
	}

	if !quiet {
		switch outputFormat {
		case "json":
			pretty, _ := json.MarshalIndent(data, "", "  ")
			fmt.Println(string(pretty))
		case "json-full":
			full := struct {
				Status     string        `json:"status"`
				StatusCode int           `json:"status_code"`
				Duration   time.Duration `json:"duration"`
				Timestamp  time.Time     `json:"timestamp"`
				Data       any           `json:"data"`
			}{
				Status:     "200 OK",
				StatusCode: 200,
				Duration:   time.Since(start),
				Timestamp:  time.Now(),
				Data:       data,
			}
			pretty, _ := json.MarshalIndent(full, "", "  ")
			fmt.Println(string(pretty))
		default:
			fmt.Println("Table output not implemented yet")
		}
	}

	return nil
}

// getCasesCmd — подкоманда для списка кейсов проекта
var getCasesCmd = &cobra.Command{
	Use:   "cases [project-id]",
	Short: "Получить кейсы проекта (требует ID проекта и ID сюиты)",
	Long: `Получить кейсы проекта.

Обязательные параметры:
	ID проекта — позиционный аргумент или флаг --project-id
	ID сюиты — флаг --suite-id (обязательно для проектов в режиме multiple suites)

Подсказки:
	ID проекта: используйте gotr get projects или TestRail UI (раздел Projects)
	ID сюиты: TestRail UI (раздел Suites проекта)

Примеры:
	gotr get cases 30 --suite-id 20069
	gotr get cases --project-id 30 --suite-id 20069 --section-id 100
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		projectIDStr := ""
		if len(args) > 0 {
			projectIDStr = args[0]
		}
		if pid, _ := cmd.Flags().GetString("project-id"); pid != "" {
			projectIDStr = pid
		}
		if projectIDStr == "" {
			return fmt.Errorf("укажите ID проекта через позиционный аргумент или флаг --project-id")
		}
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID проекта: %w", err)
		}

		suiteID, _ := cmd.Flags().GetInt64("suite-id")
		sectionID, _ := cmd.Flags().GetInt64("section-id")

		cases, err := client.GetCases(projectID, suiteID, sectionID)
		if err != nil {
			return err
		}

		return handleOutput(cmd, cases, start)
	},
}

// getCaseCmd — подкоманда для одного кейса
var getCaseCmd = &cobra.Command{
	Use:   "case <case-id>",
	Short: "Получить один кейс по ID кейса (ID кейса можно получить из команды gotr get cases <project-id> --suite-id <suite-id>)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID кейса: %w", err)
		}

		kase, err := client.GetCase(id)
		if err != nil {
			return err
		}

		return handleOutput(cmd, kase, start)
	},
}

// getCaseHistoryCmd — подкоманда для истории кейса
var getCaseHistoryCmd = &cobra.Command{
	Use:   "case-history <case-id>",
	Short: "Получить историю изменений кейса по ID кейса",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID кейса: %w", err)
		}

		history, err := client.GetHistoryForCase(id)
		if err != nil {
			return err
		}

		return handleOutput(cmd, history, start)
	},
}

// getCaseTypesCmd — подкоманда для типов кейсов
var getCaseTypesCmd = &cobra.Command{
	Use:   "case-types",
	Short: "Получить список типов кейсов",
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		types, err := client.GetCaseTypes()
		if err != nil {
			return err
		}

		return handleOutput(cmd, types, start)
	},
}

// getCaseFieldsCmd — подкоманда для полей кейсов
var getCaseFieldsCmd = &cobra.Command{
	Use:   "case-fields",
	Short: "Получить список полей кейсов",
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		fields, err := client.GetCaseFields()
		if err != nil {
			return err
		}

		return handleOutput(cmd, fields, start)
	},
}

// getProjectsCmd — подкоманда для всех проектов
var getProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Получить все проекты",
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		projects, err := client.GetProjects()
		if err != nil {
			return err
		}

		return handleOutput(cmd, projects, start)
	},
}

// getProjectCmd — подкоманда для одного проекта
var getProjectCmd = &cobra.Command{
	Use:   "project <project-id>",
	Short: "Получить один проект по ID проекта",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID проекта: %w", err)
		}

		project, err := client.GetProject(id)
		if err != nil {
			return err
		}

		return handleOutput(cmd, project, start)
	},
}

// getSharedStepsCmd — подкоманда для списка shared steps проекта
var getSharedStepsCmd = &cobra.Command{
	Use:   "sharedsteps",
	Short: "Получить shared steps проекта (требует ID проекта)",
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		projectIDStr := ""
		if len(args) > 0 {
			projectIDStr = args[0]
		}
		if pid, _ := cmd.Flags().GetString("project-id"); pid != "" {
			projectIDStr = pid
		}
		if projectIDStr == "" {
			return fmt.Errorf("укажите ID проекта через флаг --project-id или как позиционный аргумент")
		}
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID проекта: %w", err)
		}

		steps, err := client.GetSharedSteps(projectID)
		if err != nil {
			return err
		}

		return handleOutput(cmd, steps, start)
	},
}

// getSharedStepCmd — подкоманда для одного shared step
var getSharedStepCmd = &cobra.Command{
	Use:   "sharedstep <step-id>",
	Short: "Получить один shared step по ID шага",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID шага: %w", err)
		}

		step, err := client.GetSharedStep(id)
		if err != nil {
			return err
		}

		return handleOutput(cmd, step, start)
	},
}

// getSharedStepHistoryCmd — подкоманда для истории shared step
var getSharedStepHistoryCmd = &cobra.Command{
	Use:   "sharedstep-history <step-id>",
	Short: "Получить историю изменений shared step по ID шага",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID шага: %w", err)
		}

		history, err := client.GetSharedStepHistory(id)
		if err != nil {
			return err
		}

		return handleOutput(cmd, history, start)
	},
}

// getSuitesCmd — подкоманда для списка тест-сюит проекта
var getSuitesCmd = &cobra.Command{
	Use:   "suites",
	Short: "Получить тест-сюиты проекта (требует ID проекта)",
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		// Поддержка позиционного project-id
		projectIDStr := ""
		if len(args) > 0 {
			projectIDStr = args[0]
		}
		if pid, _ := cmd.Flags().GetString("project-id"); pid != "" {
			projectIDStr = pid // флаг имеет приоритет
		}
		if projectIDStr == "" {
			return fmt.Errorf("укажите ID проекта через флаг --project-id или как позиционный аргумент (пример: gotr get suites 30)")
		}
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID проекта: %w", err)
		}

		suites, err := client.GetSuites(projectID)
		if err != nil {
			return err
		}

		return handleOutput(cmd, suites, start)
	},
}

// getSuiteCmd — подкоманда для одной тест-сюиты по ID
var getSuiteCmd = &cobra.Command{
	Use:   "suite <suite-id>",
	Short: "Получить одну тест-сюиту по ID сюиты",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		client := GetClient(cmd)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID сюиты: %w", err)
		}

		suite, err := client.GetSuite(id)
		if err != nil {
			return err
		}

		return handleOutput(cmd, suite, start)
	},
}

func init() {
	// Добавляем подкоманды
	getCmd.AddCommand(getCasesCmd)
	getCmd.AddCommand(getCaseCmd)
	getCmd.AddCommand(getCaseTypesCmd)
	getCmd.AddCommand(getCaseFieldsCmd)
	getCmd.AddCommand(getCaseHistoryCmd)
	getCmd.AddCommand(getProjectsCmd)
	getCmd.AddCommand(getProjectCmd)
	getCmd.AddCommand(getSharedStepsCmd)
	getCmd.AddCommand(getSharedStepCmd)
	getCmd.AddCommand(getSharedStepHistoryCmd)
	getCmd.AddCommand(getSuitesCmd)
	getCmd.AddCommand(getSuiteCmd)

	// Локальные флаги — только для подкоманд get и их детей
	for _, subCmd := range getCmd.Commands() {
		subCmd.Flags().StringP("type", "t", "json", "Формат вывода: json, json-full, table")
		subCmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл")
		subCmd.Flags().BoolP("quiet", "q", false, "Тихий режим")
		subCmd.Flags().BoolP("jq", "j", false, "Включить jq-форматирование (переопределяет конфиг jq_format)")
		subCmd.Flags().String("jq-filter", "", "jq-фильтр")
		subCmd.Flags().BoolP("body-only", "b", false, "Сохранить только тело ответа (без метаданных)")
	}

	// Специфичные флаги для отдельных подкоманд
	getCasesCmd.Flags().Int64P("suite-id", "s", 0, "ID тест-сюиты (обязательно для проектов в режиме multiple suites)")
	getCasesCmd.Flags().Int64("section-id", 0, "ID секции (опционально)")
	getCasesCmd.MarkFlagRequired("suite-id")

	getSharedStepsCmd.MarkFlagRequired("project-id") // если хочешь оставить жёсткое требование
}
