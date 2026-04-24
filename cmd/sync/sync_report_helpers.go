package sync

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/paths"
	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/Korrnals/gotr/internal/report/pdf"
	"github.com/Korrnals/gotr/internal/service/migration"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type reportResourceStats struct {
	Resource string
	Source   int64
	Created  int64
	Updated  int64
	Skipped  int64
	Failed   int64
}

func saveMigrationReport(ctx context.Context, cmd *cobra.Command, migrationType string, srcProject, dstProject int64, startedAt time.Time, hook *snap.Hook, stats []reportResourceStats) string {
	if len(stats) == 0 {
		return ""
	}

	reportsDir, err := paths.ReportsDirPath()
	if err != nil {
		ui.Warningf(os.Stderr, "report: resolve reports dir: %v", err)
		return ""
	}

	snapshotID := "no_snapshot"
	if hook != nil && hook.Enabled && hook.Snap != nil && hook.Snap.Meta != nil && hook.Snap.Meta.ID != "" {
		snapshotID = hook.Snap.Meta.ID
	}

	user := viper.GetString("username")
	if user == "" {
		user = "unknown"
	}

	reportObj := intreport.NewMigrationReport(snapshotID, srcProject, dstProject, migrationType, user)
	reportObj.Label = resolveSnapLabel(hook)
	for _, s := range stats {
		reportObj.AddResourceStats(s.Resource, s.Source, s.Created, s.Updated, s.Skipped, s.Failed)
	}

	if snapshotID != "no_snapshot" {
		reportObj.SetRollbackInfo(snapshotID, true, []string{"case", "section", "shared_step", "suite"})
	}

	totalProcessed := reportObj.GetTotalCreated() + reportObj.GetTotalSkipped() + reportObj.GetTotalFailed()
	reportObj.SetPerformance(time.Since(startedAt), totalProcessed, 0)

	if reportObj.GetTotalFailed() > 0 {
		if reportObj.GetTotalCreated() == 0 {
			reportObj.MarkFailed()
		} else {
			reportObj.MarkPartial()
		}
	} else {
		reportObj.MarkSuccess()
	}

	service := intreport.NewService(reportsDir)
	reportPath, err := service.Save(ctx, reportObj)
	if err != nil {
		ui.Warningf(os.Stderr, "report: save migration report failed: %v", err)
		return ""
	}

	ui.Infof(os.Stdout, "Migration report saved: %s", reportPath)
	maybeRenderPDF(cmd, reportObj, reportPath)

	return reportPath
}

// maybeRenderPDF writes a PDF counterpart of the migration report when the
// --pdf-report flag is set on the command. Failures are logged but not fatal.
func maybeRenderPDF(cmd *cobra.Command, reportObj *intreport.MigrationReport, reportPath string) {
	wantPDF, _ := cmd.Flags().GetBool("pdf-report")
	if !wantPDF {
		return
	}
	pdfPath := strings.TrimSuffix(reportPath, ".md") + ".pdf"
	if err := pdf.NewGenerator().Save(reportObj, pdfPath); err != nil {
		ui.Warningf(os.Stderr, "report: render PDF failed: %v", err)
		return
	}
	ui.Infof(os.Stdout, "Migration report (PDF) saved: %s", pdfPath)
}

func toCount(v int) int64 {
	if v < 0 {
		return 0
	}
	return int64(v)
}

// resolveSnapLabel extracts the snapshot label (if any) from an active hook.
// It returns "" when no label is available so callers can fall back to the
// "default" bucket in the report hierarchy.
func resolveSnapLabel(hook *snap.Hook) string {
	if hook == nil || !hook.Enabled || hook.Snap == nil || hook.Snap.Meta == nil {
		return ""
	}
	return hook.Snap.Meta.Label
}

func filterStatsToReport(resource string, s migration.FilterStats, created, failed int64) reportResourceStats {
	return reportResourceStats{
		Resource: resource,
		Source:   toCount(s.Source),
		Created:  created,
		Skipped:  toCount(s.Duplicates),
		Failed:   failed,
	}
}
