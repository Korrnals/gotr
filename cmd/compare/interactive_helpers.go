package compare

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Korrnals/gotr/internal/interactive"
	outpututils "github.com/Korrnals/gotr/internal/output"
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

	// Persist project IDs for downstream commands (session inheritance).
	if s := interactive.SessionFromContext(ctx); s != nil {
		s.SetProjects(result.Project1ID, result.Project2ID)
	}

	hasDifferences := len(result.OnlyInFirst) > 0 || len(result.OnlyInSecond) > 0
	hasData := hasDifferences || len(result.Common) > 0

	options := []interactive.ActionOption{
		{Label: interactive.OptExit, Key: actionExit},
		{Label: "📋 View detailed results", Key: actionDrillRes, Disabled: !hasData, Hint: "no data"},
		{Label: "💾 Save results to file", Key: actionSave},
		{Label: "→ Sync: migrate differences", Key: actionSync, Disabled: !hasDifferences, Hint: "no differences found"},
		{Label: "📦 Snap: manage snapshots", Key: actionSnap},
	}

	key, err := interactive.ActionMenu(ctx, p, "Comparison complete. What next?", options)
	if err != nil {
		return actionExit
	}

	switch key {
	case actionDrillRes:
		lines := renderTableLines(result, p1Name, p2Name)
		if interactive.ShouldPage(len(lines)) {
			_ = interactive.Pager(interactive.PagerConfig{
				Lines:  lines,
				Header: fmt.Sprintf("=== %s (projects %d ↔ %d) ===", result.Resource, result.Project1ID, result.Project2ID),
			})
		} else {
			for _, line := range lines {
				fmt.Println(line)
			}
		}
		return comparePostAction(ctx, cmd, result, p1Name, p2Name)
	case actionSave:
		if err := promptAndSave(ctx, p, cmd, result, p1Name, p2Name); err != nil {
			ui.Error(os.Stdout, fmt.Sprintf("Save failed: %v", err))
		}
		// After save, show menu again.
		return comparePostAction(ctx, cmd, result, p1Name, p2Name)
	case actionSync:
		runSyncFromCompare(ctx, cmd, result)
		return comparePostAction(ctx, cmd, result, p1Name, p2Name)
	case actionSnap:
		runSnapFromPostAction(ctx, cmd)
		return comparePostAction(ctx, cmd, result, p1Name, p2Name)
	default:
		return key
	}
}

// syncMenuEntry describes a sync subcommand shown in the selection menu.
type syncMenuEntry struct {
	label      string // Display label for the menu.
	subcommand string // Cobra subcommand name under "sync".
}

// syncMenuEntries defines available sync modes presented to the user.
var syncMenuEntries = []syncMenuEntry{
	{"Full migration (cases + shared steps)", "full"},
	{"Suites", "suites"},
	{"Sections", "sections"},
	{"Shared steps", "shared-steps"},
}

// recommendedSyncIndex returns the index in syncMenuEntries best matching
// the compare resource, or 0 (full) when nothing matches specifically.
func recommendedSyncIndex(resource string) int {
	switch resource {
	case "suites":
		return 1
	case "sections":
		return 2
	case "shared_steps", "shared-steps":
		return 3
	default: // cases, "" or anything else → full
		return 0
	}
}

// runSyncFromCompare shows a sync-mode selection menu and invokes the
// chosen sync subcommand with --src-project/--dst-project pre-filled.
// Remaining interactive questions (suite, backup, etc.) are handled by sync.
func runSyncFromCompare(ctx context.Context, compareCmd *cobra.Command, result CompareResult) {
	if compareCmd == nil {
		ui.Infof(os.Stdout, "Sync is not available (no command context).")
		fmt.Println()
		return
	}

	p := interactive.PrompterFromContext(ctx)

	// Build the menu labels, marking the recommended option.
	recommended := recommendedSyncIndex(result.Resource)
	labels := make([]string, len(syncMenuEntries))
	for i, e := range syncMenuEntries {
		if i == recommended {
			labels[i] = e.label + " ★"
		} else {
			labels[i] = e.label
		}
	}

	fmt.Println()
	ui.Infof(os.Stdout, "Pre-filled from compare: src-project=%d, dst-project=%d",
		result.Project1ID, result.Project2ID)

	// Persist project IDs for downstream commands (session inheritance).
	if s := interactive.SessionFromContext(ctx); s != nil {
		s.SetProjects(result.Project1ID, result.Project2ID)
	}

	idx, _, err := p.Select("What do you want to migrate?", labels)
	if err != nil {
		if interactive.IsGoBack(err) || interactive.IsExit(err) {
			return
		}
		ui.Error(os.Stdout, fmt.Sprintf("Selection failed: %v", err))
		return
	}

	syncSub := syncMenuEntries[idx].subcommand

	root := compareCmd.Root()
	syncCmd, err := interactive.FindSubcommand(root, "sync", syncSub)
	if err != nil {
		fmt.Println()
		ui.Error(os.Stdout, err.Error())
		fmt.Println()
		return
	}

	fmt.Println()
	ui.Infof(os.Stdout, "Starting: gotr sync %s (src-project=%d, dst-project=%d)",
		syncSub, result.Project1ID, result.Project2ID)
	fmt.Println()

	// Pre-fill flags known from compare. Reset them after sync finishes
	// to avoid stale state on repeated invocations.
	setFlag := func(name, value string) {
		if syncCmd.Flags().Lookup(name) != nil {
			_ = syncCmd.Flags().Set(name, value)
		}
	}
	resetFlag := func(name, zero string) {
		if f := syncCmd.Flags().Lookup(name); f != nil {
			_ = syncCmd.Flags().Set(name, zero)
			f.Changed = false
		}
	}

	setFlag("src-project", fmt.Sprintf("%d", result.Project1ID))
	setFlag("dst-project", fmt.Sprintf("%d", result.Project2ID))

	defer func() {
		resetFlag("src-project", "0")
		resetFlag("dst-project", "0")
		resetFlag("src-suite", "0")
		resetFlag("dst-suite", "0")
	}()

	// Propagate context (prompter, client, cancellation).
	syncCmd.SetContext(ctx)

	// RunE handles the rest interactively (suite selection, snapshot,
	// confirmation, etc.) — the full protection chain of the sync command.
	if err := syncCmd.RunE(syncCmd, nil); err != nil {
		if interactive.IsGoBack(err) || interactive.IsExit(err) {
			return
		}
		ui.Error(os.Stdout, fmt.Sprintf("Sync failed: %v", err))
	}
	fmt.Println()
}

// runSnapFromPostAction launches the snap list command from a post-action menu.
func runSnapFromPostAction(ctx context.Context, cmd *cobra.Command) {
	fmt.Println()
	if err := interactive.RunSubcommand(ctx, cmd.Root(), "snap", "list"); err != nil {
		ui.Error(os.Stdout, err.Error())
	}
	fmt.Println()
}

// compareAllPostAction shows a post-action menu for compare-all with drill-down.
// resources maps resource display names to their CompareResult.
func compareAllPostAction(ctx context.Context, cmd *cobra.Command, result *allResult, p1Name, p2Name string, pid1, pid2 int64) string {
	if !interactive.HasPrompterInContext(ctx) || interactive.IsNonInteractive(ctx) {
		return actionExit
	}
	p := interactive.PrompterFromContext(ctx)

	// Persist project IDs for downstream commands (session inheritance).
	if s := interactive.SessionFromContext(ctx); s != nil {
		s.SetProjects(pid1, pid2)
	}

	options := []interactive.ActionOption{
		{Label: interactive.OptExit, Key: actionExit},
		{Label: "🔍 Drill-down: view resource details", Key: actionDrillRes},
		{Label: "💾 Save results to file", Key: actionSave},
		{Label: "📦 Snap: manage snapshots", Key: actionSnap},
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
	case actionSnap:
		runSnapFromPostAction(ctx, cmd)
		return compareAllPostAction(ctx, cmd, result, p1Name, p2Name, pid1, pid2)
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

	// Resolve __DEFAULT__ to a real path so we can show it to the user.
	if savePath == "" {
		ext := format
		if ext == "table" {
			ext = "txt"
		}
		exportsDir, _ := outpututils.GetExportsDir("compare")
		_ = os.MkdirAll(exportsDir, 0o755)
		savePath = exportsDir + "/" + outpututils.GenerateFilename("compare", ext)
	}

	if err := PrintCompareResult(cmd, result, p1Name, p2Name, format, savePath); err != nil {
		return err
	}

	return nil
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
	numWidth := 5  // row number "#"
	idWidth := 8   // TestRail ID
	// Adaptive name width: fill remaining terminal space.
	// Layout: │ # │ ID │ Name │  →  1 + (numW+2) + 1 + (idW+2) + 1 + (nameW+2) + 1
	termW := interactive.TerminalWidth()
	nameWidth := termW - numWidth - idWidth - 10 // 10 = borders + padding
	if nameWidth < 30 {
		nameWidth = 30
	}

	widths := []int{numWidth, idWidth, nameWidth}
	totalInnerWidth := 0
	for _, w := range widths {
		totalInnerWidth += w + 3
	}
	totalInnerWidth-- // correct for last column

	title := fmt.Sprintf("Only in project %d — %q (%d items)", projectID, projectName, len(items))
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
	lines = append(lines, rowLine([]string{"#", "ID", "Name"}, widths))
	lines = append(lines, separatorLine(widths))

	for i, item := range items {
		lines = append(lines, rowLine([]string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%d", item.ID),
			item.Name,
		}, widths))
	}

	lines = append(lines, hBorder("└", "┴", "┘", widths))
	lines = append(lines, "")
	return lines
}

func renderCommonLines(items []CommonItemInfo, pid1, pid2 int64) []string {
	numWidth := 5
	id1Width := 10
	id2Width := 10
	statusWidth := 10
	fixedCols := numWidth + id1Width + id2Width + statusWidth
	// Adaptive name: terminal width minus fixed columns minus borders/padding.
	// 5 columns → 5*3-1 = 14 for borders+padding
	termW := interactive.TerminalWidth()
	nameWidth := termW - fixedCols - 16
	if nameWidth < 30 {
		nameWidth = 30
	}
	widths := []int{numWidth, nameWidth, id1Width, id2Width, statusWidth}
	totalInnerWidth := 0
	for _, w := range widths {
		totalInnerWidth += w + 3
	}
	totalInnerWidth--

	title := fmt.Sprintf("Common in both projects (%d items)", len(items))
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
	lines = append(lines, rowLine([]string{
		"#",
		"Name",
		fmt.Sprintf("ID P%d", pid1),
		fmt.Sprintf("ID P%d", pid2),
		"Status",
	}, widths))
	lines = append(lines, separatorLine(widths))

	for i, item := range items {
		status := "✓ Match"
		if !item.IDsMatch {
			status = "⚠ Differ"
		}
		lines = append(lines, rowLine([]string{
			fmt.Sprintf("%d", i+1),
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

	numWidth := 5
	sourceWidth := 10
	targetWidth := 10
	termW := interactive.TerminalWidth()
	nameWidth := termW - numWidth - sourceWidth - targetWidth - 13
	if nameWidth < 30 {
		nameWidth = 30
	}
	widths := []int{numWidth, sourceWidth, targetWidth, nameWidth}
	totalInnerWidth := 0
	for _, w := range widths {
		totalInnerWidth += w + 3
	}
	totalInnerWidth--

	title := fmt.Sprintf("ID mapping — for updates (%d items)", len(mappings))
	var lines []string
	lines = append(lines, hBorder("┌", "┬", "┐", widths))
	lines = append(lines, headerLine(title, totalInnerWidth))

	if len(mappings) == 0 {
		lines = append(lines, separatorLine(widths))
		lines = append(lines, headerLine("(all IDs match)", totalInnerWidth))
		lines = append(lines, hBorder("└", "┴", "┘", widths))
		lines = append(lines, "")
		return lines
	}

	lines = append(lines, separatorLine(widths))
	lines = append(lines, rowLine([]string{"#", "Source ID", "Target ID", "Name"}, widths))
	lines = append(lines, separatorLine(widths))

	for i, item := range mappings {
		lines = append(lines, rowLine([]string{
			fmt.Sprintf("%d", i+1),
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
