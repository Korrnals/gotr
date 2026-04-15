package compare

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// Post-action option keys.
const (
	actionExit     = "exit"
	actionBack     = "back"
	actionSave     = "save"
	actionSync     = "sync"
	actionSnap     = "snap"
	actionDrillRes = "drill_resource"
)

// comparePostAction shows a post-action menu after a compare result.
// It loops until the user exits or chooses back (for re-compare).
// Returns the selected action key. Non-interactive → returns immediately.
func comparePostAction(ctx context.Context, cmd *cobra.Command, result CompareResult, p1Name, p2Name string) string {
	if !interactive.HasPrompterInContext(ctx) || interactive.IsNonInteractive(ctx) {
		return actionExit
	}
	p := interactive.PrompterFromContext(ctx)

	hasDifferences := len(result.OnlyInFirst) > 0 || len(result.OnlyInSecond) > 0

	options := []interactive.ActionOption{
		{Label: interactive.OptExit, Key: actionExit},
		{Label: "💾 Save results to file", Key: actionSave},
		{Label: "→ Sync: migrate differences", Key: actionSync, Disabled: !hasDifferences, Hint: "no differences found"},
	}

	key, err := interactive.ActionMenu(ctx, p, "Comparison complete. What next?", options)
	if err != nil {
		return actionExit
	}

	switch key {
	case actionSave:
		if err := promptAndSave(ctx, p, cmd, result, p1Name, p2Name); err != nil {
			ui.Error(os.Stdout, fmt.Sprintf("Save failed: %v", err))
		}
		// After save, show menu again.
		return comparePostAction(ctx, cmd, result, p1Name, p2Name)
	default:
		return key
	}
}

// compareAllPostAction shows a post-action menu for compare-all with drill-down.
// resources maps resource display names to their CompareResult.
func compareAllPostAction(ctx context.Context, cmd *cobra.Command, result *allResult, p1Name, p2Name string, pid1, pid2 int64) string {
	if !interactive.HasPrompterInContext(ctx) || interactive.IsNonInteractive(ctx) {
		return actionExit
	}
	p := interactive.PrompterFromContext(ctx)

	options := []interactive.ActionOption{
		{Label: interactive.OptExit, Key: actionExit},
		{Label: "🔍 Drill-down: view resource details", Key: actionDrillRes},
		{Label: "💾 Save results to file", Key: actionSave},
	}

	key, err := interactive.ActionMenu(ctx, p, "Compare all complete. What next?", options)
	if err != nil {
		return actionExit
	}

	switch key {
	case actionDrillRes:
		drillDownResource(ctx, p, result, p1Name, p2Name)
		return compareAllPostAction(ctx, cmd, result, p1Name, p2Name, pid1, pid2)
	case actionSave:
		// Save is handled by caller (save flags).
		return actionSave
	default:
		return key
	}
}

// drillDownResource lets the user pick a resource and view its detailed table (paged).
func drillDownResource(ctx context.Context, p interactive.Prompter, result *allResult, p1Name, p2Name string) {
	resources := collectDrillDownResources(result)
	if len(resources) == 0 {
		ui.Infof(os.Stdout, "No resource data available for drill-down.")
		return
	}

	labels := make([]string, len(resources))
	for i, r := range resources {
		labels[i] = fmt.Sprintf("%-16s  P1: %d unique │ P2: %d unique │ Common: %d",
			r.name,
			len(r.result.OnlyInFirst),
			len(r.result.OnlyInSecond),
			len(r.result.Common))
	}

	for {
		idx, err := interactive.Browse(ctx, p, interactive.BrowseConfig{
			Prompt:    "Select resource to view details:",
			Items:     labels,
			AllowBack: true,
		})
		if err != nil {
			return // back or exit
		}

		r := resources[idx]
		lines := renderTableLines(r.result, p1Name, p2Name)
		if interactive.ShouldPage(len(lines)) {
			_ = interactive.Pager(interactive.PagerConfig{
				Lines:  lines,
				Header: fmt.Sprintf("=== %s (projects %d ↔ %d) ===", r.result.Resource, r.result.Project1ID, r.result.Project2ID),
			})
		} else {
			for _, line := range lines {
				fmt.Println(line)
			}
		}
	}
}

type drillDownEntry struct {
	name   string
	result CompareResult
}

func collectDrillDownResources(result *allResult) []drillDownEntry {
	type named struct {
		name string
		ptr  *CompareResult
	}
	all := []named{
		{"Cases", result.Cases},
		{"Suites", result.Suites},
		{"Sections", result.Sections},
		{"Shared Steps", result.SharedSteps},
		{"Runs", result.Runs},
		{"Plans", result.Plans},
		{"Milestones", result.Milestones},
		{"Datasets", result.Datasets},
		{"Groups", result.Groups},
		{"Labels", result.Labels},
		{"Templates", result.Templates},
		{"Configurations", result.Configurations},
	}

	entries := make([]drillDownEntry, 0, len(all))
	for _, r := range all {
		if r.ptr != nil && r.ptr.Status == CompareStatusComplete {
			entries = append(entries, drillDownEntry{name: r.name, result: *r.ptr})
		}
	}
	return entries
}

// promptAndSave prompts the user for format and saves the result.
func promptAndSave(ctx context.Context, p interactive.Prompter, cmd *cobra.Command, result CompareResult, p1Name, p2Name string) error {
	formats := []string{"json", "yaml", "csv", "table (text)"}
	idx, _, err := p.Select("Save format:", formats)
	if err != nil {
		return err
	}

	format := []string{"json", "yaml", "csv", "table"}[idx]

	input, err := p.Input("File path (or empty for default):", "")
	if err != nil {
		return err
	}
	savePath := strings.TrimSpace(input)
	if savePath == "" {
		savePath = "__DEFAULT__"
	}

	return PrintCompareResult(cmd, result, p1Name, p2Name, format, savePath)
}

// renderTableLines builds the compare table as a slice of strings (one per line)
// suitable for paging or saving. Does not print to stdout.
func renderTableLines(result CompareResult, p1Name, p2Name string) []string {
	var lines []string

	lines = append(lines, fmt.Sprintf("\n=== Comparison: %s (projects %d ↔ %d) ===\n", result.Resource, result.Project1ID, result.Project2ID))
	lines = append(lines, renderOnlyInProjectLines(result.OnlyInFirst, result.Project1ID, p1Name)...)
	lines = append(lines, renderOnlyInProjectLines(result.OnlyInSecond, result.Project2ID, p2Name)...)
	lines = append(lines, renderCommonLines(result.Common, result.Project1ID, result.Project2ID)...)
	lines = append(lines, renderIDMappingLines(result.Common)...)

	return lines
}

// --- Line rendering helpers (mirror the existing print* functions but return []string) ---

func renderOnlyInProjectLines(items []ItemInfo, projectID int64, projectName string) []string {
	idWidth := 8
	nameWidth := 70
	widths := []int{idWidth, nameWidth}
	totalInnerWidth := idWidth + nameWidth + 3*len(widths) - 1

	title := fmt.Sprintf("Only in project %d - %q", projectID, projectName)
	var lines []string
	lines = append(lines, hBorder("┌", "┬", "┐", widths))
	lines = append(lines, headerLine(title, totalInnerWidth))

	if len(items) == 0 {
		lines = append(lines, separatorLine(widths))
		lines = append(lines, headerLine("(none)", totalInnerWidth))
		lines = append(lines, hBorder("└", "┴", "┘", widths))
		lines = append(lines, "")
		return lines
	}

	lines = append(lines, separatorLine(widths))
	lines = append(lines, rowLine([]string{"ID", "Name"}, widths))
	lines = append(lines, separatorLine(widths))

	for _, item := range items {
		lines = append(lines, rowLine([]string{fmt.Sprintf("%d", item.ID), item.Name}, widths))
	}

	lines = append(lines, hBorder("└", "┴", "┘", widths))
	lines = append(lines, "")
	return lines
}

func renderCommonLines(items []CommonItemInfo, pid1, pid2 int64) []string {
	nameWidth := 50
	id1Width := 12
	id2Width := 12
	statusWidth := 20
	widths := []int{nameWidth, id1Width, id2Width, statusWidth}
	totalInnerWidth := nameWidth + id1Width + id2Width + statusWidth + 3*len(widths) - 1

	var lines []string
	lines = append(lines, hBorder("┌", "┬", "┐", widths))
	lines = append(lines, headerLine("Common in both projects", totalInnerWidth))

	if len(items) == 0 {
		lines = append(lines, separatorLine(widths))
		lines = append(lines, headerLine("(none)", totalInnerWidth))
		lines = append(lines, hBorder("└", "┴", "┘", widths))
		lines = append(lines, "")
		return lines
	}

	lines = append(lines, separatorLine(widths))
	lines = append(lines, rowLine([]string{
		"Name",
		fmt.Sprintf("ID proj %d", pid1),
		fmt.Sprintf("ID proj %d", pid2),
		"ID status",
	}, widths))
	lines = append(lines, separatorLine(widths))

	for _, item := range items {
		status := "✓ Match"
		if !item.IDsMatch {
			status = "⚠ Differ"
		}
		lines = append(lines, rowLine([]string{
			item.Name,
			fmt.Sprintf("%d", item.ID1),
			fmt.Sprintf("%d", item.ID2),
			status,
		}, widths))
	}

	lines = append(lines, hBorder("└", "┴", "┘", widths))
	lines = append(lines, "")
	return lines
}

func renderIDMappingLines(items []CommonItemInfo) []string {
	var mappings []CommonItemInfo
	for _, item := range items {
		if !item.IDsMatch {
			mappings = append(mappings, item)
		}
	}

	sourceWidth := 12
	targetWidth := 12
	nameWidth := 70
	widths := []int{sourceWidth, targetWidth, nameWidth}
	totalInnerWidth := sourceWidth + targetWidth + nameWidth + 3*len(widths) - 1

	var lines []string
	lines = append(lines, hBorder("┌", "┬", "┐", widths))
	lines = append(lines, headerLine("ID mapping (for updates)", totalInnerWidth))

	if len(mappings) == 0 {
		lines = append(lines, separatorLine(widths))
		lines = append(lines, headerLine("(all IDs match)", totalInnerWidth))
		lines = append(lines, hBorder("└", "┴", "┘", widths))
		lines = append(lines, "")
		return lines
	}

	lines = append(lines, separatorLine(widths))
	lines = append(lines, rowLine([]string{"Source ID", "Target ID", "Name"}, widths))
	lines = append(lines, separatorLine(widths))

	for _, item := range mappings {
		lines = append(lines, rowLine([]string{
			fmt.Sprintf("%d", item.ID1),
			fmt.Sprintf("%d", item.ID2),
			item.Name,
		}, widths))
	}

	lines = append(lines, hBorder("└", "┴", "┘", widths))
	lines = append(lines, "")
	return lines
}

// --- String-based table rendering helpers ---

func hBorder(left, mid, right string, widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("─", w+2)
	}
	return left + strings.Join(parts, mid) + right
}

func headerLine(title string, totalWidth int) string {
	titleWidth := len([]rune(title))
	padding := totalWidth - 2 - titleWidth
	if padding < 0 {
		padding = 0
		title = truncateString(title, totalWidth-2)
	}
	leftPad := padding / 2
	rightPad := padding - leftPad
	return "│" + strings.Repeat(" ", leftPad) + title + strings.Repeat(" ", rightPad) + "│"
}

func separatorLine(widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("─", w+2)
	}
	return "├" + strings.Join(parts, "┼") + "┤"
}

func rowLine(cells []string, widths []int) string {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		parts[i] = " " + padRight(truncateString(cell, widths[i]), widths[i]) + " "
	}
	return "│" + strings.Join(parts, "│") + "│"
}
