// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/Korrnals/gotr/internal/cleanup"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// detailedEntityTypes is the column set of the per-project ×
// per-entity-type post-flight summary. Using the same canonical order
// across runs keeps tables visually comparable.
var detailedEntityTypes = []string{"case", "run", "plan", "plan_entry", "result", "test"}

// printCleanupSummaryDetailed prints a per-project × per-entity-type
// table after the executor has finished. It is a noop on empty plans.
func printCleanupSummaryDetailed(cmd *cobra.Command, plan *cleanup.Plan) {
	if plan == nil || plan.TotalCount == 0 {
		return
	}
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "\nBreakdown by project × entity type:")

	t := ui.NewTable(cmd)
	header := table.Row{"PROJECT"}
	for _, e := range detailedEntityTypes {
		header = append(header, e)
	}
	header = append(header, "TOTAL", "SIZE")
	t.AppendHeader(header)

	totals := make(map[string]int, len(detailedEntityTypes))
	var grandCount int
	var grandBytes int64

	for _, sel := range plan.Projects {
		counts := make(map[string]int, len(detailedEntityTypes))
		var projTotal int
		var projBytes int64
		for _, a := range sel.Attachments {
			kind := a.InferredEntityType()
			counts[kind]++
			projTotal++
			projBytes += a.Size
		}
		row := table.Row{fmt.Sprintf("%s (%d)", truncate(sel.ProjectName, 30), sel.ProjectID)}
		for _, e := range detailedEntityTypes {
			row = append(row, counts[e])
			totals[e] += counts[e]
		}
		row = append(row, projTotal, ui.HumanBytes(projBytes))
		t.AppendRow(row)
		grandCount += projTotal
		grandBytes += projBytes
	}

	footer := table.Row{"Total"}
	for _, e := range detailedEntityTypes {
		footer = append(footer, totals[e])
	}
	footer = append(footer, grandCount, ui.HumanBytes(grandBytes))
	t.AppendFooter(footer)

	ui.Table(cmd, t)
}

// printCleanupFinalBlock prints the post-execution summary: run-id,
// audit report paths, snapshot path and next-step hints.
func printCleanupFinalBlock(cmd *cobra.Command, plan *cleanup.Plan, res *cleanup.ExecuteResult, opts *cleanupOptions, runID string, reportPaths []string) {
	if plan == nil || plan.TotalCount == 0 {
		return
	}
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "\nFinal summary:")

	if runID != "" {
		fmt.Fprintf(out, "  Run id           : %s\n", runID)
	}
	printReportPaths(out, reportPaths)
	printSnapshotInfo(out, res, opts)
	printNextSteps(out, res, opts, runID)
}

func printReportPaths(out io.Writer, paths []string) {
	if len(paths) == 0 {
		return
	}
	fmt.Fprintln(out, "  Audit report     :")
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	for _, p := range sorted {
		abs, err := filepath.Abs(p)
		if err != nil {
			abs = p
		}
		fmt.Fprintf(out, "    - %s\n", abs)
	}
}

func printSnapshotInfo(out io.Writer, res *cleanup.ExecuteResult, opts *cleanupOptions) {
	if opts.SkipSnapshot || opts.DryRun || res == nil || res.SnapshotID == "" {
		return
	}
	fmt.Fprintf(out, "  Snapshot id      : %s\n", res.SnapshotID)
	store, err := snap.NewStore()
	if err != nil {
		return
	}
	abs, err := filepath.Abs(store.SnapDir(res.SnapshotID))
	if err != nil {
		abs = store.SnapDir(res.SnapshotID)
	}
	fmt.Fprintf(out, "  Snapshot path    : %s\n", abs)
}

func printNextSteps(out io.Writer, res *cleanup.ExecuteResult, opts *cleanupOptions, runID string) {
	if opts.DryRun {
		return
	}
	hasFailures := res != nil && res.DeleteErrors > 0
	hasSnapshot := res != nil && res.SnapshotID != ""
	if !hasSnapshot && !hasFailures {
		return
	}
	fmt.Fprintln(out, "\nNext steps:")
	if hasSnapshot {
		fmt.Fprintf(out, "  - Restore the deleted attachments:  gotr snap rollback %s\n", res.SnapshotID)
	}
	if runID != "" && hasFailures {
		fmt.Fprintf(out, "  - Resume failed/timed-out projects: gotr attachments cleanup --resume %s\n", runID)
	}
}
