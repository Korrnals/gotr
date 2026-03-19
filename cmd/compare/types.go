package compare

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	outpututils "github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// CompareStatus describes whether the comparison completed fully.
// Stored in the output JSON so CI/CD pipelines can check it reliably.
type CompareStatus string

const (
	// CompareStatusComplete means all data was fetched without interruption or page errors.
	CompareStatusComplete CompareStatus = "complete"
	// CompareStatusInterrupted means the command was interrupted (e.g. Ctrl+C).
	// Data in the output file is PARTIAL and must not be treated as authoritative.
	CompareStatusInterrupted CompareStatus = "interrupted"
	// CompareStatusPartial means the command finished but some pages had permanent fetch errors.
	// Data may be missing entries from the failed pages.
	CompareStatusPartial CompareStatus = "partial"
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
	Status       CompareStatus    `json:"status" yaml:"status"`
	OnlyInFirst  []ItemInfo       `json:"only_in_first" yaml:"only_in_first"`
	OnlyInSecond []ItemInfo       `json:"only_in_second" yaml:"only_in_second"`
	Common       []CommonItemInfo `json:"common" yaml:"common"`
}

// GetProjectNames retrieves project names for both project IDs
func GetProjectNames(ctx context.Context, cli client.ClientInterface, pid1, pid2 int64) (string, string, error) {
	proj1, err := cli.GetProject(ctx, pid1)
	if err != nil {
		return "", "", fmt.Errorf("failed to get project %d: %w", pid1, err)
	}

	proj2, err := cli.GetProject(ctx, pid2)
	if err != nil {
		return "", "", fmt.Errorf("failed to get project %d: %w", pid2, err)
	}

	return proj1.Name, proj2.Name, nil
}

// getFormatFromExtension extracts format from file extension
func getFormatFromExtension(path string) string {
	path = strings.ToLower(path)
	if strings.HasSuffix(path, ".json") {
		return "json"
	}
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		return "yaml"
	}
	if strings.HasSuffix(path, ".csv") {
		return "csv"
	}
	if strings.HasSuffix(path, ".txt") {
		return "table"
	}
	return ""
}

// PrintCompareResult prints or saves a compare result
// savePath can be:
//   - "__DEFAULT__" : save to default location (~/.gotr/exports/) as table
//   - custom path   : save to specified path (format from --format flag or detected from extension)
//   - ""            : print to stdout
func PrintCompareResult(cmd *cobra.Command, result CompareResult, project1Name, project2Name, format, savePath string) error {
	// If save path is provided (flag was used), save to file
	if savePath != "" {
		if savePath == "__DEFAULT__" {
			// --save flag was used, check format to determine output type
			switch format {
			case "json", "yaml", "csv":
				// Save in structured format with auto-generated filename
				exportsDir, _ := outpututils.GetExportsDir("compare")
				os.MkdirAll(exportsDir, 0755)
				filePath := exportsDir + "/" + outpututils.GenerateFilename("compare", format)
				if err := saveToFileWithPath(result, format, filePath); err != nil {
					return err
				}
				// Message is printed by saveToFile
				return nil
			default:
				// "table" or unknown - save as text table
				return saveTableToFile(cmd, result, project1Name, project2Name)
			}
		}
		// --save-to flag was used with custom path
		// If format is "table" (default), try to detect from file extension
		if format == "table" {
			if detected := getFormatFromExtension(savePath); detected != "" {
				format = detected
			}
		}
		switch format {
		case "json", "yaml", "csv":
			if err := saveToFileWithPath(result, format, savePath); err != nil {
				return err
			}
			if q, _ := cmd.Flags().GetBool("quiet"); !q {
				fmt.Println()
				ui.Infof(os.Stdout, "Result saved to %s", savePath)
			}
		case "table":
			// Save table as text
			return saveTableToFile(cmd, result, project1Name, project2Name, savePath)
		default:
			return fmt.Errorf("unsupported format: %s", format)
		}
		return nil
	}

	// Otherwise, print to stdout
	switch format {
	case "json":
		return ui.JSON(cmd, result)
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
	fmt.Printf("\n=== Comparison: %s (projects %d ↔ %d) ===\n\n", result.Resource, result.Project1ID, result.Project2ID)

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
	// Column widths are widened for long names.
	idWidth := 8
	nameWidth := 70

	widths := []int{idWidth, nameWidth}
	// totalInnerWidth = sum of column widths + 3 per column - 1 (for proper border alignment)
	totalInnerWidth := idWidth + nameWidth + 3*len(widths) - 1

	// Title
	title := fmt.Sprintf("Only in project %d - \"%s\"", projectID, projectName)
	printHorizontalBorder("┌", "┬", "┐", widths)
	printHeader(title, totalInnerWidth)

	// If no items, show empty message
	if len(items) == 0 {
		printSeparator(widths)
		printHeader("(none)", totalInnerWidth)
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
	// Column widths are widened for long names.
	nameWidth := 50
	id1Width := 12
	id2Width := 12
	statusWidth := 20

	widths := []int{nameWidth, id1Width, id2Width, statusWidth}
	// totalInnerWidth = sum of column widths + 3 per column - 1 (for proper border alignment)
	totalInnerWidth := nameWidth + id1Width + id2Width + statusWidth + 3*len(widths) - 1

	// Title
	printHorizontalBorder("┌", "┬", "┐", widths)
	printHeader("Common in both projects", totalInnerWidth)

	// If no items, show empty message
	if len(items) == 0 {
		printSeparator(widths)
		printHeader("(none)", totalInnerWidth)
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
		"ID status",
	}, widths)
	printSeparator(widths)

	// Data rows
	for _, item := range items {
		status := "✓ Match"
		if !item.IDsMatch {
			status = "⚠ Differ"
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
	printHeader("ID mapping (for updates)", totalInnerWidth)

	// If no mappings, don't show the table at all (or show empty message)
	if len(mappings) == 0 {
		printSeparator(widths)
		printHeader("(all IDs match)", totalInnerWidth)
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
// Deprecated: use ui.JSON(cmd, result) directly in new code
func printJSON(result CompareResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// printYAML prints the result as YAML
func printYAML(result CompareResult) error {
	data, err := yaml.Marshal(result)
	if err != nil {
		return fmt.Errorf("YAML marshal error: %w", err)
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
		return fmt.Errorf("format '%s' not supported for save", format)
	}

	if err != nil {
		return fmt.Errorf("formatting error: %w", err)
	}

	return saveToFile(data, savePath)
}

// saveCSV saves the result as CSV to a file
func saveCSV(result CompareResult, savePath string) error {
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("file create error: %w", err)
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

	return nil
}

// saveToFile saves data to a file.
// Callers are responsible for printing confirmation (respecting quiet flag).
func saveToFile(data []byte, savePath string) error {
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return fmt.Errorf("file write error: %w", err)
	}
	return nil
}

// saveTableToFile saves the table output to a file
func saveTableToFile(cmd *cobra.Command, result CompareResult, project1Name, project2Name string, customPath ...string) error {
	// Create pipe to capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("pipe create error: %w", err)
	}
	os.Stdout = w

	// Capture output in goroutine
	outChan := make(chan string, 1)
	errChan := make(chan error, 1)
	go func() {
		var buf strings.Builder
		_, err := io.Copy(&buf, r)
		if err != nil {
			errChan <- err
			return
		}
		outChan <- buf.String()
	}()

	// Print table (writes to stdout)
	printErr := printTable(result, project1Name, project2Name)

	// Restore stdout and close writer
	w.Close()
	os.Stdout = oldStdout

	// Get captured output
	var output string
	select {
	case output = <-outChan:
	case err := <-errChan:
		return fmt.Errorf("output read error: %w", err)
	}

	if printErr != nil {
		return printErr
	}

	// Determine file path
	var filePath string
	if len(customPath) > 0 && customPath[0] != "" {
		filePath = customPath[0]
	} else {
		// Use default path with .txt extension for table
		filePath = outpututils.GenerateFilename("compare", "txt")
		exportsDir, _ := outpututils.GetExportsDir("compare")
		os.MkdirAll(exportsDir, 0755)
		filePath = exportsDir + "/" + filePath
	}

	// Write to file
	if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
		return fmt.Errorf("file write error: %w", err)
	}

	if q, _ := cmd.Flags().GetBool("quiet"); !q {
		fmt.Println()
		ui.Infof(os.Stdout, "Result saved to %s", filePath)
	}
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
			return fmt.Errorf("JSON marshal error: %w", err)
		}
	case "yaml":
		data, err = yaml.Marshal(result)
		if err != nil {
			return fmt.Errorf("YAML marshal error: %w", err)
		}
	case "csv":
		return saveCSV(result, savePath)
	default:
		// Default to JSON for unknown formats
		data, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("JSON marshal error: %w", err)
		}
	}

	return saveToFile(data, savePath)
}

// GetProjectName retrieves a single project name (helper for tests)
func GetProjectName(cli client.ClientInterface, projectID int64) (string, error) {
	proj, err := cli.GetProject(context.Background(), projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get project %d: %w", projectID, err)
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
	pid1, err = flags.ParseID(pid1Str)
	if err != nil || pid1 <= 0 {
		return 0, 0, "", fmt.Errorf("specify valid pid1 (--pid1)")
	}

	pid2Str, _ := cmd.Flags().GetString("pid2")
	pid2, err = flags.ParseID(pid2Str)
	if err != nil || pid2 <= 0 {
		return 0, 0, "", fmt.Errorf("specify valid pid2 (--pid2)")
	}

	field, _ = cmd.Flags().GetString("field")
	if field == "" {
		field = "title"
	}

	return pid1, pid2, field, nil
}
