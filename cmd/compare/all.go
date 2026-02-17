package compare

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// allResult represents the combined results of comparing all resources.
type allResult struct {
	Cases          *CompareResult `json:"cases,omitempty" yaml:"cases,omitempty"`
	Suites         *CompareResult `json:"suites,omitempty" yaml:"suites,omitempty"`
	Sections       *CompareResult `json:"sections,omitempty" yaml:"sections,omitempty"`
	SharedSteps    *CompareResult `json:"shared_steps,omitempty" yaml:"shared_steps,omitempty"`
	Runs           *CompareResult `json:"runs,omitempty" yaml:"runs,omitempty"`
	Plans          *CompareResult `json:"plans,omitempty" yaml:"plans,omitempty"`
	Milestones     *CompareResult `json:"milestones,omitempty" yaml:"milestones,omitempty"`
	Datasets       *CompareResult `json:"datasets,omitempty" yaml:"datasets,omitempty"`
	Groups         *CompareResult `json:"groups,omitempty" yaml:"groups,omitempty"`
	Labels         *CompareResult `json:"labels,omitempty" yaml:"labels,omitempty"`
	Templates      *CompareResult `json:"templates,omitempty" yaml:"templates,omitempty"`
	Configurations *CompareResult `json:"configurations,omitempty" yaml:"configurations,omitempty"`
}

// newAllCmd creates the 'compare all' subcommand.
func newAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Сравнить все ресурсы между двумя проектами",
		Long: `Выполняет сравнение всех поддерживаемых ресурсов между двумя проектами.

Сравниваются следующие ресурсы:
- cases (кейсы)
- suites (сюиты)
- sections (секции)
- sharedsteps (shared steps)
- runs (test runs)
- plans (test plans)
- milestones (milestones)
- datasets (datasets)
- groups (группы)
- labels (метки)
- templates (шаблоны)
- configurations (конфигурации)

Примеры:
  # Сравнить все ресурсы
  gotr compare all --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare all --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare all --pid1 30 --pid2 31 --save-to result.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClientSafe(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			// Parse flags
			pid1, pid2, format, savePath, err := parseCommonFlags(cmd)
			if err != nil {
				return err
			}

			// Get project names
			project1Name, project2Name, err := GetProjectNames(cli, pid1, pid2)
			if err != nil {
				return err
			}

			// Create progress manager
			pm := progress.NewManager()

			// Compare all resources
			result := &allResult{}
			errors := make(map[string]error)

			// Cases
			if casesResult, err := compareCasesInternal(cli, pid1, pid2, "title", pm); err == nil {
				result.Cases = casesResult
			} else {
				errors["cases"] = err
			}

			// Suites
			if suitesResult, err := compareSuitesInternal(cli, pid1, pid2); err == nil {
				result.Suites = suitesResult
			} else {
				errors["suites"] = err
			}

			// Sections
			if sectionsResult, err := compareSectionsInternal(cli, pid1, pid2); err == nil {
				result.Sections = sectionsResult
			} else {
				errors["sections"] = err
			}

			// Shared Steps
			if sharedStepsResult, err := compareSharedStepsInternal(cli, pid1, pid2); err == nil {
				result.SharedSteps = sharedStepsResult
			} else {
				errors["shared_steps"] = err
			}

			// Runs
			if runsResult, err := compareRunsInternal(cli, pid1, pid2); err == nil {
				result.Runs = runsResult
			} else {
				errors["runs"] = err
			}

			// Plans
			if plansResult, err := comparePlansInternal(cli, pid1, pid2); err == nil {
				result.Plans = plansResult
			} else {
				errors["plans"] = err
			}

			// Milestones
			if milestonesResult, err := compareMilestonesInternal(cli, pid1, pid2); err == nil {
				result.Milestones = milestonesResult
			} else {
				errors["milestones"] = err
			}

			// Datasets
			if datasetsResult, err := compareDatasetsInternal(cli, pid1, pid2); err == nil {
				result.Datasets = datasetsResult
			} else {
				errors["datasets"] = err
			}

			// Groups
			if groupsResult, err := compareGroupsInternal(cli, pid1, pid2); err == nil {
				result.Groups = groupsResult
			} else {
				errors["groups"] = err
			}

			// Labels
			if labelsResult, err := compareLabelsInternal(cli, pid1, pid2); err == nil {
				result.Labels = labelsResult
			} else {
				errors["labels"] = err
			}

			// Templates
			if templatesResult, err := compareTemplatesInternal(cli, pid1, pid2); err == nil {
				result.Templates = templatesResult
			} else {
				errors["templates"] = err
			}

			// Configurations
			if configsResult, err := compareConfigurationsInternal(cli, pid1, pid2); err == nil {
				result.Configurations = configsResult
			} else {
				errors["configurations"] = err
			}

			// Print summary table
				printAllSummaryTable(project1Name, pid1, project2Name, pid2, result, errors)

			// Save result if requested
			if savePath != "" {
				if savePath == "__DEFAULT__" {
					// --save flag was used, save summary as text file
					return saveAllSummaryToFile(cmd, result, project1Name, pid1, project2Name, pid2, errors, "__DEFAULT__")
				}
				// --save-to flag was used with custom path
				// If format is "table" (default), try to detect from file extension
				if format == "table" {
					if detected := getFormatFromExtension(savePath); detected != "" {
						format = detected
					}
				}
				switch format {
				case "json", "yaml":
					if err := saveAllResult(result, format, savePath); err != nil {
						return err
					}
					fmt.Printf("Результат сохранён в %s\n", savePath)
				case "table":
					return saveAllSummaryToFile(cmd, result, project1Name, pid1, project2Name, pid2, errors, savePath)
				default:
					return fmt.Errorf("неподдерживаемый формат '%s' для сохранения, используйте json, yaml или table", format)
				}
				return nil
			}

			return nil
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// allCmd — экспортированная команда
var allCmd = newAllCmd()

// parseCommonFlags parses common flags for all subcommands.
// Returns savePath which can be:
//   - "__DEFAULT__" if --save flag was used (save to default location)
//   - custom path if --save-to flag was used
//   - "" if neither flag was used
func parseCommonFlags(cmd *cobra.Command) (pid1, pid2 int64, format, savePath string, err error) {
	pid1Str, _ := cmd.Flags().GetString("pid1")
	pid1, err = strconv.ParseInt(pid1Str, 10, 64)
	if err != nil || pid1 <= 0 {
		return 0, 0, "", "", fmt.Errorf("укажите корректный pid1 (--pid1)")
	}

	pid2Str, _ := cmd.Flags().GetString("pid2")
	pid2, err = strconv.ParseInt(pid2Str, 10, 64)
	if err != nil || pid2 <= 0 {
		return 0, 0, "", "", fmt.Errorf("укажите корректный pid2 (--pid2)")
	}

	format, _ = cmd.Flags().GetString("format")
	if format == "" {
		format = "table"
	}

	// Check save flags
	if cmd.Flags().Changed("save-to") {
		// --save-to has priority and specifies custom path
		savePath, _ = cmd.Flags().GetString("save-to")
	} else if cmd.Flags().Changed("save") {
		// --save means use default path
		savePath = "__DEFAULT__"
	}

	return pid1, pid2, format, savePath, nil
}

// addCommonFlags marks required flags that are already registered as persistent.
// Note: pid1, pid2, format, save, save-to are registered as persistent flags
// in register.go to ensure they appear in completion.
func addCommonFlags(cmd *cobra.Command) {
	// Mark required flags
	cmd.MarkFlagRequired("pid1")
	cmd.MarkFlagRequired("pid2")
}

// printResourceSummary prints a summary line for a resource comparison.
func printResourceSummary(resourceName string, result *CompareResult) {
	if result == nil {
		fmt.Printf("%-15s: ошибка получения данных\n", resourceName)
		return
	}
	fmt.Printf("%-15s: только в P1: %d, только в P2: %d, общих: %d\n",
		resourceName,
		len(result.OnlyInFirst),
		len(result.OnlyInSecond),
		len(result.Common))
}

// printAllSummaryTable prints a formatted table summary for compare all
func printAllSummaryTable(project1Name string, pid1 int64, project2Name string, pid2 int64, result *allResult, errors map[string]error) {
	// Table header
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    СВОДНЫЙ ОТЧЁТ СРАВНЕНИЯ ПРОЕКТОВ                         ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Проект 1: %-30s (ID: %-5d)              ║\n", truncate(project1Name, 30), pid1)
	fmt.Printf("║  Проект 2: %-30s (ID: %-5d)              ║\n", truncate(project2Name, 30), pid2)
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Resource table header
	fmt.Println("┌────────────────────┬─────────────┬─────────────┬─────────┬──────────┐")
	fmt.Println("│ Ресурс             │ Только в P1 │ Только в P2 │ Общих   │ Статус   │")
	fmt.Println("├────────────────────┼─────────────┼─────────────┼─────────┼──────────┤")

	// Print each resource
	printResourceRow("Cases", result.Cases)
	printResourceRow("Suites", result.Suites)
	printResourceRow("Sections", result.Sections)
	printResourceRow("Shared Steps", result.SharedSteps)
	printResourceRow("Runs", result.Runs)
	printResourceRow("Plans", result.Plans)
	printResourceRow("Milestones", result.Milestones)
	printResourceRow("Datasets", result.Datasets)
	printResourceRow("Groups", result.Groups)
	printResourceRow("Labels", result.Labels)
	printResourceRow("Templates", result.Templates)
	printResourceRow("Configurations", result.Configurations)

	fmt.Println("└────────────────────┴─────────────┴─────────────┴─────────┴──────────┘")

	// Legend
	fmt.Println()
	fmt.Println("Статус:  ✓  - полное совпадение  │  ⚠  - есть отличия  │  ✗  - ошибка загрузки")
	fmt.Println()

	// Print errors if any
	if len(errors) > 0 {
		fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
		fmt.Println("║                              ОШИБКИ                                          ║")
		fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
		for resource, err := range errors {
			fmt.Printf("  • %-15s: %v\n", resource, err)
		}
		fmt.Println()
	}
}

// printResourceRow prints a single resource row for the summary table
func printResourceRow(name string, result *CompareResult) {
	if result == nil {
		fmt.Printf("│ %-18s │     ✗       │     ✗       │    ✗    │    ✗     │\n", name)
		return
	}

	onlyP1 := len(result.OnlyInFirst)
	onlyP2 := len(result.OnlyInSecond)
	common := len(result.Common)

	// Status indicator
	status := "✓"
	if onlyP1 > 0 || onlyP2 > 0 {
		status = "⚠"
	}

	fmt.Printf("│ %-18s │ %11d │ %11d │ %7d │    %-5s │\n", name, onlyP1, onlyP2, common, status)
}

// truncate truncates string to max length with ellipsis
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// saveAllSummaryToFile saves the summary output to a file (for --save or --save-to with table format)
func saveAllSummaryToFile(cmd *cobra.Command, result *allResult, project1Name string, pid1 int64, project2Name string, pid2 int64, errors map[string]error, savePath string) error {
	// Create pipe to capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("ошибка создания pipe: %w", err)
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

	// Print summary table (writes to stdout)
	printAllSummaryTable(project1Name, pid1, project2Name, pid2, result, errors)

	// Restore stdout and close writer
	w.Close()
	os.Stdout = oldStdout

	// Get captured output
	var output string
	select {
	case output = <-outChan:
	case err := <-errChan:
		return fmt.Errorf("ошибка чтения вывода: %w", err)
	}

	// Determine file path
	var filePath string
	if savePath != "__DEFAULT__" {
		filePath = savePath
	} else {
		// Use default path with .txt extension
		filePath = save.GenerateFilename("compare", "txt")
		exportsDir, _ := save.GetExportsDir("compare")
		os.MkdirAll(exportsDir, 0755)
		filePath = exportsDir + "/" + filePath
	}

	// Write to file
	if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
		return fmt.Errorf("ошибка записи файла: %w", err)
	}

	fmt.Printf("Результат сохранён в %s\n", filePath)
	return nil
}

// saveAllResult saves the allResult to a file in the specified format.
func saveAllResult(result *allResult, format, savePath string) error {
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(result, "", "  ")
	case "yaml":
		output, err = yaml.Marshal(result)
	default:
		return fmt.Errorf("формат '%s' не поддерживается для сохранения всех ресурсов, используйте json или yaml", format)
	}

	if err != nil {
		return fmt.Errorf("ошибка форматирования: %w", err)
	}

	return saveToFile(output, savePath)
}


