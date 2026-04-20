package report

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServiceSave(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewService(tmpDir)

	report := NewMigrationReport("snap-abc123", 100, 200, "sync_full", "alice")
	report.AddResourceStats("cases", 100, 80, 0, 20, 0)
	report.AddResourceStats("shared_steps", 50, 45, 0, 5, 0)
	report.MarkSuccess()
	report.SetPerformance(10*time.Second, 125, 256)
	report.SetRollbackInfo("snap-abc123", true, []string{"case", "section", "shared_step", "suite"})

	path, err := service.Save(context.Background(), report)
	if err != nil {
		t.Fatalf("failed to save report: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("report file not found: %v", err)
	}

	// Verify filename format
	filename := filepath.Base(path)
	if !contains(filename, "migration-") || !contains(filename, "snap-abc123") {
		t.Errorf("unexpected filename: %s", filename)
	}

	// Verify content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "snap-abc123") {
		t.Error("report missing snapshot ID")
	}
	if !contains(contentStr, "cases") {
		t.Error("report missing cases")
	}
	if !contains(contentStr, "80") {
		t.Error("report missing created count")
	}
	if !contains(contentStr, "Rollback") {
		t.Error("report missing rollback section")
	}
}

func TestServiceSave_SnapshotIDWithSlash(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewService(tmpDir)

	report := NewMigrationReport("cases/20260418T120000_update_42", 100, 200, "sync_cases", "alice")
	report.AddResourceStats("cases", 10, 8, 0, 2, 0)
	report.MarkSuccess()

	path, err := service.Save(context.Background(), report)
	if err != nil {
		t.Fatalf("failed to save report with slash snapshot id: %v", err)
	}

	filename := filepath.Base(path)
	if contains(filename, "/") {
		t.Fatalf("filename must be sanitized, got: %s", filename)
	}
	if !contains(filename, "cases_20260418T120000_update_42") {
		t.Fatalf("sanitized snapshot id not found in filename: %s", filename)
	}
}

func TestServiceUpdateIndex(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewService(tmpDir)

	// Save multiple reports
	for i := 0; i < 3; i++ {
		report := NewMigrationReport("snap-abc", 100, 200, "sync_full", "alice")
		report.AddResourceStats("cases", 100, 80, 0, 20, 0)
		report.MarkSuccess()
		service.Save(context.Background(), report)
		time.Sleep(100 * time.Millisecond) // Ensure different timestamps
	}

	// Check index file exists
	indexPath := filepath.Join(tmpDir, "INDEX.md")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("index file not found: %v", err)
	}

	// Verify index content
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read index: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "Migration Reports Index") {
		t.Error("index missing title")
	}
	if !contains(contentStr, "migration-") {
		t.Error("index missing migration reports")
	}
}

func TestGenerateMarkdown(t *testing.T) {
	service := NewService("")
	report := NewMigrationReport("snap-123", 100, 200, "sync_full", "alice")
	report.AddResourceStats("cases", 100, 80, 0, 20, 0)
	report.AddResourceStats("shared_steps", 50, 45, 0, 5, 0)
	report.AddSkipped("cases", []SkipReason{
		{ID: 1, Reason: "custom_field_mismatch", Detail: "field not in target"},
	})
	report.MarkSuccess()
	report.SetPerformance(10*time.Second, 125, 256)
	report.SetRollbackInfo("snap-123", true, []string{"case", "section", "shared_step", "suite"})

	md := service.generateMarkdown(report)

	testCases := []string{
		"# Migration Report",
		"snap-123",
		"Source Project",
		"Target Project",
		"cases",
		"shared_steps",
		"100",
		"80",
		"TOTAL",
		"Rollback",
		"Performance",
		"Entities/sec",
	}

	for _, test := range testCases {
		if !contains(md, test) {
			t.Errorf("markdown missing: %s", test)
		}
	}
}

func TestGenerateMarkdownWithoutRollback(t *testing.T) {
	service := NewService("")
	report := NewMigrationReport("snap-123", 100, 200, "sync_full", "alice")
	report.AddResourceStats("cases", 100, 80, 0, 20, 0)
	report.MarkSuccess()

	md := service.generateMarkdown(report)

	if contains(md, "Rollback") {
		t.Error("markdown should not have rollback section when not set")
	}
}

func TestGenerateMarkdown_ReferencesSnapsPath(t *testing.T) {
	service := NewService("")
	report := NewMigrationReport("cases/snap-123", 100, 200, "sync_full", "alice")
	report.AddResourceStats("cases", 1, 1, 0, 0, 0)
	report.MarkSuccess()

	md := service.generateMarkdown(report)
	if !contains(md, "~/.gotr/snaps/") {
		t.Fatal("markdown references should point to ~/.gotr/snaps/")
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
