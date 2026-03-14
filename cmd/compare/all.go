package compare

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/flags"
	outpututils "github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/Korrnals/gotr/pkg/reporter"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			// Parse flags
			pid1, pid2, format, savePath, err := parseCommonFlags(cmd)
			if err != nil {
				return err
			}

			// Get project names
			project1Name, project2Name, err := GetProjectNames(ctx, cli, pid1, pid2)
			if err != nil {
				return err
			}

			startTime := time.Now()

			// Compare all resources
			result := &allResult{}
			errors := make(map[string]error)

			// Cases
			if casesResult, _, err := compareCasesInternal(ctx, cmd, cli, pid1, pid2, "title"); err == nil {
				result.Cases = casesResult
			} else {
				errors["cases"] = err
			}

			// Suites
			if suitesResult, err := compareSuitesInternal(ctx, cli, pid1, pid2); err == nil {
				result.Suites = suitesResult
			} else {
				errors["suites"] = err
			}

			// Sections
			if sectionsResult, err := compareSimpleInternal(cli, pid1, pid2, "sections", fetchSectionItems); err == nil {
				result.Sections = sectionsResult
			} else {
				errors["sections"] = err
			}

			// Shared Steps
			if sharedStepsResult, err := compareSimpleInternal(cli, pid1, pid2, "sharedsteps", fetchSharedStepItems); err == nil {
				result.SharedSteps = sharedStepsResult
			} else {
				errors["shared_steps"] = err
			}

			// Runs
			if runsResult, err := compareSimpleInternal(cli, pid1, pid2, "runs", fetchRunItems); err == nil {
				result.Runs = runsResult
			} else {
				errors["runs"] = err
			}

			// Plans
			if plansResult, err := compareSimpleInternal(cli, pid1, pid2, "plans", fetchPlanItems); err == nil {
				result.Plans = plansResult
			} else {
				errors["plans"] = err
			}

			// Milestones
			if milestonesResult, err := compareSimpleInternal(cli, pid1, pid2, "milestones", fetchMilestoneItems); err == nil {
				result.Milestones = milestonesResult
			} else {
				errors["milestones"] = err
			}

			// Datasets
			if datasetsResult, err := compareSimpleInternal(cli, pid1, pid2, "datasets", fetchDatasetItems); err == nil {
				result.Datasets = datasetsResult
			} else {
				errors["datasets"] = err
			}

			// Groups
			if groupsResult, err := compareSimpleInternal(cli, pid1, pid2, "groups", fetchGroupItems); err == nil {
				result.Groups = groupsResult
			} else {
				errors["groups"] = err
			}

			// Labels
			if labelsResult, err := compareSimpleInternal(cli, pid1, pid2, "labels", fetchLabelItems); err == nil {
				result.Labels = labelsResult
			} else {
				errors["labels"] = err
			}

			// Templates
			if templatesResult, err := compareSimpleInternal(cli, pid1, pid2, "templates", fetchTemplateItems); err == nil {
				result.Templates = templatesResult
			} else {
				errors["templates"] = err
			}

			// Configurations
			if configsResult, err := compareSimpleInternal(cli, pid1, pid2, "configurations", fetchConfigurationItems); err == nil {
				result.Configurations = configsResult
			} else {
				errors["configurations"] = err
			}

			// Print summary table
			elapsed := time.Since(startTime)
			printAllSummaryTable(project1Name, pid1, project2Name, pid2, result, errors, elapsed)

			// Save result if requested
			if savePath != "" {
				if savePath == "__DEFAULT__" {
					// --save flag was used, check format to determine output type
					switch format {
					case "json", "yaml":
						// Save in structured format with auto-generated filename
						exportsDir, _ := outpututils.GetExportsDir("compare")
						os.MkdirAll(exportsDir, 0755)
						filePath := exportsDir + "/" + outpututils.GenerateFilename("compare", format)
						if err := saveAllResult(result, format, filePath); err != nil {
							return err
						}
						// Message is printed by saveToFile via saveAllResult
						return nil
					default:
						// "table" or unknown - save as text summary
						return saveAllSummaryToFile(cmd, result, project1Name, pid1, project2Name, pid2, errors, "__DEFAULT__", time.Since(startTime))
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
				case "json", "yaml":
					if err := saveAllResult(result, format, savePath); err != nil {
						return err
					}
					// Print on new line after progress bar
					fmt.Println()
					ui.Infof(os.Stdout, "Result saved to %s", savePath)
				case "table":
					return saveAllSummaryToFile(cmd, result, project1Name, pid1, project2Name, pid2, errors, savePath, time.Since(startTime))
				default:
					return fmt.Errorf("unsupported format '%s' for save, use json, yaml or table", format)
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
	pid1, err = flags.ParseID(pid1Str)
	if err != nil || pid1 <= 0 {
		return 0, 0, "", "", fmt.Errorf("specify valid pid1 (--pid1)")
	}

	pid2Str, _ := cmd.Flags().GetString("pid2")
	pid2, err = flags.ParseID(pid2Str)
	if err != nil || pid2 <= 0 {
		return 0, 0, "", "", fmt.Errorf("specify valid pid2 (--pid2)")
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

// printAllSummaryTable prints a formatted table summary for compare all
// using go-pretty tables and reporter for consistent aligned output.
func printAllSummaryTable(project1Name string, pid1 int64, project2Name string, pid2 int64, result *allResult, errors map[string]error, elapsed time.Duration) {
	// Header via reporter
	rpt := reporter.New("Comparison проектов").
		Section("Проекты").
		StatFmt("📋", "Проект 1", "%s (ID: %d)", project1Name, pid1).
		StatFmt("📋", "Проект 2", "%s (ID: %d)", project2Name, pid2)
	rpt.Print()

	// Resource comparison table via go-pretty
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(table.StyleRounded)
	tw.SetTitle("СВОДКА РЕСУРСОВ")
	tw.Style().Title.Align = text.AlignCenter

	tw.AppendHeader(table.Row{"Resource", "Only in P1", "Only in P2", "Common", "Status"})
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignLeft, WidthMin: 16},
		{Number: 2, Align: text.AlignRight},
		{Number: 3, Align: text.AlignRight},
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignCenter},
	})

	appendResourceRow(tw, "Cases", result.Cases)
	appendResourceRow(tw, "Suites", result.Suites)
	appendResourceRow(tw, "Sections", result.Sections)
	appendResourceRow(tw, "Shared Steps", result.SharedSteps)
	appendResourceRow(tw, "Runs", result.Runs)
	appendResourceRow(tw, "Plans", result.Plans)
	appendResourceRow(tw, "Milestones", result.Milestones)
	appendResourceRow(tw, "Datasets", result.Datasets)
	appendResourceRow(tw, "Groups", result.Groups)
	appendResourceRow(tw, "Labels", result.Labels)
	appendResourceRow(tw, "Templates", result.Templates)
	appendResourceRow(tw, "Configurations", result.Configurations)

	fmt.Println()
	tw.Render()
	fmt.Println()

	// Footer stats via reporter
	footer := reporter.New("Итого").
		Section("Время").
		Stat("⏱️", "Execution time", elapsed.Round(time.Second))

	if len(errors) > 0 {
		footer.Section("Ошибки")
		for resource, err := range errors {
			footer.Stat("❌", resource, err)
		}
	}

	footer.Print()
}

// appendResourceRow adds a resource row to the go-pretty table.
func appendResourceRow(tw table.Writer, name string, result *CompareResult) {
	if result == nil {
		tw.AppendRow(table.Row{name, "-", "-", "-", "(X)"})
		return
	}

	onlyP1 := len(result.OnlyInFirst)
	onlyP2 := len(result.OnlyInSecond)
	common := len(result.Common)

	status := "(OK)"
	if onlyP1 > 0 || onlyP2 > 0 {
		status = "(!)"
	}

	tw.AppendRow(table.Row{name, onlyP1, onlyP2, common, status})
}

// saveAllSummaryToFile saves the summary output to a file (for --save or --save-to with table format)
func saveAllSummaryToFile(cmd *cobra.Command, result *allResult, project1Name string, pid1 int64, project2Name string, pid2 int64, errors map[string]error, savePath string, elapsed time.Duration) error {
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

	// Print summary table (writes to stdout)
	printAllSummaryTable(project1Name, pid1, project2Name, pid2, result, errors, elapsed)

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

	// Determine file path
	var filePath string
	if savePath != "__DEFAULT__" {
		filePath = savePath
	} else {
		// Use default path with .txt extension
		filePath = outpututils.GenerateFilename("compare", "txt")
		exportsDir, _ := outpututils.GetExportsDir("compare")
		os.MkdirAll(exportsDir, 0755)
		filePath = exportsDir + "/" + filePath
	}

	// Write to file
	if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
		return fmt.Errorf("file write error: %w", err)
	}

	// Print on new line after progress bar
	fmt.Println()
	ui.Infof(os.Stdout, "Result saved to %s", filePath)
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
		return fmt.Errorf("format '%s' not supported for saving all resources, use json or yaml", format)
	}

	if err != nil {
		return fmt.Errorf("formatting error: %w", err)
	}

	return saveToFile(output, savePath)
}
