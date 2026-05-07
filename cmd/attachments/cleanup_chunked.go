// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Korrnals/gotr/internal/cleanup"
	"github.com/Korrnals/gotr/internal/cleanup/checkpoint"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// resumeIncompatibleFlags lists CLI flags that drive the project set
// or filter; supplying any of them with --resume is rejected because
// those values are restored from the checkpoint.
var resumeIncompatibleFlags = []string{"project", "all-projects", "older-than", "entity-type"}

// listCheckpointsForbiddenFlags lists every cleanup flag — supplying
// any of them with --list-checkpoints is rejected because the listing
// is a read-only inventory.
var listCheckpointsForbiddenFlags = []string{
	"project", "all-projects", "older-than", "entity-type", "dry-run",
	"concurrency", "limit", "no-snapshot", "snapshot-retention",
	"snap-label", "max-size-gb", "compress", "force", "scan-strategy",
	"no-report", "chunk-size", "resume", "scan-timeout-per-project",
}

func assertResumeExclusive(cmd *cobra.Command) error {
	for _, name := range resumeIncompatibleFlags {
		if cmd.Flags().Changed(name) {
			return fmt.Errorf("--resume is mutually exclusive with --%s (filter values are restored from the checkpoint)", name)
		}
	}
	return nil
}

func assertListCheckpointsExclusive(cmd *cobra.Command) error {
	for _, name := range listCheckpointsForbiddenFlags {
		if cmd.Flags().Changed(name) {
			return fmt.Errorf("--list-checkpoints is mutually exclusive with --%s", name)
		}
	}
	return nil
}

// loadOptsFromCheckpoint hydrates opts from the persisted checkpoint
// for opts.ResumeRunID. The filter snapshot, project list and
// chunk size are restored verbatim; runtime knobs (concurrency,
// scan timeout) keep their CLI-provided values.
func loadOptsFromCheckpoint(opts *cleanupOptions) error {
	store, err := checkpoint.NewStore()
	if err != nil {
		return err
	}
	cp, err := store.Load(opts.ResumeRunID)
	if err != nil {
		return err
	}
	opts.AllProjects = cp.AllProjects
	opts.ProjectIDs = append([]int64(nil), cp.ProjectIDs...)
	opts.OlderThanRaw = cp.Filter.OlderThanRaw
	opts.EntityTypes = append([]string(nil), cp.Filter.EntityTypes...)
	opts.Limit = cp.Filter.Limit
	opts.ScanStrategy = cp.Filter.ScanStrategy
	if opts.ChunkSize <= 0 || cp.ChunkSize > 0 {
		opts.ChunkSize = cp.ChunkSize
	}
	if opts.OlderThanRaw != "" {
		d, err := parseHumanDuration(opts.OlderThanRaw)
		if err != nil {
			return fmt.Errorf("parse stored --older-than %q: %w", opts.OlderThanRaw, err)
		}
		opts.OlderThan = d
	}
	return nil
}

func runListCheckpoints(cmd *cobra.Command) error {
	store, err := checkpoint.NewStore()
	if err != nil {
		return err
	}
	list, err := store.List()
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if len(list) == 0 {
		fmt.Fprintln(out, "No cleanup checkpoints under ~/.gotr/cache/cleanup-attachments/")
		return nil
	}
	t := ui.NewTable(cmd)
	t.AppendHeader(table.Row{"RUN_ID", "STARTED", "UPDATED", "TOTAL", "DONE", "PENDING", "FAILED", "TIMEOUT"})
	for _, s := range list {
		t.AppendRow(table.Row{
			s.RunID,
			s.StartedAt.Format("2006-01-02 15:04:05"),
			s.UpdatedAt.Format("2006-01-02 15:04:05"),
			s.Total,
			s.Done,
			s.Pending,
			s.Failed,
			s.Timeout,
		})
	}
	ui.Table(cmd, t)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Resume an interrupted run with: gotr attachments cleanup --resume <RUN_ID>")
	return nil
}

// buildPlanWithChunkingStatus drives BuildPlanChunked under the
// rich multiline progress UI (or per-project INFO logs when stderr is
// not a TTY / --quiet is set).
//nolint:gocyclo // Sequential pre/post-flight stages (config build + status spinner + post-summary) are clearer kept inline.
func buildPlanWithChunkingStatus(
	ctx context.Context,
	cmd *cobra.Command,
	lister cleanup.ProjectLister,
	scanner cleanup.AttachmentScanner,
	opts *cleanupOptions,
	filter cleanup.AttachmentFilter,
) (*cleanup.Plan, string, error) {
	quiet, _ := cmd.Flags().GetBool("quiet")
	store, err := checkpoint.NewStore()
	if err != nil {
		return nil, "", fmt.Errorf("checkpoint store: %w", err)
	}

	var ids []int64
	if !opts.AllProjects {
		ids = opts.ProjectIDs
	}

	stderr := cmd.ErrOrStderr()
	live := ui.NewMultilineStatus(ui.MultilineConfig{
		Writer: stderr,
		Quiet:  quiet,
	})
	var logSink io.Writer
	if !live.Active() && !quiet {
		logSink = stderr
	}
	progress := newScanProgressAdapter("Scanning attachments", scanner.Name(), live, logSink)

	cfg := cleanup.ChunkConfig{
		ChunkSize:             opts.ChunkSize,
		ScanTimeoutPerProject: opts.ScanTimeout,
		RunID:                 opts.ResumeRunID,
		Resume:                opts.ResumeRunID != "",
		Store:                 store,
		FilterSnapshot: checkpoint.FilterSnapshot{
			OlderThanRaw:          opts.OlderThanRaw,
			EntityTypes:           append([]string(nil), opts.EntityTypes...),
			Limit:                 opts.Limit,
			Concurrency:           opts.Concurrency,
			ScanStrategy:          opts.ScanStrategy,
			ScanTimeoutPerProject: opts.ScanTimeoutRaw,
		},
		AllProjects: opts.AllProjects,
		CLIArgs:     cliArgsFor(cmd),
		Progress:    progress,
		OnChunkComplete: func(idx, total int, partial *cleanup.Plan) {
			if quiet || logSink == nil {
				return
			}
			fmt.Fprintf(logSink, "INFO: chunk %d/%d done — running totals: %d projects with hits, %d attachments, %s\n",
				idx, total, len(partial.Projects), partial.TotalCount, humanBytes(partial.TotalBytes))
		},
	}

	if cfg.Resume {
		fmt.Fprintf(stderr, "INFO: resuming run %s\n", cfg.RunID)
	}

	stop := live.Start(ctx)
	plan, cp, runErr := cleanup.BuildPlanChunked(ctx, lister, scanner, ids, filter, opts.Concurrency, cfg)
	stop()

	runID := ""
	if cp != nil {
		runID = cp.RunID
	}
	if runErr != nil {
		if errors.Is(runErr, cleanup.ErrCheckpointMismatch) {
			return nil, runID, fmt.Errorf("resume: %w (delete the checkpoint dir under ~/.gotr/cache/cleanup-attachments/%s or rerun without --resume)", runErr, runID)
		}
		return nil, runID, runErr
	}

	if !cfg.Resume && runID != "" && !quiet {
		fmt.Fprintf(stderr, "INFO: run-id=%s — resume with: gotr attachments cleanup --resume %s\n", runID, runID)
	}

	// Surface failed/timeout summary so the operator can decide
	// whether to rerun with --resume.
	if cp != nil && !quiet {
		var failed, timedOut []int64
		for _, ps := range cp.Projects {
			switch ps.State {
			case checkpoint.StateFailed:
				failed = append(failed, ps.ID)
			case checkpoint.StateTimeout:
				timedOut = append(timedOut, ps.ID)
			}
		}
		if len(failed) > 0 || len(timedOut) > 0 {
			fmt.Fprintf(stderr, "WARN: scan completed with failures: %d failed %v, %d timed out %v — checkpoint preserved at ~/.gotr/cache/cleanup-attachments/%s\n",
				len(failed), failed, len(timedOut), timedOut, runID)
		}
	}

	return plan, runID, nil
}
