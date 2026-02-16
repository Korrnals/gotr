package compare

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ItemInfo represents a generic item with ID and name
type ItemInfo struct {
	ID   int64  `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
}

// CommonItemInfo represents a common item found in both projects
type CommonItemInfo struct {
	Name     string `json:"name" yaml:"name"`
	ID1      int64  `json:"id1" yaml:"id1"`
	ID2      int64  `json:"id2" yaml:"id2"`
	IDsMatch bool   `json:"ids_match" yaml:"ids_match"`
}

// CompareResult represents the result of comparing resources between two projects
type CompareResult struct {
	Resource     string           `json:"resource" yaml:"resource"`
	Project1ID   int64            `json:"project1_id" yaml:"project1_id"`
	Project2ID   int64            `json:"project2_id" yaml:"project2_id"`
	OnlyInFirst  []ItemInfo       `json:"only_in_first" yaml:"only_in_first"`
	OnlyInSecond []ItemInfo       `json:"only_in_second" yaml:"only_in_second"`
	Common       []CommonItemInfo `json:"common" yaml:"common"`
}

// ResourceDiff represents a diff between two resources (legacy structure)
type ResourceDiff struct {
	Resource    string   `json:"resource"`
	TotalFirst  int      `json:"total_first"`
	TotalSecond int      `json:"total_second"`
	Common      []string `json:"common"`
	OnlyFirst   []string `json:"only_first"`
	OnlySecond  []string `json:"only_second"`
}

// GetProjectNames retrieves project names for both project IDs
func GetProjectNames(cli client.ClientInterface, pid1, pid2 int64) (string, string, error) {
	proj1, err := cli.GetProject(pid1)
	if err != nil {
		return "", "", fmt.Errorf("ошибка получения проекта %d: %w", pid1, err)
	}

	proj2, err := cli.GetProject(pid2)
	if err != nil {
		return "", "", fmt.Errorf("ошибка получения проекта %d: %w", pid2, err)
	}

	return proj1.Name, proj2.Name, nil
}

// PrintCompareResult prints or saves a compare result
func PrintCompareResult(result CompareResult, project1Name, project2Name, format, savePath string) error {
	// If save path is provided, save to file
	if savePath != "" {
		return saveCompareResult(result, format, savePath)
	}

	// Otherwise, print to stdout
	switch format {
	case "json":
		return printJSON(result)
	case "yaml":
		return printYAML(result)
	case "csv":
		return printCSV(result)
	default:
		return printTable(result, project1Name, project2Name)
	}
}

// printTable prints the result in table format
func printTable(result CompareResult, project1Name, project2Name string) error {
	fmt.Printf("\n=== Сравнение ресурса: %s ===\n", result.Resource)
	fmt.Printf("Проект 1: %s (ID: %d)\n", project1Name, result.Project1ID)
	fmt.Printf("Проект 2: %s (ID: %d)\n\n", project2Name, result.Project2ID)

	fmt.Printf("Только в проекте 1 (%d):\n", len(result.OnlyInFirst))
	if len(result.OnlyInFirst) == 0 {
		fmt.Println("  (нет)")
	} else {
		for _, item := range result.OnlyInFirst {
			fmt.Printf("  - %s (ID: %d)\n", item.Name, item.ID)
		}
	}

	fmt.Printf("\nТолько в проекте 2 (%d):\n", len(result.OnlyInSecond))
	if len(result.OnlyInSecond) == 0 {
		fmt.Println("  (нет)")
	} else {
		for _, item := range result.OnlyInSecond {
			fmt.Printf("  - %s (ID: %d)\n", item.Name, item.ID)
		}
	}

	fmt.Printf("\nОбщие элементы (%d):\n", len(result.Common))
	if len(result.Common) == 0 {
		fmt.Println("  (нет)")
	} else {
		for _, item := range result.Common {
			matchStr := ""
			if !item.IDsMatch {
				matchStr = fmt.Sprintf(" [разные ID: %d vs %d]", item.ID1, item.ID2)
			}
			fmt.Printf("  - %s (ID: %d vs %d)%s\n", item.Name, item.ID1, item.ID2, matchStr)
		}
	}

	return nil
}

// printJSON prints the result as JSON
func printJSON(result CompareResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// printYAML prints the result as YAML
func printYAML(result CompareResult) error {
	data, err := yaml.Marshal(result)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга YAML: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// printCSV prints the result as CSV
func printCSV(result CompareResult) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Header
	if err := writer.Write([]string{"Type", "Name", "ID Project 1", "ID Project 2"}); err != nil {
		return err
	}

	// Only in first
	for _, item := range result.OnlyInFirst {
		if err := writer.Write([]string{"Only in Project 1", item.Name, fmt.Sprintf("%d", item.ID), ""}); err != nil {
			return err
		}
	}

	// Only in second
	for _, item := range result.OnlyInSecond {
		if err := writer.Write([]string{"Only in Project 2", item.Name, "", fmt.Sprintf("%d", item.ID)}); err != nil {
			return err
		}
	}

	// Common
	for _, item := range result.Common {
		if err := writer.Write([]string{"Common", item.Name, fmt.Sprintf("%d", item.ID1), fmt.Sprintf("%d", item.ID2)}); err != nil {
			return err
		}
	}

	return nil
}

// saveCompareResult saves the result to a file
func saveCompareResult(result CompareResult, format, savePath string) error {
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(result, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(result)
	case "csv":
		return saveCSV(result, savePath)
	default:
		return fmt.Errorf("формат '%s' не поддерживается для сохранения", format)
	}

	if err != nil {
		return fmt.Errorf("ошибка форматирования: %w", err)
	}

	return saveToFile(data, savePath)
}

// saveCSV saves the result as CSV to a file
func saveCSV(result CompareResult, savePath string) error {
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	if err := writer.Write([]string{"Type", "Name", "ID Project 1", "ID Project 2"}); err != nil {
		return err
	}

	// Only in first
	for _, item := range result.OnlyInFirst {
		if err := writer.Write([]string{"Only in Project 1", item.Name, fmt.Sprintf("%d", item.ID), ""}); err != nil {
			return err
		}
	}

	// Only in second
	for _, item := range result.OnlyInSecond {
		if err := writer.Write([]string{"Only in Project 2", item.Name, "", fmt.Sprintf("%d", item.ID)}); err != nil {
			return err
		}
	}

	// Common
	for _, item := range result.Common {
		if err := writer.Write([]string{"Common", item.Name, fmt.Sprintf("%d", item.ID1), fmt.Sprintf("%d", item.ID2)}); err != nil {
			return err
		}
	}

	fmt.Printf("Результат сохранён в %s\n", savePath)
	return nil
}

// saveToFile saves data to a file
func saveToFile(data []byte, savePath string) error {
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи файла: %w", err)
	}
	fmt.Printf("Результат сохранён в %s\n", savePath)
	return nil
}

// GetProjectName retrieves a single project name (helper for tests)
func GetProjectName(cli client.ClientInterface, projectID int64) (string, error) {
	proj, err := cli.GetProject(projectID)
	if err != nil {
		return "", fmt.Errorf("ошибка получения проекта %d: %w", projectID, err)
	}
	if proj == nil {
		return fmt.Sprintf("Project %d", projectID), nil
	}
	return proj.Name, nil
}

// collectNames collects non-empty names from a slice using a getter function
func collectNames(size int, getter func(int) string) []string {
	if size == 0 {
		return nil
	}
	names := make([]string, 0, size)
	for i := 0; i < size; i++ {
		name := strings.TrimSpace(getter(i))
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

// IDMappingPair represents a pair of IDs for mapping between projects
type IDMappingPair struct {
	ID1  int64  `json:"id1"`
	ID2  int64  `json:"id2"`
	Name string `json:"name"`
}

// getClientSafe safely retrieves the client from the package variable
func getClientSafe(cmd *cobra.Command) client.ClientInterface {
	if getClient != nil {
		return getClient(cmd)
	}
	return nil
}

// parseFlags parses common flags for compare commands
func parseFlags(cmd *cobra.Command) (pid1, pid2 int64, field string, err error) {
	pid1Str, _ := cmd.Flags().GetString("pid1")
	pid1, err = strconv.ParseInt(pid1Str, 10, 64)
	if err != nil || pid1 <= 0 {
		return 0, 0, "", fmt.Errorf("укажите корректный pid1 (--pid1)")
	}

	pid2Str, _ := cmd.Flags().GetString("pid2")
	pid2, err = strconv.ParseInt(pid2Str, 10, 64)
	if err != nil || pid2 <= 0 {
		return 0, 0, "", fmt.Errorf("укажите корректный pid2 (--pid2)")
	}

	field, _ = cmd.Flags().GetString("field")
	if field == "" {
		field = "title"
	}

	return pid1, pid2, field, nil
}

// buildResourceDiff builds a ResourceDiff from two string slices
func buildResourceDiff(resource string, first, second []string) ResourceDiff {
	firstSet := make(map[string]bool)
	secondSet := make(map[string]bool)

	for _, f := range first {
		firstSet[strings.ToLower(strings.TrimSpace(f))] = true
	}

	for _, s := range second {
		secondSet[strings.ToLower(strings.TrimSpace(s))] = true
	}

	var common, onlyFirst, onlySecond []string

	for f := range firstSet {
		if secondSet[f] {
			common = append(common, f)
		} else {
			onlyFirst = append(onlyFirst, f)
		}
	}

	for s := range secondSet {
		if !firstSet[s] {
			onlySecond = append(onlySecond, s)
		}
	}

	return ResourceDiff{
		Resource:    resource,
		TotalFirst:  len(firstSet),
		TotalSecond: len(secondSet),
		Common:      common,
		OnlyFirst:   onlyFirst,
		OnlySecond:  onlySecond,
	}
}
