package compare

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
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
// savePath can be:
//   - "__DEFAULT__" : save to default location (~/.gotr/exports/)
//   - custom path   : save to specified path
//   - ""            : print to stdout
func PrintCompareResult(cmd *cobra.Command, result CompareResult, project1Name, project2Name, format, savePath string) error {
	// If save path is provided (flag was used), save to file
	if savePath != "" {
		if savePath == "__DEFAULT__" {
			// --save flag was used, save to default location
			_, err := save.Output(cmd, result, "compare", format)
			return err
		}
		// --save-to flag was used with custom path
		return saveToFileWithPath(result, format, savePath)
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

// tableCell represents a cell in the table with content and width
type tableCell struct {
	content string
	width   int
}

// truncateString truncates a string to maxWidth with ellipsis if needed
func truncateString(s string, maxWidth int) string {
	if maxWidth <= 3 {
		if utf8.RuneCountInString(s) > maxWidth {
			return string([]rune(s)[:maxWidth])
		}
		return s
	}
	if utf8.RuneCountInString(s) > maxWidth {
		return string([]rune(s)[:maxWidth-3]) + "..."
	}
	return s
}

// padRight pads a string to the right to reach target width
func padRight(s string, width int) string {
	runeCount := utf8.RuneCountInString(s)
	if runeCount >= width {
		return s
	}
	return s + strings.Repeat(" ", width-runeCount)
}

// printHorizontalBorder prints a horizontal border line
func printHorizontalBorder(left, mid, right string, widths []int) {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("─", w+2)
	}
	fmt.Println(left + strings.Join(parts, mid) + right)
}

// printRow prints a data row with given widths
func printRow(cells []string, widths []int) {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		parts[i] = " " + padRight(truncateString(cell, widths[i]), widths[i]) + " "
	}
	fmt.Println("│" + strings.Join(parts, "│") + "│")
}

// printHeader prints a header row (title spanning all columns)
func printHeader(title string, totalWidth int) {
	titleWidth := utf8.RuneCountInString(title)
	padding := totalWidth - 2 - titleWidth
	if padding < 0 {
		padding = 0
		title = truncateString(title, totalWidth-2)
	}
	leftPad := padding / 2
	rightPad := padding - leftPad
	fmt.Println("│" + strings.Repeat(" ", leftPad) + title + strings.Repeat(" ", rightPad) + "│")
}

// printSeparator prints a separator line between header and data
func printSeparator(widths []int) {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("─", w+2)
	}
	fmt.Println("├" + strings.Join(parts, "┼") + "┤")
}

// printTable prints the result in table format
func printTable(result CompareResult, project1Name, project2Name string) error {
	fmt.Printf("\n=== Сравнение: %s (проекты %d ↔ %d) ===\n\n", result.Resource, result.Project1ID, result.Project2ID)

	// Table 1: Only in Project 1
	printOnlyInProjectTable(result.OnlyInFirst, result.Project1ID, project1Name)

	// Table 2: Only in Project 2
	printOnlyInProjectTable(result.OnlyInSecond, result.Project2ID, project2Name)

	// Table 3: Common items
	printCommonTable(result.Common, result.Project1ID, result.Project2ID)

	// Table 4: ID Mapping (for items with different IDs)
	printIDMappingTable(result.Common)

	return nil
}

// printOnlyInProjectTable prints a table for items only in one project
func printOnlyInProjectTable(items []ItemInfo, projectID int64, projectName string) {
	// Column widths - increased for long Russian names
	idWidth := 8
	nameWidth := 70

	widths := []int{idWidth, nameWidth}
	// totalInnerWidth = sum of column widths + 3 per column - 1 (for proper border alignment)
	totalInnerWidth := idWidth + nameWidth + 3*len(widths) - 1

	// Title
	title := fmt.Sprintf("Только в проекте %d - \"%s\"", projectID, projectName)
	printHorizontalBorder("┌", "┬", "┐", widths)
	printHeader(title, totalInnerWidth)

	// If no items, show empty message
	if len(items) == 0 {
		printSeparator(widths)
		printHeader("(нет)", totalInnerWidth)
		printHorizontalBorder("└", "┴", "┘", widths)
		fmt.Println()
		return
	}

	// Column headers
	printSeparator(widths)
	printRow([]string{"ID", "Name"}, widths)
	printSeparator(widths)

	// Data rows
	for _, item := range items {
		printRow([]string{fmt.Sprintf("%d", item.ID), item.Name}, widths)
	}

	printHorizontalBorder("└", "┴", "┘", widths)
	fmt.Println()
}

// printCommonTable prints a table for common items
func printCommonTable(items []CommonItemInfo, project1ID, project2ID int64) {
	// Column widths - increased for long Russian names
	nameWidth := 50
	id1Width := 12
	id2Width := 12
	statusWidth := 20

	widths := []int{nameWidth, id1Width, id2Width, statusWidth}
	// totalInnerWidth = sum of column widths + 3 per column - 1 (for proper border alignment)
	totalInnerWidth := nameWidth + id1Width + id2Width + statusWidth + 3*len(widths) - 1

	// Title
	printHorizontalBorder("┌", "┬", "┐", widths)
	printHeader("Общие в обоих проектах", totalInnerWidth)

	// If no items, show empty message
	if len(items) == 0 {
		printSeparator(widths)
		printHeader("(нет)", totalInnerWidth)
		printHorizontalBorder("└", "┴", "┘", widths)
		fmt.Println()
		return
	}

	// Column headers
	printSeparator(widths)
	printRow([]string{
		"Name",
		fmt.Sprintf("ID proj %d", project1ID),
		fmt.Sprintf("ID proj %d", project2ID),
		"Статус ID",
	}, widths)
	printSeparator(widths)

	// Data rows
	for _, item := range items {
		status := "✓ Совпадают"
		if !item.IDsMatch {
			status = "⚠ Различаются"
		}
		printRow([]string{
			item.Name,
			fmt.Sprintf("%d", item.ID1),
			fmt.Sprintf("%d", item.ID2),
			status,
		}, widths)
	}

	printHorizontalBorder("└", "┴", "┘", widths)
	fmt.Println()
}

// printIDMappingTable prints a table for ID mapping (items with different IDs)
func printIDMappingTable(items []CommonItemInfo) {
	// Filter items with different IDs
	var mappings []CommonItemInfo
	for _, item := range items {
		if !item.IDsMatch {
			mappings = append(mappings, item)
		}
	}

	// Column widths
	sourceWidth := 12
	targetWidth := 12
	nameWidth := 70

	widths := []int{sourceWidth, targetWidth, nameWidth}
	totalInnerWidth := sourceWidth + targetWidth + nameWidth + 3*len(widths) - 1

	// Title
	printHorizontalBorder("┌", "┬", "┐", widths)
	printHeader("Маппинг ID (для обновления)", totalInnerWidth)

	// If no mappings, don't show the table at all (or show empty message)
	if len(mappings) == 0 {
		printSeparator(widths)
		printHeader("(все ID совпадают)", totalInnerWidth)
		printHorizontalBorder("└", "┴", "┘", widths)
		fmt.Println()
		return
	}

	// Column headers
	printSeparator(widths)
	printRow([]string{"Source ID", "Target ID", "Name"}, widths)
	printSeparator(widths)

	// Data rows - show only items with different IDs
	for _, item := range mappings {
		printRow([]string{
			fmt.Sprintf("%d", item.ID1),
			fmt.Sprintf("%d", item.ID2),
			item.Name,
		}, widths)
	}

	printHorizontalBorder("└", "┴", "┘", widths)
	fmt.Println()
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

// saveToFileWithPath saves the result to a specific file path
func saveToFileWithPath(result CompareResult, format, savePath string) error {
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("ошибка маршалинга JSON: %w", err)
		}
	case "yaml":
		data, err = yaml.Marshal(result)
		if err != nil {
			return fmt.Errorf("ошибка маршалинга YAML: %w", err)
		}
	case "csv":
		return saveCSV(result, savePath)
	default:
		// Default to JSON for unknown formats
		data, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("ошибка маршалинга JSON: %w", err)
		}
	}

	return saveToFile(data, savePath)
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
