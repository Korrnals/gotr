package report

import "time"

// MigrationReport represents a complete migration summary
type MigrationReport struct {
	ID              string
	SnapshotID      string
	Timestamp       time.Time
	Duration        time.Duration
	Status          string // "success", "partial", "failed"

	SourceProjectID int64
	TargetProjectID int64
	MigrationType   string // "sync_full", "sync_cases", "sync_shared_steps", etc
	User            string

	Summary     map[string]*ResourceStats // "cases", "shared_steps", "sections", "suites", "attachments"
	Skipped     map[string][]SkipReason
	Rollback    RollbackInfo
	Performance PerfMetrics
}

// ResourceStats holds migration statistics for a resource type
type ResourceStats struct {
	SourceCount int64
	Created     int64
	Updated     int64
	Skipped     int64
	Failed      int64
}

// SkipReason explains why a resource was skipped
type SkipReason struct {
	ID     int64
	Reason string
	Detail string
}

// RollbackInfo contains rollback information for this migration
type RollbackInfo struct {
	SnapshotID  string
	Enabled     bool
	DeleteOrder []string // ["case", "section", "shared_step", "suite"]
	Command     string   // Full rollback command
}

// PerfMetrics tracks migration performance
type PerfMetrics struct {
	TotalTime      time.Duration
	EntitiesPerSec float64
	PeakMemoryMB   int64
}

// NewMigrationReport creates a new migration report
func NewMigrationReport(
	snapshotID string,
	sourceProjectID int64,
	targetProjectID int64,
	migrationType string,
	user string,
) *MigrationReport {
	return &MigrationReport{
		ID:              generateReportID(),
		SnapshotID:      snapshotID,
		Timestamp:       time.Now(),
		SourceProjectID: sourceProjectID,
		TargetProjectID: targetProjectID,
		MigrationType:   migrationType,
		User:            user,
		Status:          "pending",
		Summary:         make(map[string]*ResourceStats),
		Skipped:         make(map[string][]SkipReason),
	}
}

// AddResourceStats adds statistics for a resource type
func (mr *MigrationReport) AddResourceStats(
	resourceType string,
	sourceCount, created, updated, skipped, failed int64,
) {
	mr.Summary[resourceType] = &ResourceStats{
		SourceCount: sourceCount,
		Created:     created,
		Updated:     updated,
		Skipped:     skipped,
		Failed:      failed,
	}
}

// AddSkipped records skipped resources with reasons
func (mr *MigrationReport) AddSkipped(resourceType string, reasons []SkipReason) {
	mr.Skipped[resourceType] = reasons
}

// SetRollbackInfo sets rollback information
func (mr *MigrationReport) SetRollbackInfo(
	snapshotID string,
	enabled bool,
	deleteOrder []string,
) {
	mr.Rollback = RollbackInfo{
		SnapshotID:  snapshotID,
		Enabled:     enabled,
		DeleteOrder: deleteOrder,
		Command:     "gotr snap rollback " + snapshotID,
	}
}

// SetPerformance sets performance metrics
func (mr *MigrationReport) SetPerformance(
	duration time.Duration,
	totalEntities int64,
	peakMemoryMB int64,
) {
	mr.Duration = duration
	mr.Performance = PerfMetrics{
		TotalTime: duration,
		EntitiesPerSec: func() float64 {
			if duration.Seconds() == 0 {
				return 0
			}
			return float64(totalEntities) / duration.Seconds()
		}(),
		PeakMemoryMB: peakMemoryMB,
	}
}

// MarkSuccess marks migration as successful
func (mr *MigrationReport) MarkSuccess() {
	mr.Status = "success"
}

// MarkPartial marks migration as partial (some skipped/failed)
func (mr *MigrationReport) MarkPartial() {
	mr.Status = "partial"
}

// MarkFailed marks migration as failed
func (mr *MigrationReport) MarkFailed() {
	mr.Status = "failed"
}

// GetTotalCreated returns total entities created across all resource types
func (mr *MigrationReport) GetTotalCreated() int64 {
	var total int64
	for _, stats := range mr.Summary {
		total += stats.Created
	}
	return total
}

// GetTotalSkipped returns total entities skipped across all resource types
func (mr *MigrationReport) GetTotalSkipped() int64 {
	var total int64
	for _, stats := range mr.Summary {
		total += stats.Skipped
	}
	return total
}

// GetTotalFailed returns total entities failed across all resource types
func (mr *MigrationReport) GetTotalFailed() int64 {
	var total int64
	for _, stats := range mr.Summary {
		total += stats.Failed
	}
	return total
}

func generateReportID() string {
	return time.Now().Format("20060102150405")
}
