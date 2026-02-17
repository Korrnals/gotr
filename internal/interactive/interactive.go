// internal/interactive/interactive.go
// Пакет для интерактивного взаимодействия с пользователем
package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
)

// SelectProjectInteractively показывает список проектов и просит выбрать
func SelectProjectInteractively(httpClient client.ClientInterface) (int64, error) {
	projects, err := httpClient.GetProjects()
	if err != nil {
		return 0, fmt.Errorf("не удалось получить список проектов: %w", err)
	}

	if len(projects) == 0 {
		return 0, fmt.Errorf("не найдено проектов")
	}

	fmt.Println("\nДоступные проекты:")
	fmt.Println(strings.Repeat("-", 70))

	for i, p := range projects {
		fmt.Printf("  [%d] ID: %d | %s\n", i+1, p.ID, p.Name)
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Выберите номер проекта (1-%d): ", len(projects))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ввода: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(projects) {
		return 0, fmt.Errorf("неверный выбор: %s (ожидается число от 1 до %d)", input, len(projects))
	}

	selected := projects[choice-1]
	fmt.Printf("\nВыбран проект: %s (ID: %d)\n\n", selected.Name, selected.ID)

	return selected.ID, nil
}

// SelectSuiteInteractively показывает список сьютов и просит выбрать
func SelectSuiteInteractively(suites data.GetSuitesResponse) (int64, error) {
	fmt.Println("\nВ проекте найдено несколько сьютов:")
	fmt.Println(strings.Repeat("-", 60))

	for i, suite := range suites {
		fmt.Printf("  [%d] ID: %d | %s\n", i+1, suite.ID, suite.Name)
		if suite.Description != "" {
			desc := suite.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			fmt.Printf("       %s\n", desc)
		}
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Выберите номер сьюта (1-%d): ", len(suites))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ввода: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(suites) {
		return 0, fmt.Errorf("неверный выбор: %s (ожидается число от 1 до %d)", input, len(suites))
	}

	selectedSuite := suites[choice-1]
	fmt.Printf("\nВыбран сьют: %s (ID: %d)\n\n", selectedSuite.Name, selectedSuite.ID)

	return selectedSuite.ID, nil
}

// SelectRunInteractively показывает список runs и просит выбрать
func SelectRunInteractively(runs data.GetRunsResponse) (int64, error) {
	fmt.Println("\nДоступные test runs:")
	fmt.Println(strings.Repeat("-", 70))

	for i, run := range runs {
		status := "🟢"
		if run.IsCompleted {
			status = "🔴"
		}
		fmt.Printf("  [%d] %s ID: %d | %s\n", i+1, status, run.ID, run.Name)
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Выберите номер run (1-%d): ", len(runs))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ввода: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(runs) {
		return 0, fmt.Errorf("неверный выбор: %s (ожидается число от 1 до %d)", input, len(runs))
	}

	selected := runs[choice-1]
	fmt.Printf("\nВыбран run: %s (ID: %d)\n\n", selected.Name, selected.ID)

	return selected.ID, nil
}

// ConfirmAction запрашивает подтверждение действия у пользователя
func ConfirmAction(message string) bool {
	fmt.Printf("%s [y/N]: ", message)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}
