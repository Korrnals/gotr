// Package report provides migration report generation and management.
//
// Reports are saved to ~/.gotr/reports/ with detailed migration statistics,
// including resource counts, skip reasons, rollback information, and performance metrics.
//
// Example usage:
//
//	svc := report.NewService(reportsDir)
//	rep := report.NewMigrationReport("snap-123", 100, 200, "sync_full", "alice")
//	rep.AddResourceStats("cases", 100, 80, 0, 20, 0)
//	rep.MarkSuccess()
//	path, err := svc.Save(ctx, rep)
package report
