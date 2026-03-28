package compare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	outpututils "github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/Korrnals/gotr/pkg/reporter"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var compareAllStages = []string{
	"cases",
	"suites",
	"sections",
	"shared steps",
	"runs",
	"plans",
	"milestones",
	"datasets",
	"groups",
	"labels",
	"templates",
	"configurations",
}

func printCompareAllStageProgress(w io.Writer, current string) {
	if w == nil {
		w = os.Stderr
	}

	ui.Section(w, "Compare all stages")
	activeIndex := -1
	for i, stage := range compareAllStages {
		if stage == current {
			activeIndex = i
			break
		}
	}

	for i, stage := range compareAllStages {
		switch {
		case i < activeIndex:
			ui.Stat(w, "✅", stage, "done")
		case i == activeIndex:
			ui.Stat(w, "⏳", stage, "active")
		default:
			ui.Stat(w, "•", stage, "pending")
		}
	}
}

func isContextCancellationError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "context canceled") || strings.Contains(msg, "deadline exceeded")
}

func isUnsupportedEndpointError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "unknown method") {
		return false
	}
	return strings.Contains(msg, "404") || strings.Contains(msg, "file not found")
}

func splitErrorsBySupport(errors map[string]error) (unsupported []string, regular []string) {
	unsupported = make([]string, 0)
	regular = make([]string, 0)
	for resource, err := range errors {
		if isUnsupportedEndpointError(err) {
			unsupported = append(unsupported, resource)
			continue
		}
		regular = append(regular, resource)
	}
	sort.Strings(unsupported)
	sort.Strings(regular)
	return unsupported, regular
}

func interruptedResult(resource string, pid1, pid2 int64) *CompareResult {
	return &CompareResult{
		Resource:     resource,
		Project1ID:   pid1,
		Project2ID:   pid2,
		Status:       CompareStatusInterrupted,
		OnlyInFirst:  []ItemInfo{},
		OnlyInSecond: []ItemInfo{},
		Common:       []CommonItemInfo{},
	}
}

func partialResult(resource string, pid1, pid2 int64) *CompareResult {
	return &CompareResult{
		Resource:     resource,
		Project1ID:   pid1,
		Project2ID:   pid2,
		Status:       CompareStatusPartial,
		OnlyInFirst:  []ItemInfo{},
		OnlyInSecond: []ItemInfo{},
		Common:       []CommonItemInfo{},
	}
}

func fillResourcePartialResult(result *allResult, resource string, pid1, pid2 int64) {
	if result == nil {
		return
	}

	switch resource {
	case "cases":
		if result.Cases == nil {
			result.Cases = partialResult("cases", pid1, pid2)
		}
	case "suites":
		if result.Suites == nil {
			result.Suites = partialResult("suites", pid1, pid2)
		}
	case "sections":
		if result.Sections == nil {
			result.Sections = partialResult("sections", pid1, pid2)
		}
	case "shared_steps":
		if result.SharedSteps == nil {
			result.SharedSteps = partialResult("sharedsteps", pid1, pid2)
		}
	case "runs":
		if result.Runs == nil {
			result.Runs = partialResult("runs", pid1, pid2)
		}
	case "plans":
		if result.Plans == nil {
			result.Plans = partialResult("plans", pid1, pid2)
		}
	case "milestones":
		if result.Milestones == nil {
			result.Milestones = partialResult("milestones", pid1, pid2)
		}
	case "datasets":
		if result.Datasets == nil {
			result.Datasets = partialResult("datasets", pid1, pid2)
		}
	case "groups":
		if result.Groups == nil {
			result.Groups = partialResult("groups", pid1, pid2)
		}
	case "labels":
		if result.Labels == nil {
			result.Labels = partialResult("labels", pid1, pid2)
		}
	case "templates":
		if result.Templates == nil {
			result.Templates = partialResult("templates", pid1, pid2)
		}
	case "configurations":
		if result.Configurations == nil {
			result.Configurations = partialResult("configurations", pid1, pid2)
		}
	}
}

func fillInterruptedResults(result *allResult, pid1, pid2 int64) {
	if result.Cases == nil {
		result.Cases = interruptedResult("cases", pid1, pid2)
	}
	if result.Suites == nil {
		result.Suites = interruptedResult("suites", pid1, pid2)
	}
	if result.Sections == nil {
		result.Sections = interruptedResult("sections", pid1, pid2)
	}
	if result.SharedSteps == nil {
		result.SharedSteps = interruptedResult("sharedsteps", pid1, pid2)
	}
	if result.Runs == nil {
		result.Runs = interruptedResult("runs", pid1, pid2)
	}
	if result.Plans == nil {
		result.Plans = interruptedResult("plans", pid1, pid2)
	}
	if result.Milestones == nil {
		result.Milestones = interruptedResult("milestones", pid1, pid2)
	}
	if result.Datasets == nil {
		result.Datasets = interruptedResult("datasets", pid1, pid2)
	}
	if result.Groups == nil {
		result.Groups = interruptedResult("groups", pid1, pid2)
	}
	if result.Labels == nil {
		result.Labels = interruptedResult("labels", pid1, pid2)
	}
	if result.Templates == nil {
		result.Templates = interruptedResult("templates", pid1, pid2)
	}
	if result.Configurations == nil {
		result.Configurations = interruptedResult("configurations", pid1, pid2)
	}
}

type allResultMeta struct {
	ExecutionStatus      CompareStatus `json:"execution_status" yaml:"execution_status"`
	Interrupted          bool          `json:"interrupted" yaml:"interrupted"`
	Elapsed              string        `json:"elapsed" yaml:"elapsed"`
	ElapsedMs            int64         `json:"elapsed_ms" yaml:"elapsed_ms"`
	ErrorCount           int           `json:"error_summary_count" yaml:"error_summary_count"`
	ErrorResources       []string      `json:"error_resources,omitempty" yaml:"error_resources,omitempty"`
	UnsupportedCount     int           `json:"unsupported_summary_count,omitempty" yaml:"unsupported_summary_count,omitempty"`
	UnsupportedResources []string      `json:"unsupported_resources,omitempty" yaml:"unsupported_resources,omitempty"`
	GeneratedAt          string        `json:"generated_at" yaml:"generated_at"`
}

func compareResultStatus(res *CompareResult) CompareStatus {
	if res == nil || res.Status == "" {
		return CompareStatusInterrupted
	}
	return res.Status
}

func deriveAllExecutionStatus(result *allResult, interrupted bool, errors map[string]error) CompareStatus {
	if interrupted {
		return CompareStatusInterrupted
	}
	if len(errors) > 0 {
		return CompareStatusPartial
	}

	statuses := []CompareStatus{
		compareResultStatus(result.Cases),
		compareResultStatus(result.Suites),
		compareResultStatus(result.Sections),
		compareResultStatus(result.SharedSteps),
		compareResultStatus(result.Runs),
		compareResultStatus(result.Plans),
		compareResultStatus(result.Milestones),
		compareResultStatus(result.Datasets),
		compareResultStatus(result.Groups),
		compareResultStatus(result.Labels),
		compareResultStatus(result.Templates),
		compareResultStatus(result.Configurations),
	}

	hasPartial := false
	for _, status := range statuses {
		switch status {
		case CompareStatusInterrupted:
			return CompareStatusInterrupted
		case CompareStatusPartial:
			hasPartial = true
		}
	}
	if hasPartial {
		return CompareStatusPartial
	}
	return CompareStatusComplete
}

func buildAllMeta(result *allResult, interrupted bool, errors map[string]error, elapsed time.Duration) allResultMeta {
	unsupportedResources, errorResources := splitErrorsBySupport(errors)

	return allResultMeta{
		ExecutionStatus:      deriveAllExecutionStatus(result, interrupted, errors),
		Interrupted:          interrupted,
		Elapsed:              elapsed.Round(time.Millisecond).String(),
		ElapsedMs:            elapsed.Milliseconds(),
		ErrorCount:           len(errorResources),
		ErrorResources:       errorResources,
		UnsupportedCount:     len(unsupportedResources),
		UnsupportedResources: unsupportedResources,
		GeneratedAt:          time.Now().UTC().Format(time.RFC3339),
	}
}

// allResult represents the combined results of comparing all resources.
type allResult struct {
	Meta           allResultMeta  `json:"meta" yaml:"meta"`
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
		Short: "Compare all resources between two projects",
		Long: `Compares all supported resources between two projects.

Compared resources:
- cases
- suites
- sections
- sharedsteps (shared steps)
- runs (test runs)
- plans (test plans)
- milestones (milestones)
- datasets (datasets)
- groups
- labels
- templates
- configurations

Examples:
	# Compare all resources
  gotr compare all --pid1 30 --pid2 31

	# Save result to the default file
  gotr compare all --pid1 30 --pid2 31 --save

	# Save result to a specific file
  gotr compare all --pid1 30 --pid2 31 --save-to result.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClientSafe(cmd)
			ctx := cmd.Context()
			quiet, _ := cmd.Flags().GetBool("quiet")
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			// Parse flags
			pid1, pid2, format, savePath, err := parseCommonFlags(cmd, cli)
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
			interrupted := false
			preloadedSuites, suitesPreloadErr := cli.GetSuitesParallel(ctx, []int64{pid1, pid2}, 2, nil)
			if suitesPreloadErr != nil {
				preloadedSuites = nil
			}

			announce := func(resource string) {
				if !quiet {
					printCompareAllStageProgress(os.Stderr, resource)
					ui.Infof(os.Stderr, "Comparing %s...", resource)
				}
			}

			recordErr := func(resource string, err error) {
				if err == nil {
					return
				}
				if isContextCancellationError(err) || ctx.Err() != nil {
					interrupted = true
					return
				}
				errors[resource] = err
				fillResourcePartialResult(result, resource, pid1, pid2)
			}

			// Cases
			announce("cases")
			if casesResult, _, err := compareCasesInternal(ctx, cmd, cli, pid1, pid2, "title", preloadedSuites); err == nil {
				result.Cases = casesResult
			} else {
				recordErr("cases", err)
			}
			if interrupted {
				goto done
			}

			// Suites
			announce("suites")
			if suitesResult, err := compareSuitesInternalWithSuites(ctx, cli, pid1, pid2, quiet, preloadedSuites); err == nil {
				result.Suites = suitesResult
			} else {
				recordErr("suites", err)
			}
			if interrupted {
				goto done
			}

			// Sections
			announce("sections")
			if sectionsResult, err := compareSectionsInternalWithSuites(ctx, cmd, cli, pid1, pid2, quiet, preloadedSuites); err == nil {
				result.Sections = sectionsResult
			} else {
				recordErr("sections", err)
			}
			if interrupted {
				goto done
			}

			// Shared Steps
			announce("shared steps")
			if sharedStepsResult, err := compareSimpleInternal(ctx, cli, pid1, pid2, "sharedsteps", fetchSharedStepItems, quiet); err == nil {
				result.SharedSteps = sharedStepsResult
			} else {
				recordErr("shared_steps", err)
			}
			if interrupted {
				goto done
			}

			// Runs
			announce("runs")
			if runsResult, err := compareSimpleInternal(ctx, cli, pid1, pid2, "runs", fetchRunItems, quiet); err == nil {
				result.Runs = runsResult
			} else {
				recordErr("runs", err)
			}
			if interrupted {
				goto done
			}

			// Plans
			announce("plans")
			if plansResult, err := compareSimpleInternal(ctx, cli, pid1, pid2, "plans", fetchPlanItems, quiet); err == nil {
				result.Plans = plansResult
			} else {
				recordErr("plans", err)
			}
			if interrupted {
				goto done
			}

			// Milestones
			announce("milestones")
			if milestonesResult, err := compareSimpleInternal(ctx, cli, pid1, pid2, "milestones", fetchMilestoneItems, quiet); err == nil {
				result.Milestones = milestonesResult
			} else {
				recordErr("milestones", err)
			}
			if interrupted {
				goto done
			}

			// Datasets
			announce("datasets")
			if datasetsResult, err := compareSimpleInternal(ctx, cli, pid1, pid2, "datasets", fetchDatasetItems, quiet); err == nil {
				result.Datasets = datasetsResult
			} else {
				recordErr("datasets", err)
			}
			if interrupted {
				goto done
			}

			// Groups
			announce("groups")
			if groupsResult, err := compareSimpleInternal(ctx, cli, pid1, pid2, "groups", fetchGroupItems, quiet); err == nil {
				result.Groups = groupsResult
			} else {
				recordErr("groups", err)
			}
			if interrupted {
				goto done
			}

			// Labels
			announce("labels")
			if labelsResult, err := compareSimpleInternal(ctx, cli, pid1, pid2, "labels", fetchLabelItems, quiet); err == nil {
				result.Labels = labelsResult
			} else {
				recordErr("labels", err)
			}
			if interrupted {
				goto done
			}

			// Templates
			announce("templates")
			if templatesResult, err := compareSimpleInternal(ctx, cli, pid1, pid2, "templates", fetchTemplateItems, quiet); err == nil {
				result.Templates = templatesResult
			} else {
				recordErr("templates", err)
			}
			if interrupted {
				goto done
			}

			// Configurations
			announce("configurations")
			if configsResult, err := compareSimpleInternal(ctx, cli, pid1, pid2, "configurations", fetchConfigurationItems, quiet); err == nil {
				result.Configurations = configsResult
			} else {
				recordErr("configurations", err)
			}

		done:
			if interrupted {
				fillInterruptedResults(result, pid1, pid2)
				errors["execution"] = fmt.Errorf("interrupted by user; remaining resources were not processed")
			}

			// Print summary table
			elapsed := time.Since(startTime)
			result.Meta = buildAllMeta(result, interrupted, errors, elapsed)
			if !quiet {
				printAllSummaryTable(project1Name, pid1, project2Name, pid2, result, errors, elapsed)
			}

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
					if !quiet {
						fmt.Println()
						ui.Infof(os.Stdout, "Result saved to %s", savePath)
					}
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

// allCmd is the exported command.
var allCmd = newAllCmd()

// parseCommonFlags parses common flags for all subcommands.
// When pid1 or pid2 is not provided, falls back to interactive project selection.
// Returns savePath which can be:
//   - "__DEFAULT__" if --save flag was used (save to default location)
//   - custom path if --save-to flag was used
//   - "" if neither flag was used
func parseCommonFlags(cmd *cobra.Command, cli client.ClientInterface) (pid1, pid2 int64, format, savePath string, err error) {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)
	usedInteractivePID := false

	pid1Str, _ := cmd.Flags().GetString("pid1")
	pid1, _ = flags.ParseID(pid1Str)
	if pid1 <= 0 {
		pid1, err = interactive.SelectProject(ctx, p, cli, "Select first project (pid1):")
		if err != nil {
			return 0, 0, "", "", fmt.Errorf("pid1 not specified and interactive selection failed: %w", err)
		}
		usedInteractivePID = true
	}

	pid2Str, _ := cmd.Flags().GetString("pid2")
	pid2, _ = flags.ParseID(pid2Str)
	if pid2 <= 0 {
		pid2, err = interactive.SelectProject(ctx, p, cli, "Select second project (pid2):")
		if err != nil {
			return 0, 0, "", "", fmt.Errorf("pid2 not specified and interactive selection failed: %w", err)
		}
		usedInteractivePID = true
	}

	format, _ = cmd.Flags().GetString("format")
	if format == "" {
		format = "table"
	}

	// Check save flags first, then interactive fallback if pid selection was interactive.
	savePath, explicitSaveChoice, err := outpututils.ResolveSavePathFromFlags(cmd)
	if err != nil {
		return 0, 0, "", "", err
	}

	if !explicitSaveChoice && usedInteractivePID {
		savePath, err = outpututils.PromptSavePath(p, "compare result")
		if err != nil {
			return 0, 0, "", "", err
		}
	}

	return pid1, pid2, format, savePath, nil
}

// addCommonFlags configures common flag defaults for compare subcommands.
// Note: pid1, pid2, format, save, save-to are registered as persistent flags
// in register.go to ensure they appear in completion.
// pid1 and pid2 are not marked required — interactive fallback handles missing values.
func addCommonFlags(cmd *cobra.Command) {
	_ = cmd
	// No MarkFlagRequired — interactive selection handles missing pid1/pid2.
}

// printAllSummaryTable prints a formatted table summary for compare all
// using go-pretty tables and reporter for consistent aligned output.
func printAllSummaryTable(project1Name string, pid1 int64, project2Name string, pid2 int64, result *allResult, errors map[string]error, elapsed time.Duration) {
	// Header via reporter
	rpt := reporter.New("Project comparison").
		Section("Projects").
		StatFmt("📋", "Project 1", "%s (ID: %d)", project1Name, pid1).
		StatFmt("📋", "Project 2", "%s (ID: %d)", project2Name, pid2)
	rpt.Print()

	// Resource comparison table via go-pretty
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(table.StyleRounded)
	tw.SetTitle("RESOURCE SUMMARY")
	tw.Style().Title.Align = text.AlignCenter

	tw.AppendHeader(table.Row{"Resource", "Only in P1", "Only in P2", "Common", "Data status"})
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignLeft, WidthMin: 16},
		{Number: 2, Align: text.AlignRight},
		{Number: 3, Align: text.AlignRight},
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignCenter},
	})

	appendResourceRow(tw, "Cases", result.Cases, errors["cases"])
	appendResourceRow(tw, "Suites", result.Suites, errors["suites"])
	appendResourceRow(tw, "Sections", result.Sections, errors["sections"])
	appendResourceRow(tw, "Shared Steps", result.SharedSteps, errors["shared_steps"])
	appendResourceRow(tw, "Runs", result.Runs, errors["runs"])
	appendResourceRow(tw, "Plans", result.Plans, errors["plans"])
	appendResourceRow(tw, "Milestones", result.Milestones, errors["milestones"])
	appendResourceRow(tw, "Datasets", result.Datasets, errors["datasets"])
	appendResourceRow(tw, "Groups", result.Groups, errors["groups"])
	appendResourceRow(tw, "Labels", result.Labels, errors["labels"])
	appendResourceRow(tw, "Templates", result.Templates, errors["templates"])
	appendResourceRow(tw, "Configurations", result.Configurations, errors["configurations"])

	fmt.Println()
	tw.Render()
	fmt.Println()

	// Footer stats via reporter
	footer := reporter.New("Totals").
		Section("Timing").
		Stat("⏱️", "Execution time", elapsed.Round(time.Second))

	if len(errors) > 0 {
		unsupportedResources, regularResources := splitErrorsBySupport(errors)

		if len(unsupportedResources) > 0 {
			footer.Section("Unsupported endpoints")
			for _, resource := range unsupportedResources {
				footer.Stat("ℹ️", resource, "not supported by server API")
			}
		}

		if len(regularResources) > 0 {
			footer.Section("Errors")
			for _, resource := range regularResources {
				footer.Stat("❌", resource, errors[resource])
			}
		}
	}

	footer.Print()
}

// appendResourceRow adds a resource row to the go-pretty table.
func appendResourceRow(tw table.Writer, name string, result *CompareResult, resourceErr error) {
	if result == nil {
		status := "INTERRUPTED"
		if isUnsupportedEndpointError(resourceErr) {
			status = reporter.Yellow("UNSUPPORTED")
		}
		tw.AppendRow(table.Row{name, "-", "-", "-", status})
		return
	}

	onlyP1 := len(result.OnlyInFirst)
	onlyP2 := len(result.OnlyInSecond)
	common := len(result.Common)

	status := "COMPLETE"
	switch result.Status {
	case CompareStatusInterrupted:
		status = reporter.Red("INTERRUPTED")
	case CompareStatusPartial:
		if isUnsupportedEndpointError(resourceErr) {
			status = reporter.Yellow("UNSUPPORTED")
		} else {
			status = reporter.Yellow("PARTIAL")
		}
	case CompareStatusComplete:
		status = reporter.Green("COMPLETE")
	default:
		status = reporter.Yellow("UNKNOWN")
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

	quiet, _ := cmd.Flags().GetBool("quiet")
	if !quiet {
		fmt.Println()
		ui.Infof(os.Stdout, "Result saved to %s", filePath)
	}
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
