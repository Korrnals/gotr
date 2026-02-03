package get

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/internal/client"
)

// selectProjectInteractively показывает список проектов и просит выбрать
func selectProjectInteractively(client *client.HTTPClient) (int64, error) {
	projects, err := client.GetProjects()
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
