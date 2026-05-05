// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Korrnals/gotr/internal/cleanup"
	"github.com/Korrnals/gotr/internal/cleanup/checkpoint"
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
func writeCleanupReport(cmd *cobra.Command, plan *cleanup.Plan, res *cleanup.ExecuteResult, opts *cleanupOptions, runID string) []string {
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

	in := cleanupreport.BuildInput{
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

		RunID:                 runID,
		ChunkSize:             opts.ChunkSize,
		ScanTimeoutPerProject: opts.ScanTimeoutRaw,
		DeleteConcurrency:     opts.Concurrency,
		BackupConcurrency:     opts.BackupConcurrency,
		ResumedFrom:           opts.ResumeRunID,
		SkipReferences:        opts.SkipReferences,
		Compress:              opts.Compress,
	}

	// Populate snapshot artifact info from disk when a snapshot exists.
	populateSnapshotArtifacts(&in, res)
	// Resolve checkpoint cache directory for the current run.
	if runID != "" {
		if cps, err := checkpoint.NewStore(); err == nil {
			in.CheckpointDir = filepath.Join(cps.Root(), runID)
		}
	}

	rep := cleanupreport.Build(in)

	wopts := cleanupreport.AllFormats()
	wopts.PDFRenderer = pdf.NewCleanupGenerator()

	// Self-reference: predict output file paths before writing so the
	// "Artifacts" section can list them. PredictPaths must mirror the
	// Write naming convention.
	in.ReportPaths = cleanupreport.PredictPaths(reportsDir, rep, wopts)
	rep = cleanupreport.Build(in)

	out, err := cleanupreport.Write(context.Background(), reportsDir, rep, wopts)
	if err != nil {
		ui.Warningf(os.Stderr, "report: write cleanup report: %v", err)
		return nil
	}
	return out.Files()
}

// populateSnapshotArtifacts inspects the on-disk snapshot directory for
// the run's snapshot (when present) and fills the snapshot-related
// fields on BuildInput. All errors degrade gracefully — the report is
// still produced, just with fewer counters.
func populateSnapshotArtifacts(in *cleanupreport.BuildInput, res *cleanup.ExecuteResult) {
	if res == nil || res.SnapshotID == "" {
		return
	}
	store, _, err := openSnapStore()
	if err != nil {
		return
	}
	snapID := res.SnapshotID
	dir := store.SnapDir(snapID)
	in.SnapshotPath = dir
	in.MetaPath = filepath.Join(dir, "meta.json")
	in.MappingPath = filepath.Join(dir, "attachments.json")
	in.ReferencesPath = filepath.Join(dir, "references.json")
	in.IntegrityPath = filepath.Join(dir, "integrity.json")
	in.FilesDir = filepath.Join(dir, "files")

	populateMappingCounters(in, store, snapID)
	populateIntegrityCounters(in, store, snapID)
	populateFilesCounters(in)
	populateReferencesCounters(in, store, snapID)
}

// populateMappingCounters fills MappingSchemaVersion / MappingTotal /
// MappingRestorable from attachments.json.
func populateMappingCounters(in *cleanupreport.BuildInput, store *snap.Store, snapID string) {
	m, err := snap.LoadMapping(store, snapID)
	if err != nil || m == nil {
		return
	}
	in.MappingSchemaVersion = m.SchemaVersion
	in.MappingTotal = len(m.Entries)
	for i := range m.Entries {
		if m.Entries[i].Restorable {
			in.MappingRestorable++
		}
	}
}

// populateIntegrityCounters fills IntegrityFiles from integrity.json.
func populateIntegrityCounters(in *cleanupreport.BuildInput, store *snap.Store, snapID string) {
	idx, err := loadIntegrityIndex(store, snapID)
	if err != nil || idx == nil {
		return
	}
	in.IntegrityFiles = len(idx.Files)
}

// populateFilesCounters scans <snap>/files/ for binary payloads and
// fills FilesCount + FilesBytes.
func populateFilesCounters(in *cleanupreport.BuildInput) {
	entries, err := os.ReadDir(in.FilesDir)
	if err != nil {
		return
	}
	var count int
	var bytes int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, ierr := e.Info()
		if ierr != nil {
			continue
		}
		if info.Mode()&fs.ModeType != 0 {
			continue
		}
		count++
		bytes += info.Size()
	}
	in.FilesCount = count
	in.FilesBytes = bytes
}

// populateReferencesCounters fills ReferencesByEntity from
// references.json (audit-only, no rewrite in v3.6.0).
func populateReferencesCounters(in *cleanupreport.BuildInput, store *snap.Store, snapID string) {
	list, err := snap.LoadReferencesSidecar(store, snapID)
	if err != nil {
		return
	}
	byEntity := map[string]int{}
	for _, e := range list {
		if len(e.Refs) == 0 {
			continue
		}
		byEntity[e.EntityType] += len(e.Refs)
	}
	if len(byEntity) > 0 {
		in.ReferencesByEntity = byEntity
	}
}

// loadIntegrityIndex reads <snap>/integrity.json. Errors degrade
// gracefully (used for report counters only).
func loadIntegrityIndex(store *snap.Store, snapID string) (*snap.IntegrityIndex, error) {
	var idx snap.IntegrityIndex
	if err := store.LoadData(snapID, "integrity.json", &idx); err != nil {
		return nil, err
	}
	return &idx, nil
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
