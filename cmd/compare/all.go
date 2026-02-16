package compare

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
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

			// Compare all resources
			result := &allResult{}
			errors := make(map[string]error)

			// Cases
			if casesResult, err := compareCasesInternal(cli, pid1, pid2, "title"); err == nil {
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

			// Print summary
			fmt.Printf("\n=== Сводный отчёт сравнения проектов ===\n")
			fmt.Printf("Проект 1: %s (ID: %d)\n", project1Name, pid1)
			fmt.Printf("Проект 2: %s (ID: %d)\n\n", project2Name, pid2)

			printResourceSummary("Cases", result.Cases)
			printResourceSummary("Suites", result.Suites)
			printResourceSummary("Sections", result.Sections)
			printResourceSummary("Shared Steps", result.SharedSteps)
			printResourceSummary("Runs", result.Runs)
			printResourceSummary("Plans", result.Plans)
			printResourceSummary("Milestones", result.Milestones)
			printResourceSummary("Datasets", result.Datasets)
			printResourceSummary("Groups", result.Groups)
			printResourceSummary("Labels", result.Labels)
			printResourceSummary("Templates", result.Templates)
			printResourceSummary("Configurations", result.Configurations)

			// Print errors if any
			if len(errors) > 0 {
				fmt.Printf("\n=== Ошибки ===\n")
				for resource, err := range errors {
					fmt.Printf("- %s: %v\n", resource, err)
				}
			}

			// Save result if requested
			if savePath != "" {
				// For saving, default to json if table format is specified
				saveFormat := format
				if saveFormat == "table" || saveFormat == "" {
					saveFormat = "json"
				}

				if savePath == "__DEFAULT__" {
					// --save flag was used, save to default location
					filepath, err := save.Output(cmd, result, "compare", saveFormat)
					if err != nil {
						return err
					}
					fmt.Printf("Результат сохранён в %s\n", filepath)
					return nil
				}
				// --save-to flag was used with custom path
				if err := saveAllResult(result, saveFormat, savePath); err != nil {
					return err
				}
				fmt.Printf("Результат сохранён в %s\n", savePath)
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


