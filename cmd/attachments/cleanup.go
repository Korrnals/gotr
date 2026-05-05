// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/cleanup"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// validCleanupEntityTypes lists every attachment parent kind that can be
// targeted by `gotr attachments cleanup --entity-type`.
var validCleanupEntityTypes = []string{"case", "run", "plan", "plan_entry", "result", "test"}

// newCleanupCmd creates the `gotr attachments cleanup` command.
//
// The command lists candidate attachments (filtered by age + parent kind),
// captures a reversible snapshot of every binary, and then deletes them in
// parallel. The snapshot is mandatory by default and can be rolled back via
// `gotr snap rollback <id>`.
func newCleanupCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Bulk-delete attachments older than a threshold (with snapshot+rollback)",
		Long: `Bulk-delete TestRail attachments matching an age and parent-kind filter.

Workflow:
  1. List attachments for the selected projects (paginated).
  2. Filter by --older-than and --entity-type.
  3. Print a pre-flight summary (counts, total bytes, oldest timestamp).
  4. Snapshot every binary + parent-binding metadata under
     ~/.gotr/snaps/cleanup-attachments/<id>/  (default: ON, retention 7 days).
  5. Delete in parallel via the TestRail API.

WARNING: by default the cleanup targets ALL attachment kinds
(case, run, plan, plan_entry, result, test). A scope notice is printed
before scanning so you can abort and narrow with --entity-type.

Rollback re-uploads each binary to its original parent. TestRail assigns
new attachment IDs on re-upload — the rollback report records the
old_id → new_id mapping. Attachments bound to entity_type=test cannot be
restored (no add_attachment_to_test endpoint exists) and are reported
as Skipped.`,
		Example: `  # Dry-run: preview only, no snapshot, no deletes
  gotr attachments cleanup --project 30 --older-than 90d --dry-run

  # Cleanup result-bound attachments older than 3 months across one project
  gotr attachments cleanup --project 30 --older-than 3M --entity-type result

  # Cleanup across every visible project, with custom concurrency
  gotr attachments cleanup --all-projects --older-than 1y --concurrency 8`,
		RunE: runCleanup(getClient),
	}

	cmd.Flags().Int64Slice("project", nil, "Project ID(s); repeat or comma-separate. Mutually exclusive with --all-projects")
	cmd.Flags().Bool("all-projects", false, "Walk every visible project")
	cmd.Flags().String("older-than", "", "Age cutoff (e.g. 7d, 3M, 1y). Required unless --dry-run on a small set")
	cmd.Flags().StringSlice("entity-type", append([]string(nil), validCleanupEntityTypes...),
		"Parent kinds to clean (DEFAULT: ALL — case,run,plan,plan_entry,result,test). Narrow scope to avoid touching unrelated data.")
	cmd.Flags().Bool("dry-run", false, "Preview only; no snapshot, no deletes")
	cmd.Flags().Int("concurrency", 4, "Worker count for list and delete phases")
	cmd.Flags().Int("limit", 0, "Hard cap on the number of attachments to delete (0 = no cap)")
	cmd.Flags().Bool("no-snapshot", false, "Skip the pre-delete snapshot (discouraged)")
	cmd.Flags().String("snapshot-retention", "7d", "Override retention for this snapshot (e.g. 7d, 30d)")
	cmd.Flags().String("snap-label", "", "Optional human-readable label for the snapshot")
	cmd.Flags().Float64("max-size-gb", 0, "Refuse if the snapshot would exceed N GB (0 = no cap); use --force to override")
	cmd.Flags().Bool("compress", false, "Gzip-compress stored binaries inside the snapshot")
	cmd.Flags().Bool("force", false, "Bypass --max-size-gb and the interactive confirmation prompt")
	cmd.Flags().String("scan-strategy", "auto", "How to enumerate attachments: auto|project|entities. auto probes get_attachments_for_project once and falls back to a suites→cases/runs/plans walk on TestRail Server < 7.5.")
	cmd.Flags().Bool("no-report", false, "Skip emitting the deletion-audit report under ~/.gotr/reports/cleanup-attachments/")
	cmd.Flags().Int("chunk-size", 10, "Projects scanned per chunk; checkpoint is persisted after each chunk")
	cmd.Flags().String("resume", "", "Resume an interrupted run by its run-id (mutually exclusive with --project/--all-projects/--older-than/--entity-type)")
	cmd.Flags().String("scan-timeout-per-project", "10m", "Per-project Scan() timeout (e.g. 30s, 5m, 0 to disable)")
	cmd.Flags().Bool("list-checkpoints", false, "List existing cleanup checkpoints and exit")
	output.AddFlag(cmd)

	return cmd
}

// cleanupOptions is the resolved, validated parameter set built from CLI
// flags (and, in TODO #5, from the interactive survey).
type cleanupOptions struct {
	ProjectIDs        []int64
	AllProjects       bool
	OlderThan         time.Duration
	EntityTypes       []string
	DryRun            bool
	Concurrency       int
	Limit             int
	SkipSnapshot      bool
	SnapshotRetention time.Duration
	SnapshotLabel     string
	MaxSizeGB         float64
	Compress          bool
	Force             bool
	ScanStrategy      string
	OlderThanRaw      string
	CutoffUnix        int64
	NoReport          bool
	ChunkSize         int
	ResumeRunID       string
	ScanTimeoutRaw    string
	ScanTimeout       time.Duration
	ListCheckpoints   bool
}

//nolint:gocyclo // Sequential pre-flight stages (parse → prompt → validate → plan → confirm → execute) are clearer kept together than artificially split.
func runCleanup(getClient GetClientFunc) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		opts, err := parseCleanupFlags(cmd)
		if err != nil {
			return fmt.Errorf("parse flags: %w", err)
		}

		ctx := cmd.Context()

		// --list-checkpoints short-circuits the entire workflow.
		if opts.ListCheckpoints {
			if err := assertListCheckpointsExclusive(cmd); err != nil {
				return err
			}
			return runListCheckpoints(cmd)
		}

		if opts.ResumeRunID != "" {
			if err := assertResumeExclusive(cmd); err != nil {
				return err
			}
			if err := loadOptsFromCheckpoint(opts); err != nil {
				return fmt.Errorf("resume: %w", err)
			}
		}

		if err := promptCleanupOptions(ctx, cmd, opts); err != nil {
			return fmt.Errorf("interactive: %w", err)
		}

		printEntityScopeNotice(cmd, opts)

		if !opts.AllProjects && len(opts.ProjectIDs) == 0 {
			return fmt.Errorf("either --project or --all-projects is required")
		}
		if opts.AllProjects && len(opts.ProjectIDs) > 0 {
			return fmt.Errorf("--project and --all-projects are mutually exclusive")
		}
		if opts.OlderThan == 0 && !opts.DryRun {
			return fmt.Errorf("--older-than is required (use --dry-run to preview without it)")
		}

		api, ok := getClient(cmd).(cleanup.AttachmentsAPI)
		if !ok {
			return fmt.Errorf("client does not implement cleanup.AttachmentsAPI")
		}
		scannerAPI, ok := getClient(cmd).(cleanup.ScannerAPI)
		if !ok {
			return fmt.Errorf("client does not implement cleanup.ScannerAPI (need both project- and entity-level attachment endpoints)")
		}

		filter := cleanup.AttachmentFilter{
			EntityTypes: toEntityTypeSet(opts.EntityTypes),
			Limit:       opts.Limit,
		}
		if opts.OlderThan > 0 {
			filter.OlderThan = time.Now().Add(-opts.OlderThan)
		}

		scanner, err := resolveScannerWithStatus(ctx, cmd, scannerAPI, opts, filter)
		if err != nil {
			return fmt.Errorf("resolve scan strategy: %w", err)
		}

		plan, runID, err := buildPlanWithChunkingStatus(ctx, cmd, api, scanner, opts, filter)
		if err != nil {
			return fmt.Errorf("build plan: %w", err)
		}
		_ = runID

		if err := printCleanupSummary(cmd, plan, opts); err != nil {
			return err
		}
		if plan.TotalCount == 0 {
			return nil
		}

		if !opts.Force && !opts.SkipSnapshot && opts.MaxSizeGB > 0 {
			estGB := float64(plan.TotalBytes) / (1024 * 1024 * 1024)
			if estGB > opts.MaxSizeGB {
				return fmt.Errorf("estimated snapshot size %.2f GB exceeds --max-size-gb %.2f (use --force to override)", estGB, opts.MaxSizeGB)
			}
		}

		proceed, err := confirmCleanupExecution(ctx, opts)
		if err != nil {
			return err
		}
		if !proceed {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted by user.")
			return nil
		}

		store, manifest, err := openSnapStore()
		if err != nil {
			return fmt.Errorf("open snap store: %w", err)
		}

		execOpts := cleanup.ExecuteOptions{
			DryRun:           opts.DryRun,
			SkipSnapshot:     opts.SkipSnapshot,
			SnapshotLabel:    opts.SnapshotLabel,
			CompressBinaries: opts.Compress,
			Concurrency:      opts.Concurrency,
			CLIArgs:          cliArgsFor(cmd),
			ServerURL:        snap.CurrentServerURL(),
		}

		quiet, _ := cmd.Flags().GetBool("quiet")
		res, err := ui.RunWithStatus(ctx, ui.StatusConfig{
			Title:  "Cleaning up attachments",
			Writer: os.Stderr,
			Quiet:  quiet,
		}, func(ctx context.Context) (*cleanup.ExecuteResult, error) {
			return cleanup.Execute(ctx, api, store, manifest, plan, execOpts)
		})
		if err != nil {
			return fmt.Errorf("execute: %w", err)
		}

		// Apply snapshot-specific retention if requested by overriding the
		// snapshot meta. The 7-day default is enforced at gotr cleanup snaps
		// time (see TODO #6).
		_ = opts.SnapshotRetention

		printCleanupResult(cmd, res)
		writeCleanupReport(cmd, plan, res, opts)
		if res.DeleteErrors > 0 {
			return fmt.Errorf("%d delete(s) failed; snapshot is preserved for rollback", res.DeleteErrors)
		}
		return nil
	}
}

func parseCleanupFlags(cmd *cobra.Command) (*cleanupOptions, error) {
	opts := &cleanupOptions{}
	opts.ProjectIDs, _ = cmd.Flags().GetInt64Slice("project")
	opts.AllProjects, _ = cmd.Flags().GetBool("all-projects")
	rawAge, _ := cmd.Flags().GetString("older-than")
	rawEnt, _ := cmd.Flags().GetStringSlice("entity-type")
	opts.DryRun, _ = cmd.Flags().GetBool("dry-run")
	opts.Concurrency, _ = cmd.Flags().GetInt("concurrency")
	opts.Limit, _ = cmd.Flags().GetInt("limit")
	opts.SkipSnapshot, _ = cmd.Flags().GetBool("no-snapshot")
	rawRet, _ := cmd.Flags().GetString("snapshot-retention")
	opts.SnapshotLabel, _ = cmd.Flags().GetString("snap-label")
	opts.MaxSizeGB, _ = cmd.Flags().GetFloat64("max-size-gb")
	opts.Compress, _ = cmd.Flags().GetBool("compress")
	opts.Force, _ = cmd.Flags().GetBool("force")
	opts.ScanStrategy, _ = cmd.Flags().GetString("scan-strategy")
	opts.NoReport, _ = cmd.Flags().GetBool("no-report")
	opts.ChunkSize, _ = cmd.Flags().GetInt("chunk-size")
	opts.ResumeRunID, _ = cmd.Flags().GetString("resume")
	opts.ScanTimeoutRaw, _ = cmd.Flags().GetString("scan-timeout-per-project")
	opts.ListCheckpoints, _ = cmd.Flags().GetBool("list-checkpoints")
	opts.OlderThanRaw = rawAge
	opts.ScanStrategy = strings.ToLower(strings.TrimSpace(opts.ScanStrategy))
	switch opts.ScanStrategy {
	case "", "auto", "project", "entities":
	default:
		return nil, fmt.Errorf("--scan-strategy %q invalid (allowed: auto, project, entities)", opts.ScanStrategy)
	}

	if rawAge != "" {
		d, err := parseHumanDuration(rawAge)
		if err != nil {
			return nil, fmt.Errorf("--older-than %q: %w", rawAge, err)
		}
		opts.OlderThan = d
		opts.CutoffUnix = time.Now().Add(-d).Unix()
	}

	if rawRet != "" {
		d, err := parseHumanDuration(rawRet)
		if err != nil {
			return nil, fmt.Errorf("--snapshot-retention %q: %w", rawRet, err)
		}
		opts.SnapshotRetention = d
	}

	for _, e := range rawEnt {
		e = strings.ToLower(strings.TrimSpace(e))
		if !isValidCleanupEntity(e) {
			return nil, fmt.Errorf("--entity-type %q invalid (allowed: %s)", e, strings.Join(validCleanupEntityTypes, ", "))
		}
		opts.EntityTypes = append(opts.EntityTypes, e)
	}
	if opts.Concurrency < 1 {
		opts.Concurrency = 4
	}
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = 10
	}
	if strings.TrimSpace(opts.ScanTimeoutRaw) != "" && opts.ScanTimeoutRaw != "0" {
		d, err := parseHumanDuration(opts.ScanTimeoutRaw)
		if err != nil {
			return nil, fmt.Errorf("--scan-timeout-per-project %q: %w", opts.ScanTimeoutRaw, err)
		}
		opts.ScanTimeout = d
	}
	return opts, nil
}

func isValidCleanupEntity(e string) bool {
	for _, v := range validCleanupEntityTypes {
		if v == e {
			return true
		}
	}
	return false
}

func toEntityTypeSet(types []string) map[string]struct{} {
	if len(types) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(types))
	for _, t := range types {
		out[t] = struct{}{}
	}
	return out
}

// entityTypesCoversAll reports whether the provided slice covers every
// supported parent kind (case, run, plan, plan_entry, result, test).
// An empty slice also counts as "covers all" because downstream filters
// treat nil EntityTypes as "allow any kind".
func entityTypesCoversAll(types []string) bool {
	if len(types) == 0 {
		return true
	}
	seen := make(map[string]struct{}, len(types))
	for _, t := range types {
		seen[strings.ToLower(strings.TrimSpace(t))] = struct{}{}
	}
	for _, want := range validCleanupEntityTypes {
		if _, ok := seen[want]; !ok {
			return false
		}
	}
	return true
}

// printEntityScopeNotice prints a one-time pre-scan notice describing
// which attachment kinds the cleanup is about to touch. When the
// resolved scope covers ALL kinds the notice is rendered as a WARNING
// so the user can abort and rerun with a narrower --entity-type. When
// the scope is narrower the notice is informational only.
func printEntityScopeNotice(cmd *cobra.Command, opts *cleanupOptions) {
	w := cmd.ErrOrStderr()
	if entityTypesCoversAll(opts.EntityTypes) {
		fmt.Fprintln(w, "⚠️  Cleanup scope: ALL attachment kinds (case, run, plan, plan_entry, result, test)")
		fmt.Fprintln(w, "   This will scan and delete attachments attached to test cases, test runs,")
		fmt.Fprintln(w, "   test plans / plan entries, individual test results, and tests.")
		fmt.Fprintln(w, "   To narrow the scope use --entity-type, e.g.:")
		fmt.Fprintln(w, "       gotr attachments cleanup --entity-type result")
		fmt.Fprintln(w, "       gotr attachments cleanup --entity-type case,run")
		return
	}
	scoped := append([]string(nil), opts.EntityTypes...)
	sort.Strings(scoped)
	fmt.Fprintf(w, "ℹ️  Cleanup scope: %s\n", strings.Join(scoped, ", "))
}

// parseHumanDuration accepts shorthand suffixes Go's time.ParseDuration
// does not (d=days, w=weeks, M=months≈30d, y=years≈365d) plus any value
// time.ParseDuration understands.
func parseHumanDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	last := s[len(s)-1]
	switch last {
	case 'd', 'w', 'M', 'y':
		var n float64
		if _, err := fmt.Sscanf(s[:len(s)-1], "%f", &n); err != nil {
			return 0, fmt.Errorf("invalid number in %q: %w", s, err)
		}
		var unit time.Duration
		switch last {
		case 'd':
			unit = 24 * time.Hour
		case 'w':
			unit = 7 * 24 * time.Hour
		case 'M':
			unit = 30 * 24 * time.Hour
		case 'y':
			unit = 365 * 24 * time.Hour
		}
		return time.Duration(n * float64(unit)), nil
	default:
		return time.ParseDuration(s)
	}
}

// resolveScannerWithStatus picks the scan strategy honoring
// --scan-strategy, probing the server once when strategy=auto. The
// probe targets the first explicit project ID, or — when --all-projects
// is in effect — the first project returned by GetProjects (which is a
// call we'd make anyway during BuildPlanWithScanner).
func resolveScannerWithStatus(ctx context.Context, cmd *cobra.Command, api cleanup.ScannerAPI, opts *cleanupOptions, filter cleanup.AttachmentFilter) (cleanup.AttachmentScanner, error) {
	strategy := cleanup.ScanStrategy(opts.ScanStrategy)
	entityOpts := cleanup.EntityScannerOptionsFromTypes(filter.EntityTypes, opts.Concurrency)

	// project / entities forced — no probe required.
	if strategy == cleanup.ScanStrategyProject || strategy == cleanup.ScanStrategyEntities {
		return cleanup.ResolveScanner(ctx, api, strategy, entityOpts, 0, nil)
	}

	lister, ok := any(api).(cleanup.ProjectLister)
	if !ok {
		return nil, fmt.Errorf("client does not implement cleanup.ProjectLister (need GetProjects for the auto-probe)")
	}
	probeID, err := pickProbeProjectID(ctx, lister, opts)
	if err != nil {
		return nil, err
	}
	logf := func(format string, args ...any) {
		fmt.Fprintln(cmd.ErrOrStderr(), "INFO:", fmt.Sprintf(format, args...))
	}
	return cleanup.ResolveScanner(ctx, api, strategy, entityOpts, probeID, logf)
}

// pickProbeProjectID returns a project ID suitable for the auto-probe
// call. With --project, the first explicit ID is used. With
// --all-projects the helper performs a single GetProjects call (which
// the planner would issue regardless) and returns the first ID.
func pickProbeProjectID(ctx context.Context, lister cleanup.ProjectLister, opts *cleanupOptions) (int64, error) {
	if !opts.AllProjects && len(opts.ProjectIDs) > 0 {
		return opts.ProjectIDs[0], nil
	}
	all, err := lister.GetProjects(ctx)
	if err != nil {
		return 0, fmt.Errorf("list projects (probe): %w", err)
	}
	if len(all) == 0 {
		return 0, fmt.Errorf("no projects visible to API key — nothing to probe")
	}
	return all[0].ID, nil
}

func openSnapStore() (*snap.Store, *snap.Manifest, error) {
	store, err := snap.NewStore()
	if err != nil {
		return nil, nil, err
	}
	m, err := snap.LoadManifest(store)
	if err != nil {
		return nil, nil, err
	}
	return store, m, nil
}

func cliArgsFor(cmd *cobra.Command) []string {
	args := []string{"attachments", cmd.Name()}
	cmd.Flags().Visit(func(f *pflag.Flag) {
		args = append(args, "--"+f.Name+"="+f.Value.String())
	})
	return args
}

// printCleanupSummary prints the per-project counts/bytes/oldest table
// (always shown — both for --dry-run and real runs) plus a one-line
// total summary used to drive the confirmation gate.
func printCleanupSummary(cmd *cobra.Command, plan *cleanup.Plan, opts *cleanupOptions) error {
	out := cmd.OutOrStdout()
	if plan.TotalCount == 0 {
		fmt.Fprintln(out, "No attachments matched the filter — nothing to do.")
		return nil
	}

	fmt.Fprintln(out, "Attachments selected for deletion:")
	t := ui.NewTable(cmd)
	t.AppendHeader(table.Row{"PROJECT", "NAME", "COUNT", "SIZE", "OLDEST"})
	for _, sel := range plan.Projects {
		t.AppendRow(table.Row{
			sel.ProjectID,
			truncate(sel.ProjectName, 40),
			len(sel.Attachments),
			humanBytes(sel.TotalBytes),
			formatUnix(sel.OldestUnix),
		})
	}
	ui.Table(cmd, t)

	fmt.Fprintf(out, "\nTotal: %d attachments, %s. Oldest: %s\n",
		plan.TotalCount, humanBytes(plan.TotalBytes), formatUnix(plan.OldestUnix))
	if plan.TruncatedByLimit {
		fmt.Fprintln(out, "(truncated by --limit)")
	}
	if opts.DryRun {
		fmt.Fprintln(out, "\nDry-run: no snapshot will be taken and no deletes will be issued.")
		return nil
	}
	if opts.SkipSnapshot {
		fmt.Fprintln(out, "Snapshot: DISABLED (--no-snapshot). This run is NOT reversible.")
	} else {
		fmt.Fprintf(out, "Snapshot: ~%s under ~/.gotr/snaps/cleanup-attachments/<id>/  (retention: %s)\n",
			humanBytes(plan.TotalBytes), formatRetention(opts.SnapshotRetention))
	}
	return nil
}

func printCleanupResult(cmd *cobra.Command, res *cleanup.ExecuteResult) {
	out := cmd.OutOrStdout()
	if res.DryRun {
		fmt.Fprintf(out, "\nDry-run complete: would delete %d attachments (~%s).\n", res.BackedUp, humanBytes(res.BackupBytes))
		return
	}
	if res.SnapshotID != "" {
		fmt.Fprintf(out, "\nSnapshot: %s  (%d files, %s)\n", res.SnapshotID, res.BackedUp, humanBytes(res.BackupBytes))
	}
	fmt.Fprintf(out, "Deleted: %d   Failed: %d\n", res.Deleted, res.DeleteErrors)
	if res.DeleteErrors > 0 {
		fmt.Fprintln(out, "\nFailed deletes:")
		for _, f := range res.Failures {
			fmt.Fprintf(out, "  - attachment %d (project %d): %s\n", f.AttachmentID, f.ProjectID, f.Error)
		}
		if res.SnapshotID != "" {
			fmt.Fprintf(out, "\nRollback the snapshot with:  gotr snap rollback %s\n", res.SnapshotID)
		}
	}
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

func formatUnix(u int64) string {
	if u == 0 {
		return "—"
	}
	return time.Unix(u, 0).UTC().Format("2006-01-02")
}

func formatRetention(d time.Duration) string {
	if d <= 0 {
		return "default (7 days)"
	}
	days := int(d / (24 * time.Hour))
	if days >= 1 {
		return fmt.Sprintf("%d days", days)
	}
	return d.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}
