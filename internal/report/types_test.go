package report

import (
	"testing"
	"time"
)

func TestNewMigrationReport(t *testing.T) {
	report := NewMigrationReport("snap-123", 100, 200, "sync_full", "alice")

	if report.SnapshotID != "snap-123" {
		t.Errorf("expected snapshot ID snap-123, got %s", report.SnapshotID)
	}

	if report.SourceProjectID != 100 {
		t.Errorf("expected source project 100, got %d", report.SourceProjectID)
	}

	if report.TargetProjectID != 200 {
		t.Errorf("expected target project 200, got %d", report.TargetProjectID)
	}

	if report.MigrationType != "sync_full" {
		t.Errorf("expected migration type sync_full, got %s", report.MigrationType)
	}

	if report.User != "alice" {
		t.Errorf("expected user alice, got %s", report.User)
	}

	if report.Status != "pending" {
		t.Errorf("expected status pending, got %s", report.Status)
	}
}

func TestAddResourceStats(t *testing.T) {
	report := NewMigrationReport("snap-123", 100, 200, "sync_full", "alice")

	report.AddResourceStats("cases", 100, 80, 0, 20, 0)
	report.AddResourceStats("shared_steps", 50, 45, 0, 5, 0)

	if stats, exists := report.Summary["cases"]; !exists || stats.Created != 80 {
		t.Errorf("expected cases created: 80, got %v", stats)
	}

	if stats, exists := report.Summary["shared_steps"]; !exists || stats.Created != 45 {
		t.Errorf("expected shared_steps created: 45, got %v", stats)
	}
}

func TestGetTotals(t *testing.T) {
	report := NewMigrationReport("snap-123", 100, 200, "sync_full", "alice")

	report.AddResourceStats("cases", 100, 80, 5, 10, 5)
	report.AddResourceStats("shared_steps", 50, 45, 2, 3, 0)

	total := report.GetTotalCreated()
	if total != 125 {
		t.Errorf("expected total created 125, got %d", total)
	}

	skipped := report.GetTotalSkipped()
	if skipped != 13 {
		t.Errorf("expected total skipped 13, got %d", skipped)
	}

	failed := report.GetTotalFailed()
	if failed != 5 {
		t.Errorf("expected total failed 5, got %d", failed)
	}
}

func TestSetPerformance(t *testing.T) {
	report := NewMigrationReport("snap-123", 100, 200, "sync_full", "alice")

	duration := time.Duration(10 * time.Second)
	report.SetPerformance(duration, 1000, 512)

	if report.Duration != duration {
		t.Errorf("expected duration 10s, got %v", report.Duration)
	}

	expectedRate := 1000.0 / 10.0
	if report.Performance.EntitiesPerSec != expectedRate {
		t.Errorf("expected rate %.1f, got %.1f", expectedRate, report.Performance.EntitiesPerSec)
	}

	if report.Performance.PeakMemoryMB != 512 {
		t.Errorf("expected peak memory 512MB, got %d", report.Performance.PeakMemoryMB)
	}
}

func TestSetRollbackInfo(t *testing.T) {
	report := NewMigrationReport("snap-123", 100, 200, "sync_full", "alice")

	deleteOrder := []string{"case", "section", "shared_step", "suite"}
	report.SetRollbackInfo("snap-123", true, deleteOrder)

	if !report.Rollback.Enabled {
		t.Error("expected rollback enabled")
	}

	if report.Rollback.SnapshotID != "snap-123" {
		t.Errorf("expected snapshot ID snap-123, got %s", report.Rollback.SnapshotID)
	}

	if len(report.Rollback.DeleteOrder) != 4 {
		t.Errorf("expected 4 delete orders, got %d", len(report.Rollback.DeleteOrder))
	}
}

func TestMarkStatus(t *testing.T) {
	report := NewMigrationReport("snap-123", 100, 200, "sync_full", "alice")

	report.MarkSuccess()
	if report.Status != "success" {
		t.Errorf("expected status success, got %s", report.Status)
	}

	report.MarkPartial()
	if report.Status != "partial" {
		t.Errorf("expected status partial, got %s", report.Status)
	}

	report.MarkFailed()
	if report.Status != "failed" {
		t.Errorf("expected status failed, got %s", report.Status)
	}
}

func TestAddSkipped(t *testing.T) {
	report := NewMigrationReport("snap-123", 100, 200, "sync_full", "alice")

	reasons := []SkipReason{
		{ID: 1, Reason: "custom_field_mismatch", Detail: "field 'priority' not in target"},
		{ID: 2, Reason: "custom_field_mismatch", Detail: "field 'priority' not in target"},
	}

	report.AddSkipped("cases", reasons)

	if len(report.Skipped["cases"]) != 2 {
		t.Errorf("expected 2 skipped cases, got %d", len(report.Skipped["cases"]))
	}

	if report.Skipped["cases"][0].Reason != "custom_field_mismatch" {
		t.Errorf("expected reason custom_field_mismatch, got %s", report.Skipped["cases"][0].Reason)
	}
}
