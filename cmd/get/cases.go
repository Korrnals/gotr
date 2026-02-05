package get

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// casesCmd — подкоманда для списка кейсов проекта
var casesCmd = &cobra.Command{
	Use:   "cases [project-id]",
	Short: "Получить кейсы проекта",
	Long: `Получить кейсы проекта.

Если проект содержит несколько сьютов и --suite-id не указан, 
будет предложено выбрать сьют из списка.

Обязательные параметры:
	ID проекта — позиционный аргумент или флаг --project-id

Опциональные параметры:
	ID сюиты — флаг --suite-id (обязателен для проектов с multiple suites)
	ID секции — флаг --section-id
	Все сьюты — флаг --all-suites (получить кейсы из всех сьютов проекта)

Примеры:
	# Автоматический выбор сьюта (если один) или интерактивный выбор
	gotr get cases 30

	# Явное указание сьюта
	gotr get cases 30 --suite-id 20069

	# Получить кейсы из всех сьютов проекта
	gotr get cases 30 --all-suites

	# С фильтрацией по секции
	gotr get cases 30 --suite-id 20069 --section-id 100
`,
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		projectIDStr := ""
		if len(args) > 0 {
			projectIDStr = args[0]
		}
		if pid, _ := command.Flags().GetString("project-id"); pid != "" {
			projectIDStr = pid
		}
		var projectID int64
		var err error

		if projectIDStr == "" {
			// Интерактивный выбор проекта
			projectID, err = selectProjectInteractively(client)
			if err != nil {
				return err
			}
		} else {
			projectID, err = strconv.ParseInt(projectIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("некорректный ID проекта: %w", err)
			}
		}

		sectionID, _ := command.Flags().GetInt64("section-id")
		allSuites, _ := command.Flags().GetBool("all-suites")
		suiteID, _ := command.Flags().GetInt64("suite-id")

		// Если указан конкретный suite-id — используем его
		if suiteID != 0 {
			return fetchAndOutputCases(command, client, projectID, suiteID, sectionID, start)
		}

		// Получаем список сьютов проекта
		suites, err := client.GetSuites(projectID)
		if err != nil {
			return fmt.Errorf("не удалось получить список сьютов проекта %d: %w", projectID, err)
		}

		if len(suites) == 0 {
			return fmt.Errorf("в проекте %d не найдено сьютов", projectID)
		}

		// Если --all-suites — собираем кейсы из всех сьютов
		if allSuites {
			return fetchCasesFromAllSuites(command, client, projectID, suites, sectionID, start)
		}

		// Если только один сьют — используем его автоматически
		if len(suites) == 1 {
			fmt.Printf("В проекте найден один сьют (ID: %d), используем его автоматически...\n", suites[0].ID)
			return fetchAndOutputCases(command, client, projectID, suites[0].ID, sectionID, start)
		}

		// Несколько сьютов — интерактивный выбор
		selectedSuiteID, err := selectSuiteInteractively(suites)
		if err != nil {
			return err
		}

		return fetchAndOutputCases(command, client, projectID, selectedSuiteID, sectionID, start)
	},
}

// fetchAndOutputCases получает кейсы и выводит результат
func fetchAndOutputCases(cmd *cobra.Command, client *client.HTTPClient, projectID, suiteID, sectionID int64, start time.Time) error {
	cases, err := client.GetCases(projectID, suiteID, sectionID)
	if err != nil {
		return err
	}

	return handleOutput(cmd, cases, start)
}

// fetchCasesFromAllSuites получает кейсы из всех сьютов проекта
func fetchCasesFromAllSuites(cmd *cobra.Command, client *client.HTTPClient, projectID int64, suites data.GetSuitesResponse, sectionID int64, start time.Time) error {
	fmt.Printf("Получение кейсов из %d сьютов проекта...\n\n", len(suites))

	allCases := make(data.GetCasesResponse, 0)
	for _, suite := range suites {
		fmt.Printf("Сьют: %s (ID: %d)... ", suite.Name, suite.ID)
		cases, err := client.GetCases(projectID, suite.ID, sectionID)
		if err != nil {
			fmt.Printf("ОШИБКА: %v\n", err)
			continue
		}
		fmt.Printf("найдено %d кейсов\n", len(cases))
		allCases = append(allCases, cases...)
	}

	fmt.Printf("\n=== Итого: %d кейсов из %d сьютов ===\n\n", len(allCases), len(suites))
	return handleOutput(cmd, allCases, start)
}

// caseCmd — подкоманда для одного кейса
var caseCmd = &cobra.Command{
	Use:   "case <case-id>",
	Short: "Получить один кейс по ID кейса",
	Long:  "Получает детальную информацию о конкретном тест-кейсе по его ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID кейса: %w", err)
		}

		kase, err := client.GetCase(id)
		if err != nil {
			return err
		}

		return handleOutput(command, kase, start)
	},
}
