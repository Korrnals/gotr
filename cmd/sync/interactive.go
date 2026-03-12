package sync

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/internal/client"
)

// selectProjectInteractively показывает список проектов и просит выбрать
func selectProjectInteractively(ctx context.Context, cli client.ClientInterface, prompt string) (int64, error) {
	projects, err := cli.GetProjects(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get projects list: %w", err)
	}

	if len(projects) == 0 {
		return 0, fmt.Errorf("no projects found")
	}

	fmt.Printf("\n%s\n", prompt)
	fmt.Println(strings.Repeat("-", 70))

	for i, p := range projects {
		fmt.Printf("  [%d] ID: %d | %s\n", i+1, p.ID, p.Name)
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Select project number (1-%d): ", len(projects))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("input read error: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(projects) {
		return 0, fmt.Errorf("invalid choice: %s (expected number from 1 to %d)", input, len(projects))
	}

	selected := projects[choice-1]
	fmt.Printf("✓ Selected project: %s (ID: %d)\n\n", selected.Name, selected.ID)

	return selected.ID, nil
}

// selectSuiteInteractively показывает список сьютов проекта и просит выбрать
func selectSuiteInteractively(ctx context.Context, cli client.ClientInterface, projectID int64, prompt string) (int64, error) {
	suites, err := cli.GetSuites(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get suites for project %d: %w", projectID, err)
	}

	if len(suites) == 0 {
		return 0, fmt.Errorf("no suites found in project %d", projectID)
	}

	// Если только один сьют — выбираем автоматически
	if len(suites) == 1 {
		fmt.Printf("Project has one suite: %s (ID: %d)\n✓ Using automatically.\n\n",
			suites[0].Name, suites[0].ID)
		return suites[0].ID, nil
	}

	fmt.Printf("\n%s\n", prompt)
	fmt.Println(strings.Repeat("-", 70))

	for i, s := range suites {
		fmt.Printf("  [%d] ID: %d | %s\n", i+1, s.ID, s.Name)
		if s.Description != "" {
			desc := s.Description
			if len(desc) > 45 {
				desc = desc[:42] + "..."
			}
			fmt.Printf("       %s\n", desc)
		}
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Select suite number (1-%d): ", len(suites))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("input read error: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(suites) {
		return 0, fmt.Errorf("invalid choice: %s (expected number from 1 to %d)", input, len(suites))
	}

	selected := suites[choice-1]
	fmt.Printf("✓ Selected suite: %s (ID: %d)\n\n", selected.Name, selected.ID)

	return selected.ID, nil
}
