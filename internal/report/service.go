package report

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Service handles migration report generation and storage
type Service struct {
	reportsDir string
}

// NewService creates a new report service
func NewService(reportsDir string) *Service {
	return &Service{
		reportsDir: reportsDir,
	}
}

// Save saves a migration report to disk
func (s *Service) Save(ctx context.Context, report *MigrationReport) (string, error) {
	// Ensure reports directory exists
	if err := os.MkdirAll(s.reportsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Generate markdown content
	md := s.generateMarkdown(report)

	// Create filename: migration-{timestamp}-{snapshot-id}.md
	filename := fmt.Sprintf("migration-%s-%s.md",
		report.Timestamp.Format("20060102T150405Z"),
		sanitizeSnapshotID(report.SnapshotID),
	)
	filepath := filepath.Join(s.reportsDir, filename)

	// Write report file
	if err := os.WriteFile(filepath, []byte(md), 0644); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	// Update index
	if err := s.updateIndex(ctx); err != nil {
		// Log but don't fail if index update fails
		fmt.Fprintf(os.Stderr, "warning: failed to update report index: %v\n", err)
	}

	return filepath, nil
}

// generateMarkdown generates the markdown content for a report
func (s *Service) generateMarkdown(report *MigrationReport) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# Migration Report: `%s`\n", report.SnapshotID))
	sb.WriteString(fmt.Sprintf("**Date:** %s | **Duration:** %.1fs | **Status:** %s\n\n",
		report.Timestamp.Format("2006-01-02 15:04:05 UTC"),
		report.Duration.Seconds(),
		statusEmoji(report.Status),
	))

	// Configuration Table
	sb.WriteString("## Configuration\n")
	sb.WriteString("| Parameter | Value |\n")
	sb.WriteString("|-----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Source Project | %d |\n", report.SourceProjectID))
	sb.WriteString(fmt.Sprintf("| Target Project | %d |\n", report.TargetProjectID))
	sb.WriteString(fmt.Sprintf("| Migration Type | `%s` |\n", report.MigrationType))
	sb.WriteString(fmt.Sprintf("| User | %s |\n\n", report.User))

	// Summary Table
	sb.WriteString("## Summary\n")
	sb.WriteString("| Resource Type | Source | Created | Updated | Skipped | Failed |\n")
	sb.WriteString("|---------------|--------|---------|---------|---------|--------|\n")

	// Sort resource types for consistent output
	var resourceTypes []string
	for rt := range report.Summary {
		resourceTypes = append(resourceTypes, rt)
	}
	sort.Strings(resourceTypes)

	var totalSource, totalCreated, totalUpdated, totalSkipped, totalFailed int64

	for _, rt := range resourceTypes {
		stats := report.Summary[rt]
		sb.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %d | %d |\n",
			rt, stats.SourceCount, stats.Created, stats.Updated, stats.Skipped, stats.Failed))

		totalSource += stats.SourceCount
		totalCreated += stats.Created
		totalUpdated += stats.Updated
		totalSkipped += stats.Skipped
		totalFailed += stats.Failed
	}

	sb.WriteString(fmt.Sprintf("| **TOTAL** | **%d** | **%d** | **%d** | **%d** | **%d** |\n\n",
		totalSource, totalCreated, totalUpdated, totalSkipped, totalFailed))

	// Details Section
	if len(report.Skipped) > 0 {
		sb.WriteString("## Skipped Resources\n")
		for _, rt := range resourceTypes {
			if reasons, exists := report.Skipped[rt]; exists && len(reasons) > 0 {
				sb.WriteString(fmt.Sprintf("\n### %s\n", strings.Title(rt)))
				sb.WriteString(fmt.Sprintf("**Total Skipped:** %d\n\n", len(reasons)))

				// Group by reason
				reasonMap := make(map[string]int)
				for _, reason := range reasons {
					reasonMap[reason.Reason]++
				}

				for reason, count := range reasonMap {
					sb.WriteString(fmt.Sprintf("- %s: %d\n", reason, count))
				}
			}
		}
		sb.WriteString("\n")
	}

	// Rollback Section
	if report.Rollback.SnapshotID != "" {
		sb.WriteString("## Rollback\n")
		sb.WriteString("| Parameter | Value |\n")
		sb.WriteString("|-----------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Snapshot ID | `%s` |\n", report.Rollback.SnapshotID))
		sb.WriteString(fmt.Sprintf("| Enabled | %v |\n", report.Rollback.Enabled))
		sb.WriteString(fmt.Sprintf("| Deletion Order | %s |\n", strings.Join(report.Rollback.DeleteOrder, " → ")))
		sb.WriteString(fmt.Sprintf("| Command | `%s` |\n\n", report.Rollback.Command))
	}

	// Performance Section
	if report.Duration > 0 {
		sb.WriteString("## Performance\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Total Time | %.1fs |\n", report.Duration.Seconds()))
		sb.WriteString(fmt.Sprintf("| Entities/sec | %.1f |\n", report.Performance.EntitiesPerSec))
		if report.Performance.PeakMemoryMB > 0 {
			sb.WriteString(fmt.Sprintf("| Peak Memory | %d MB |\n", report.Performance.PeakMemoryMB))
		}
		sb.WriteString("\n")
	}

	// References Section
	sb.WriteString("## References\n")
	sb.WriteString(fmt.Sprintf("- Snapshot: `~/.gotr/snaps/%s`\n", report.SnapshotID))
	sb.WriteString(fmt.Sprintf("- Report: %s\n", report.ID))

	return sb.String()
}

// updateIndex updates the report index file
func (s *Service) updateIndex(ctx context.Context) error {
	// List all migration reports
	entries, err := os.ReadDir(s.reportsDir)
	if err != nil {
		return err
	}

	var reports []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "migration-") && strings.HasSuffix(entry.Name(), ".md") {
			reports = append(reports, entry.Name())
		}
	}

	// Sort in reverse chronological order
	sort.Sort(sort.Reverse(sort.StringSlice(reports)))

	// Generate index content
	var sb strings.Builder
	sb.WriteString("# Migration Reports Index\n\n")
	sb.WriteString("Recent migration reports (newest first):\n\n")

	for i, report := range reports {
		if i >= 50 { // Limit to last 50 in index
			break
		}
		// Extract timestamp from filename for display
		timestamp := strings.TrimPrefix(strings.TrimSuffix(report, ".md"), "migration-")
		sb.WriteString(fmt.Sprintf("- [%s](%s)\n", timestamp, report))
	}

	// Write index file
	indexPath := filepath.Join(s.reportsDir, "INDEX.md")
	return os.WriteFile(indexPath, []byte(sb.String()), 0644)
}

// GetReportsDir returns the reports directory path
func (s *Service) GetReportsDir() string {
	return s.reportsDir
}

func statusEmoji(status string) string {
	switch status {
	case "success":
		return "✅ Success"
	case "partial":
		return "⚠️ Partial"
	case "failed":
		return "❌ Failed"
	default:
		return status
	}
}

func sanitizeSnapshotID(snapshotID string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", " ", "_")
	clean := replacer.Replace(snapshotID)
	if clean == "" {
		return "no_snapshot"
	}
	return clean
}
