// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Korrnals/gotr/internal/cleanup"
	"github.com/Korrnals/gotr/internal/paths"
	cleanupreport "github.com/Korrnals/gotr/internal/report/cleanup"
	"github.com/Korrnals/gotr/internal/report/pdf"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
)

// writeCleanupReport persists a deletion-audit report under
// ~/.gotr/reports/cleanup-attachments/ unless --no-report is set.
// Failures are reported as warnings — they never abort the cleanup.
// Returns absolute paths of generated report files (empty when skipped
// or on error).
func writeCleanupReport(cmd *cobra.Command, plan *cleanup.Plan, res *cleanup.ExecuteResult, opts *cleanupOptions) []string {
	if opts.NoReport {
		return nil
	}
	if plan == nil || plan.TotalCount == 0 {
		// Nothing to report on.
		return nil
	}

	reportsDir, err := paths.ReportsDirPath()
	if err != nil {
		ui.Warningf(os.Stderr, "report: resolve reports dir: %v", err)
		return nil
	}

	user := viper.GetString("username")
	if user == "" {
		user = "unknown"
	}

	rep := cleanupreport.Build(cleanupreport.BuildInput{
		Plan:         plan,
		Result:       res,
		Timestamp:    time.Now().UTC(),
		Server:       snap.CurrentServerURL(),
		GotrVer:      rootVersionFromCmd(cmd),
		Label:        opts.SnapshotLabel,
		User:         user,
		CLIArgs:      cliArgsFor(cmd),
		ProjectIDs:   opts.ProjectIDs,
		AllProjects:  opts.AllProjects,
		OlderThanRaw: opts.OlderThanRaw,
		CutoffUnix:   opts.CutoffUnix,
		EntityTypes:  opts.EntityTypes,
		ScanStrategy: opts.ScanStrategy,
		Limit:        opts.Limit,
	})

	wopts := cleanupreport.AllFormats()
	wopts.PDFRenderer = pdf.NewCleanupGenerator()

	out, err := cleanupreport.Write(context.Background(), reportsDir, rep, wopts)
	if err != nil {
		ui.Warningf(os.Stderr, "report: write cleanup report: %v", err)
		return nil
	}
	return out.Files()
}

// rootVersionFromCmd extracts the gotr version from the root command's
// Version field. Returns an empty string if the root has no version set
// (e.g. during tests with a synthetic command tree).
func rootVersionFromCmd(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	root := cmd.Root()
	if root == nil {
		return ""
	}
	return root.Version
}
