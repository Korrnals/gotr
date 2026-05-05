// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"fmt"
	"strings"
	"time"
)

// RenderMarkdown produces the human-readable Markdown rendition of the
// deletion report. Layout is deterministic so golden tests stay stable.
func RenderMarkdown(r *Report) string {
	var sb strings.Builder

	title := "Attachments Cleanup Report"
	if r.DryRun {
		title += " (DRY-RUN)"
	}
	fmt.Fprintf(&sb, "# %s\n\n", title)

	if r.DryRun {
		sb.WriteString("> **DRY-RUN** — no snapshot was taken and no attachments were deleted.\n\n")
	}

	// Header table.
	sb.WriteString("## Run\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("|-------|-------|\n")
	row(&sb, "Report ID", code(r.ID))
	row(&sb, "Snapshot ID", code(notEmpty(r.SnapshotID, "—")))
	row(&sb, "Timestamp (UTC)", r.Timestamp.UTC().Format(time.RFC3339))
	row(&sb, "Server", notEmpty(r.Server, "—"))
	row(&sb, "gotr version", notEmpty(r.GotrVer, "—"))
	row(&sb, "Label", notEmpty(r.Label, "—"))
	row(&sb, "User", notEmpty(r.User, "—"))
	row(&sb, "Dry-run", boolStr(r.DryRun))
	if len(r.CLIArgs) > 0 {
		row(&sb, "CLI", code("gotr "+strings.Join(r.CLIArgs, " ")))
	}
	sb.WriteString("\n")

	// Filters.
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
	row(&sb, "Scope", scope)
	row(&sb, "Older than", notEmpty(r.Filters.OlderThan, "—"))
	if r.Filters.CutoffUnix > 0 {
		row(&sb, "Cutoff (UTC)", time.Unix(r.Filters.CutoffUnix, 0).UTC().Format(time.RFC3339))
	}
	row(&sb, "Entity types", joinStr(r.Filters.EntityTypes))
	row(&sb, "Scan strategy", notEmpty(r.Filters.ScanStrategy, "auto"))
	if r.Filters.Limit > 0 {
		row(&sb, "Limit", fmt.Sprintf("%d", r.Filters.Limit))
	}
	sb.WriteString("\n")

	// Summary.
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	row(&sb, "Total selected", fmt.Sprintf("%d", r.Summary.TotalSelected))
	row(&sb, "Backed up (count)", fmt.Sprintf("%d", r.Summary.BackedUp))
	row(&sb, "Backed up (bytes)", humanBytes(r.Summary.BackupBytes))
	row(&sb, "Deleted", fmt.Sprintf("%d", r.Summary.Deleted))
	row(&sb, "Failed", fmt.Sprintf("%d", r.Summary.Failed))
	sb.WriteString("\n")

	// Per-project.
	if len(r.Projects) > 0 {
		sb.WriteString("## Per-project breakdown\n\n")
		sb.WriteString("| Project | Name | Count | Bytes | Oldest |\n")
		sb.WriteString("|---------|------|-------|-------|--------|\n")
		for _, p := range r.Projects {
			oldest := "—"
			if p.OldestUnix > 0 {
				oldest = time.Unix(p.OldestUnix, 0).UTC().Format("2006-01-02")
			}
			fmt.Fprintf(&sb, "| %d | %s | %d | %s | %s |\n",
				p.ProjectID, escapePipe(p.ProjectName), p.Count, humanBytes(p.TotalBytes), oldest)
		}
		sb.WriteString("\n")
	}

	// Full deleted list.
	if hasItems(r) {
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
				fmt.Fprintf(&sb, "| %d | %d | %s | %s | %s | %s |\n",
					p.ProjectID, it.AttachmentID, escapePipe(it.Name),
					humanBytes(it.Size), parent, created)
			}
		}
		sb.WriteString("\n")
	}

	// Failures.
	if len(r.Failures) > 0 {
		sb.WriteString("## Failures\n\n")
		sb.WriteString("| Attachment ID | Project ID | Error |\n")
		sb.WriteString("|---------------|------------|-------|\n")
		for _, f := range r.Failures {
			fmt.Fprintf(&sb, "| %d | %d | %s |\n",
				f.AttachmentID, f.ProjectID, escapePipe(f.Error))
		}
		sb.WriteString("\n")
	}

	// Footer.
	sb.WriteString("## Rollback\n\n")
	if r.SnapshotID != "" && !r.DryRun {
		fmt.Fprintf(&sb, "Snapshot is preserved under `~/.gotr/snaps/cleanup-attachments/%s/`.\n\n", r.SnapshotID)
		fmt.Fprintf(&sb, "Restore with:\n\n```\ngotr snap rollback %s\n```\n", r.SnapshotID)
	} else if r.DryRun {
		sb.WriteString("_No snapshot was taken (dry-run)._\n")
	} else {
		sb.WriteString("_No snapshot reference recorded._\n")
	}

	return sb.String()
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
