// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// RenderMarkdown produces the human-readable Markdown rendition of the
// deletion report. Layout is deterministic so golden tests stay stable.
func RenderMarkdown(r *Report) string {
	var sb strings.Builder
	writeMDTitle(&sb, r)
	writeMDRun(&sb, r)
	writeMDFilters(&sb, r)
	writeMDChunking(&sb, r)
	writeMDSummary(&sb, r)
	writeMDProjects(&sb, r)
	writeMDEntityBreakdown(&sb, r)
	writeMDItems(&sb, r)
	writeMDFailures(&sb, r)
	writeMDSnapshot(&sb, r)
	writeMDReferenceLimitations(&sb, r)
	writeMDArtifacts(&sb, r)
	writeMDRollback(&sb, r)
	return sb.String()
}

// writeMDTitle emits the H1 title and the DRY-RUN warning banner when
// the report describes a dry-run invocation.
func writeMDTitle(sb *strings.Builder, r *Report) {
	title := "Attachments Cleanup Report"
	if r.DryRun {
		title += " (DRY-RUN)"
	}
	fmt.Fprintf(sb, "# %s\n\n", title)
	if r.DryRun {
		sb.WriteString("> **DRY-RUN** — no snapshot was taken and no attachments were deleted.\n\n")
	}
}

// writeMDRun emits the run header table (Report ID, Snapshot ID,
// timestamp, server, gotr version, label, user, dry-run flag and the
// reconstructed CLI invocation).
func writeMDRun(sb *strings.Builder, r *Report) {
	sb.WriteString("## Run\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("|-------|-------|\n")
	row(sb, "Report ID", code(r.ID))
	if r.RunID != "" {
		row(sb, "Run ID", code(r.RunID))
	}
	row(sb, "Snapshot ID", code(notEmpty(r.SnapshotID, "—")))
	row(sb, "Timestamp (UTC)", r.Timestamp.UTC().Format(time.RFC3339))
	row(sb, "Server", notEmpty(r.Server, "—"))
	row(sb, "gotr version", notEmpty(r.GotrVer, "—"))
	row(sb, "Label", notEmpty(r.Label, "—"))
	row(sb, "User", notEmpty(r.User, "—"))
	row(sb, "Dry-run", boolStr(r.DryRun))
	if len(r.CLIArgs) > 0 {
		row(sb, "CLI", code("gotr "+strings.Join(r.CLIArgs, " ")))
	}
	sb.WriteString("\n")
}

// writeMDFilters emits the table of selection filters that produced
// the report's scope (project IDs / all-projects, age cutoff, entity
// types, scan strategy and limit).
func writeMDFilters(sb *strings.Builder, r *Report) {
	sb.WriteString("## Filters\n\n")
	sb.WriteString("| Filter | Value |\n")
	sb.WriteString("|--------|-------|\n")
	scope := "—"
	switch {
	case r.Filters.AllProjects:
		scope = "all projects"
	case len(r.Filters.ProjectIDs) > 0:
		scope = joinInts(r.Filters.ProjectIDs)
	}
	row(sb, "Scope", scope)
	row(sb, "Older than", notEmpty(r.Filters.OlderThan, "—"))
	if r.Filters.CutoffUnix > 0 {
		row(sb, "Cutoff (UTC)", time.Unix(r.Filters.CutoffUnix, 0).UTC().Format(time.RFC3339))
	}
	row(sb, "Entity types", joinStr(r.Filters.EntityTypes))
	row(sb, "Scan strategy", notEmpty(r.Filters.ScanStrategy, "auto"))
	if r.Filters.Limit > 0 {
		row(sb, "Limit", fmt.Sprintf("%d", r.Filters.Limit))
	}
	sb.WriteString("\n")
}

// writeMDSummary emits the aggregate counts (total selected, backed
// up, bytes, deleted, failed).
func writeMDSummary(sb *strings.Builder, r *Report) {
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	row(sb, "Total selected", fmt.Sprintf("%d", r.Summary.TotalSelected))
	row(sb, "Backed up (count)", fmt.Sprintf("%d", r.Summary.BackedUp))
	row(sb, "Backed up (bytes)", humanBytes(r.Summary.BackupBytes))
	row(sb, "Deleted", fmt.Sprintf("%d", r.Summary.Deleted))
	row(sb, "Failed", fmt.Sprintf("%d", r.Summary.Failed))
	row(sb, "Freed on server", humanBytes(r.Summary.FreedBytes))
	sb.WriteString("\n")
}

// writeMDProjects emits the per-project breakdown (count, total bytes,
// oldest timestamp). Skipped when no projects are present.
func writeMDProjects(sb *strings.Builder, r *Report) {
	if len(r.Projects) == 0 {
		return
	}
	sb.WriteString("## Per-project breakdown\n\n")
	sb.WriteString("| Project | Name | Count | Bytes | Oldest |\n")
	sb.WriteString("|---------|------|-------|-------|--------|\n")
	for _, p := range r.Projects {
		oldest := "—"
		if p.OldestUnix > 0 {
			oldest = time.Unix(p.OldestUnix, 0).UTC().Format("2006-01-02")
		}
		fmt.Fprintf(sb, "| %d | %s | %d | %s | %s |\n",
			p.ProjectID, escapePipe(p.ProjectName), p.Count, humanBytes(p.TotalBytes), oldest)
	}
	sb.WriteString("\n")
}

// writeMDItems emits the full per-attachment table (id, name, size,
// parent, created date). Skipped when no project carries any items.
func writeMDItems(sb *strings.Builder, r *Report) {
	if !hasItems(r) {
		return
	}
	sb.WriteString("## Deleted attachments\n\n")
	sb.WriteString("| Project | Attachment ID | Name | Size | Parent | Created (UTC) |\n")
	sb.WriteString("|---------|---------------|------|------|--------|---------------|\n")
	for _, p := range r.Projects {
		for _, it := range p.Items {
			created := "—"
			if it.CreatedUnix > 0 {
				created = time.Unix(it.CreatedUnix, 0).UTC().Format("2006-01-02 15:04")
			}
			parent := "—"
			if it.ParentKind != "" {
				parent = it.ParentKind
				if it.ParentID != "" {
					parent = parent + ":" + it.ParentID
				}
			}
			fmt.Fprintf(sb, "| %d | %d | %s | %s | %s | %s |\n",
				p.ProjectID, it.AttachmentID, escapePipe(it.Name),
				humanBytes(it.Size), parent, created)
		}
	}
	sb.WriteString("\n")
}

// writeMDFailures emits the list of per-attachment failures observed
// during deletion. Skipped when there are none.
func writeMDFailures(sb *strings.Builder, r *Report) {
	if len(r.Failures) == 0 {
		return
	}
	sb.WriteString("## Failures\n\n")
	sb.WriteString("| Attachment ID | Project ID | Error |\n")
	sb.WriteString("|---------------|------------|-------|\n")
	for _, f := range r.Failures {
		fmt.Fprintf(sb, "| %d | %d | %s |\n",
			f.AttachmentID, f.ProjectID, escapePipe(f.Error))
	}
	sb.WriteString("\n")
}

// writeMDRollback emits the footer: snapshot location and the
// `gotr snap rollback` command for real runs, or an explanatory note
// for dry-runs / runs without a snapshot.
func writeMDRollback(sb *strings.Builder, r *Report) {
	sb.WriteString("## Rollback\n\n")
	switch {
	case r.SnapshotID != "" && !r.DryRun:
		path := "~/.gotr/snaps/cleanup-attachments/" + r.SnapshotID
		if r.Snapshot != nil && r.Snapshot.Path != "" {
			path = r.Snapshot.Path
		}
		fmt.Fprintf(sb, "Snapshot is preserved under `%s/`.\n\n", path)
		sb.WriteString("Restore with:\n\n```\ngotr snap rollback ")
		sb.WriteString(r.SnapshotID)
		sb.WriteString("\n```\n\n")
		sb.WriteString("Recommended (verifies integrity.json before re-uploading):\n\n```\ngotr snap rollback ")
		sb.WriteString(r.SnapshotID)
		sb.WriteString(" --verify-integrity\n```\n")
	case r.DryRun:
		sb.WriteString("_No snapshot was taken (dry-run)._\n")
	default:
		sb.WriteString("_No snapshot reference recorded._\n")
	}
}

// writeMDChunking emits the chunked-execution & concurrency profile
// of this run. Skipped when no chunking metadata is attached.
func writeMDChunking(sb *strings.Builder, r *Report) {
	if r.Chunking == nil {
		return
	}
	c := r.Chunking
	sb.WriteString("## Execution & concurrency\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("|-------|-------|\n")
	if c.ChunkSize > 0 {
		row(sb, "Chunk size", fmt.Sprintf("%d", c.ChunkSize))
	}
	if c.ChunksTotal > 0 {
		row(sb, "Chunks", fmt.Sprintf("%d / %d", c.ChunksCompleted, c.ChunksTotal))
	}
	if c.ScanTimeoutPerProject != "" {
		row(sb, "Scan timeout / project", c.ScanTimeoutPerProject)
	}
	if c.DeleteConcurrency > 0 {
		row(sb, "Delete concurrency", fmt.Sprintf("%d", c.DeleteConcurrency))
	}
	if c.BackupConcurrency > 0 {
		row(sb, "Backup concurrency", fmt.Sprintf("%d", c.BackupConcurrency))
	}
	if c.ResumedFrom != "" {
		row(sb, "Resumed from", code(c.ResumedFrom))
	}
	row(sb, "Reference scan", boolStr(!c.SkipReferences))
	row(sb, "Compress binaries", boolStr(c.Compress))
	sb.WriteString("\n")
}

// writeMDEntityBreakdown emits the per-project × entity-type matrix
// using the canonical column order. Skipped when the matrix is empty.
func writeMDEntityBreakdown(sb *strings.Builder, r *Report) {
	if len(r.EntityBreakdown) == 0 {
		return
	}
	cols := []string{"case", "run", "plan", "plan_entry", "result", "test"}
	sb.WriteString("## Per-project × entity-type breakdown\n\n")
	sb.WriteString("| Project |")
	for _, c := range cols {
		fmt.Fprintf(sb, " %s |", c)
	}
	sb.WriteString(" Total | Bytes |\n|---------|")
	for range cols {
		sb.WriteString("------|")
	}
	sb.WriteString("-------|-------|\n")
	totals := make(map[string]int, len(cols))
	var grandTotal int
	var grandBytes int64
	for _, row := range r.EntityBreakdown {
		name := fmt.Sprintf("%s (%d)", escapePipe(row.ProjectName), row.ProjectID)
		fmt.Fprintf(sb, "| %s |", name)
		for _, c := range cols {
			n := row.Counts[c]
			totals[c] += n
			fmt.Fprintf(sb, " %d |", n)
		}
		fmt.Fprintf(sb, " %d | %s |\n", row.Total, humanBytes(row.Bytes))
		grandTotal += row.Total
		grandBytes += row.Bytes
	}
	sb.WriteString("| **TOTAL** |")
	for _, c := range cols {
		fmt.Fprintf(sb, " **%d** |", totals[c])
	}
	fmt.Fprintf(sb, " **%d** | **%s** |\n\n", grandTotal, humanBytes(grandBytes))
}

// writeMDSnapshot emits the snapshot-artifacts inventory: file paths
// inside the snapshot directory and their counters (mapping/integrity
// /references). Skipped on dry-run or when no snapshot was taken.
func writeMDSnapshot(sb *strings.Builder, r *Report) {
	if r.Snapshot == nil || r.SnapshotID == "" || r.DryRun {
		return
	}
	s := r.Snapshot
	sb.WriteString("## Snapshot artifacts\n\n")
	sb.WriteString("Files written to the snapshot directory and their role:\n\n")
	sb.WriteString("| File | Role | Path |\n|------|------|------|\n")
	if s.MetaPath != "" {
		fmt.Fprintf(sb, "| `meta.json` | Snapshot metadata (id, op, entity_ids, label) | `%s` |\n", s.MetaPath)
	}
	if s.MappingPath != "" {
		fmt.Fprintf(sb, "| `attachments.json` | v3.6 mapping schema=%d, sha256 per file | `%s` |\n",
			s.MappingSchemaVersion, s.MappingPath)
	}
	if s.ReferencesPath != "" {
		fmt.Fprintf(sb, "| `references.json` | Markdown URL refs in case/run/plan/milestone bodies | `%s` |\n", s.ReferencesPath)
	}
	if s.IntegrityPath != "" {
		fmt.Fprintf(sb, "| `integrity.json` | Per-file sha256 + Merkle root over the snapshot dir | `%s` |\n", s.IntegrityPath)
	}
	if s.FilesDir != "" {
		fmt.Fprintf(sb, "| `files/` | Backed-up attachment binaries (one per attachment) | `%s` |\n", s.FilesDir)
	}
	sb.WriteString("\n")

	sb.WriteString("Counters:\n\n| Metric | Value |\n|--------|-------|\n")
	if s.MappingTotal > 0 {
		row(sb, "Mapping entries", fmt.Sprintf("%d", s.MappingTotal))
		row(sb, "Restorable entries", fmt.Sprintf("%d / %d", s.MappingRestorable, s.MappingTotal))
	}
	if s.FilesCount > 0 {
		row(sb, "Files in `files/`", fmt.Sprintf("%d (%s)", s.FilesCount, humanBytes(s.FilesBytes)))
	}
	if s.ReferencesSkipped {
		row(sb, "Reference scan", "skipped (--skip-references)")
	} else {
		row(sb, "Entities with references", fmt.Sprintf("%d", s.EntitiesScanned))
		row(sb, "Markdown URL refs indexed", fmt.Sprintf("%d", s.RefsIndexed))
	}
	if s.IntegrityRoot != "" {
		row(sb, "Integrity Merkle root", code(s.IntegrityRoot))
	}
	if s.IntegrityFiles > 0 {
		row(sb, "Integrity files covered", fmt.Sprintf("%d", s.IntegrityFiles))
	}
	sb.WriteString("\n")
}

// writeMDReferenceLimitations emits the audit-only "Known limitations"
// callout for indexed-but-not-rewritten markdown URL references. v3.6.0
// records refs in references.json but does NOT modify external entity
// bodies. The section is skipped on dry-run, when the index is empty,
// or when the reference scan was disabled.
func writeMDReferenceLimitations(sb *strings.Builder, r *Report) {
	if r.Snapshot == nil || r.DryRun {
		return
	}
	s := r.Snapshot
	if s.ReferencesSkipped || s.RefsIndexed == 0 {
		return
	}
	sb.WriteString("## Known limitations: markdown references\n\n")
	fmt.Fprintf(sb, "**%d markdown URL reference(s)** to deleted attachments were indexed in `references.json` "+
		"across %d entity(ies), but **NOT rewritten** in v3.6.0.\n\n", s.RefsIndexed, s.EntitiesScanned)
	if len(s.ReferencesByEntity) > 0 {
		sb.WriteString("Distribution by entity type:\n\n")
		sb.WriteString("| Entity type | Refs indexed |\n|-------------|--------------|\n")
		// Stable order.
		keys := make([]string, 0, len(s.ReferencesByEntity))
		for k := range s.ReferencesByEntity {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(sb, "| %s | %d |\n", k, s.ReferencesByEntity[k])
		}
		sb.WriteString("\n")
	}
	sb.WriteString("After this cleanup, any markdown link of the form ")
	sb.WriteString("`[label](.../attachments/get/<old_id>)` inside another entity's body will return 404. ")
	sb.WriteString("On rollback the attachments are restored under **new** TestRail IDs, so the old links remain broken even after a successful restore.\n\n")
	sb.WriteString("Mitigation: the full ref index is preserved in `references.json` for manual rewrite. ")
	sb.WriteString("Automatic rewrite is tracked for a future release (out of v3.6.0 scope due to read-modify-write race risk on shared TestRail entities).\n\n")
}

// writeMDArtifacts lists every file produced by this run on the local
// file system: audit reports (md/json/csv/pdf), snapshot directory,
// and the checkpoint cache (used by --resume).
func writeMDArtifacts(sb *strings.Builder, r *Report) {
	if r.Artifacts == nil {
		return
	}
	a := r.Artifacts
	if len(a.ReportPaths) == 0 && a.SnapshotPath == "" && a.CheckpointDir == "" {
		return
	}
	sb.WriteString("## Files on disk\n\n")
	if len(a.ReportPaths) > 0 {
		sb.WriteString("**Audit reports** (this report in every supported format):\n\n")
		for _, p := range a.ReportPaths {
			fmt.Fprintf(sb, "- `%s`\n", p)
		}
		sb.WriteString("\n")
	}
	if a.SnapshotPath != "" {
		fmt.Fprintf(sb, "**Snapshot directory**: `%s`\n\n", a.SnapshotPath)
	}
	if a.CheckpointDir != "" {
		fmt.Fprintf(sb, "**Checkpoint cache** (used by `--resume`): `%s`\n\n", a.CheckpointDir)
	}
}

func hasItems(r *Report) bool {
	for _, p := range r.Projects {
		if len(p.Items) > 0 {
			return true
		}
	}
	return false
}

func row(sb *strings.Builder, k, v string) {
	fmt.Fprintf(sb, "| %s | %s |\n", k, v)
}

func notEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func code(s string) string { return "`" + s + "`" }

func joinInts(xs []int64) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = fmt.Sprintf("%d", x)
	}
	return strings.Join(parts, ", ")
}

func joinStr(xs []string) string {
	if len(xs) == 0 {
		return "—"
	}
	return strings.Join(xs, ", ")
}

func escapePipe(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

func humanBytes(n int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.2f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.2f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
