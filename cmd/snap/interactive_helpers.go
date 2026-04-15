package snap

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/Korrnals/gotr/internal/interactive"
	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/spf13/cobra"
)

// backOption is the label for "go back" in multi-level pickers.
const backOption = "← Back"

// exitOption is the label for exiting interactive mode.
const exitOption = "✕ Exit"

// errGoBack is a sentinel to signal user chose "← Back".
var errGoBack = fmt.Errorf("go back")

// errExit is a sentinel to signal user chose "✕ Exit".
var errExit = fmt.Errorf("exit")

// serverLabel returns a normalised display string for a server URL.
// Strips API path, shows only scheme+host. Empty → "(legacy)" for old snapshots.
func serverLabel(rawURL string) string {
	if rawURL == "" {
		return "(legacy)"
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		// Not a parseable URL — return as-is (e.g. "test").
		return rawURL
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}

// entityIDsLabel formats entity IDs for display in picker labels.
func entityIDsLabel(ids []int64) string {
	if len(ids) == 0 {
		return ""
	}
	if len(ids) == 1 {
		return fmt.Sprintf(" #%d", ids[0])
	}
	if len(ids) <= 3 {
		parts := make([]string, len(ids))
		for i, id := range ids {
			parts[i] = fmt.Sprintf("%d", id)
		}
		return fmt.Sprintf(" #%s", strings.Join(parts, ","))
	}
	return fmt.Sprintf(" #%d,…(+%d)", ids[0], len(ids)-1)
}

// isInterruptError returns true if the error is caused by user pressing Ctrl+C.
func isInterruptError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "interrupt")
}

// wrapInterrupt converts survey interrupt errors to context.Canceled
// so root.go handles them as graceful cancellation.
func wrapInterrupt(err error) error {
	if isInterruptError(err) {
		return context.Canceled
	}
	return err
}

// ---------------------------------------------------------------------------
// Picker label formatting — aligned columns
// ---------------------------------------------------------------------------

// shortID extracts a short unique identifier from a full snapshot ID.
// E.g. "cases/20260413T194322_add_bulk_0" → "20260413T194322_add_bulk_0".
func shortID(id string) string {
	if idx := strings.LastIndex(id, "/"); idx >= 0 {
		return id[idx+1:]
	}
	return id
}

// formatPickerLabel creates an aligned display label for a snapshot in pickers.
func formatPickerLabel(idx int, e snaplib.ManifestEntry) string {
	var b strings.Builder

	fmt.Fprintf(&b, "[%d] %s %s", idx, e.Operation, e.EntityType)
	b.WriteString(entityIDsLabel(e.EntityIDs))

	if e.Name != "" {
		fmt.Fprintf(&b, " %q", e.Name)
	}

	fmt.Fprintf(&b, " | %s | T%d | %s | %s",
		e.Status, e.RollbackTier,
		e.Timestamp.Format("2006-01-02 15:04"),
		shortID(e.ID))

	return b.String()
}

// alignedPickerLabels returns picker labels with consistent column widths.
// Includes a header row as the first element.
func alignedPickerLabels(entries []snaplib.ManifestEntry) (header string, labels []string) {
	type row struct {
		idx      string
		opEntity string
		ids      string
		cat      string
		name     string
		label    string
		status   string
		tier     string
		ts       string
		snapID   string
	}

	rows := make([]row, len(entries))
	maxIdx, maxOp, maxIDs, maxCat, maxName, maxLabel, maxStatus, maxTier, maxTs, maxSnap := 0, 0, 0, 0, 0, 0, 0, 0, 0, 0

	// Header widths.
	hIdx, hOp, hIDs, hCat, hName := "#", "OPERATION", "IDS", "CATEGORY", ""
	hLabel := ""
	hStatus, hTier, hTs, hSnap := "STATUS", "TIER", "DATE", "SNAPSHOT ID"

	for i, e := range entries {
		idsLabel := strings.TrimSpace(entityIDsLabel(e.EntityIDs))
		if idsLabel == "" {
			idsLabel = "–"
		}
		r := row{
			idx:      fmt.Sprintf("[%d]", i+1),
			opEntity: fmt.Sprintf("%s %s", e.Operation, e.EntityType),
			ids:      idsLabel,
			cat:      string(e.Category),
			status:   string(e.Status),
			tier:     fmt.Sprintf("T%d", e.RollbackTier),
			ts:       e.Timestamp.Format("2006-01-02 15:04"),
			snapID:   shortID(e.ID),
		}
		if e.Name != "" {
			r.name = fmt.Sprintf("%q", e.Name)
		}
		if e.Label != "" {
			r.label = fmt.Sprintf("🏷 %s", e.Label)
		}
		rows[i] = r

		if len(r.idx) > maxIdx { maxIdx = len(r.idx) }
		if len(r.opEntity) > maxOp { maxOp = len(r.opEntity) }
		if len(r.ids) > maxIDs { maxIDs = len(r.ids) }
		if len(r.cat) > maxCat { maxCat = len(r.cat) }
		if len(r.name) > maxName { maxName = len(r.name) }
		if len(r.label) > maxLabel { maxLabel = len(r.label) }
		if len(r.status) > maxStatus { maxStatus = len(r.status) }
		if len(r.tier) > maxTier { maxTier = len(r.tier) }
		if len(r.ts) > maxTs { maxTs = len(r.ts) }
		if len(r.snapID) > maxSnap { maxSnap = len(r.snapID) }
	}

	// Ensure header labels don't shrink columns.
	if len(hIdx) > maxIdx { maxIdx = len(hIdx) }
	if len(hOp) > maxOp { maxOp = len(hOp) }
	if len(hIDs) > maxIDs { maxIDs = len(hIDs) }
	if len(hCat) > maxCat { maxCat = len(hCat) }
	if len(hLabel) > maxLabel { maxLabel = len(hLabel) }
	if len(hStatus) > maxStatus { maxStatus = len(hStatus) }
	if len(hTier) > maxTier { maxTier = len(hTier) }
	if len(hTs) > maxTs { maxTs = len(hTs) }
	if len(hSnap) > maxSnap { maxSnap = len(hSnap) }

	// Build format function — all columns separated by │.
	fmtRow := func(idx, op, ids, cat, name, label, status, tier, ts, snap string) string {
		var b strings.Builder
		fmt.Fprintf(&b, "%-*s %-*s │ %-*s │ %-*s",
			maxIdx, idx, maxOp, op, maxIDs, ids, maxCat, cat)
		if maxName > 0 {
			fmt.Fprintf(&b, " │ %-*s", maxName, name)
		}
		if maxLabel > 0 {
			fmt.Fprintf(&b, " │ %-*s", maxLabel, label)
		}
		fmt.Fprintf(&b, " │ %-*s │ %-*s │ %-*s │ %-*s",
			maxStatus, status, maxTier, tier, maxTs, ts, maxSnap, snap)
		return b.String()
	}

	// Header.
	header = fmtRow(hIdx, hOp, hIDs, hCat, hName, hLabel, hStatus, hTier, hTs, hSnap)

	// Data rows.
	labels = make([]string, len(entries))
	for i, r := range rows {
		labels[i] = fmtRow(r.idx, r.opEntity, r.ids, r.cat, r.name, r.label, r.status, r.tier, r.ts, r.snapID)
	}
	return header, labels
}

// ---------------------------------------------------------------------------
// Operation grouping helpers
// ---------------------------------------------------------------------------

// operationGroup pairs an operation label with its snapshot entries.
type operationGroup struct {
	Label   string
	Entries []snaplib.ManifestEntry
}

// groupByOperation groups entries by Operation, sorted alphabetically.
func groupByOperation(entries []snaplib.ManifestEntry) []operationGroup {
	m := make(map[snaplib.Operation][]snaplib.ManifestEntry)
	for _, e := range entries {
		m[e.Operation] = append(m[e.Operation], e)
	}

	groups := make([]operationGroup, 0, len(m))
	for op, ents := range m {
		groups = append(groups, operationGroup{Label: string(op), Entries: ents})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Label < groups[j].Label
	})
	return groups
}

// ---------------------------------------------------------------------------
// Category (resource) grouping helpers
// ---------------------------------------------------------------------------

// categoryGroup pairs a category label with its snapshot entries.
type categoryGroup struct {
	Label   string
	Entries []snaplib.ManifestEntry
}

// groupByCategory groups entries by Category, sorted alphabetically.
func groupByCategory(entries []snaplib.ManifestEntry) []categoryGroup {
	m := make(map[snaplib.Category][]snaplib.ManifestEntry)
	for _, e := range entries {
		m[e.Category] = append(m[e.Category], e)
	}

	groups := make([]categoryGroup, 0, len(m))
	for cat, ents := range m {
		groups = append(groups, categoryGroup{Label: string(cat), Entries: ents})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Label < groups[j].Label
	})
	return groups
}

// ---------------------------------------------------------------------------
// Three-level picker: server → operation → snapshot
// ---------------------------------------------------------------------------

// pickerOpts configures selectSnapshot behaviour.
type pickerOpts struct {
	// statusFilter limits which statuses are selectable (nil = all).
	statusFilter []snaplib.Status
}

// selectSnapshot shows a multi-level interactive picker.
// Level 1: server (if >1) — Level 2: operation — Level 3: snapshot.
// Supports "← Back" navigation between levels.
func selectSnapshot(ctx context.Context, manifest *snaplib.Manifest, prompt string, opts *pickerOpts) (string, error) {
	entries := manifest.All()
	if len(entries) == 0 {
		return "", fmt.Errorf("no snapshots found")
	}

	// Apply status filter.
	if opts != nil && len(opts.statusFilter) > 0 {
		allowed := make(map[snaplib.Status]bool, len(opts.statusFilter))
		for _, s := range opts.statusFilter {
			allowed[s] = true
		}
		filtered := entries[:0]
		for _, e := range entries {
			if allowed[e.Status] {
				filtered = append(filtered, e)
			}
		}
		if len(filtered) == 0 {
			return "", fmt.Errorf("no snapshots with status %v found", opts.statusFilter)
		}
		entries = filtered
	}

	p := interactive.PrompterFromContext(ctx)
	groups := groupByServer(entries)

	for { // Server-level loop (allows back navigation).
		var selected serverGroup

		if len(groups) == 1 {
			selected = groups[0]
		} else {
			options := make([]string, len(groups))
			for i, g := range groups {
				options[i] = fmt.Sprintf("%s — %d snapshots", serverLabel(g.URL), len(g.Entries))
			}

			sPrompt := fmt.Sprintf("Select server (%d servers, %d snapshots):", len(groups), len(entries))
			idx, _, err := p.Select(sPrompt, options)
			if err != nil {
				return "", wrapInterrupt(err)
			}
			selected = groups[idx]
		}

		// Operation-level loop.
		snapID, err := pickByOperation(p, selected.Entries, prompt, len(groups) > 1)
		if err == errGoBack {
			if len(groups) == 1 {
				// Can't go back from single-server.
				return "", context.Canceled
			}
			continue // Re-show server picker.
		}
		if err != nil {
			return "", err
		}
		return snapID, nil
	}
}

// pickByOperation shows operation → snapshot selection with back support.
func pickByOperation(p interactive.Prompter, entries []snaplib.ManifestEntry, prompt string, allowBack bool) (string, error) {
	opGroups := groupByOperation(entries)

	for { // Operation-level loop.
		var opEntries []snaplib.ManifestEntry

		if len(opGroups) == 1 {
			opEntries = opGroups[0].Entries
		} else {
			options := make([]string, 0, len(opGroups)+1)
			if allowBack {
				options = append(options, backOption)
			}
			for _, g := range opGroups {
				options = append(options, fmt.Sprintf("%-12s — %d snapshots", g.Label, len(g.Entries)))
			}

			qPrompt := fmt.Sprintf("Select operation (%d types, %d snapshots):", len(opGroups), len(entries))
			idx, _, err := p.Select(qPrompt, options)
			if err != nil {
				return "", wrapInterrupt(err)
			}

			if allowBack && idx == 0 {
				return "", errGoBack
			}
			offset := 0
			if allowBack {
				offset = 1
			}
			opEntries = opGroups[idx-offset].Entries
		}

		// Category-level selection.
		snapID, err := pickByCategory(p, opEntries, prompt, len(opGroups) > 1 || allowBack)
		if err == errGoBack {
			if len(opGroups) == 1 {
				if allowBack {
					return "", errGoBack // Propagate back to server level.
				}
				return "", context.Canceled
			}
			continue // Re-show operation picker.
		}
		if err != nil {
			return "", err
		}
		return snapID, nil
	}
}

// pickByCategory shows category → snapshot selection with back support.
// If only one category exists, the category picker is skipped.
func pickByCategory(p interactive.Prompter, entries []snaplib.ManifestEntry, prompt string, allowBack bool) (string, error) {
	catGroups := groupByCategory(entries)

	// Single category — skip picker.
	if len(catGroups) == 1 {
		return pickSnapshot(p, entries, prompt, allowBack)
	}

	for { // Category-level loop.
		options := make([]string, 0, len(catGroups)+1)
		if allowBack {
			options = append(options, backOption)
		}
		for _, g := range catGroups {
			options = append(options, fmt.Sprintf("%-16s — %d snapshots", g.Label, len(g.Entries)))
		}

		qPrompt := fmt.Sprintf("Select resource (%d types, %d snapshots):", len(catGroups), len(entries))
		idx, _, err := p.Select(qPrompt, options)
		if err != nil {
			return "", wrapInterrupt(err)
		}

		if allowBack && idx == 0 {
			return "", errGoBack
		}
		offset := 0
		if allowBack {
			offset = 1
		}
		catEntries := catGroups[idx-offset].Entries

		snapID, err := pickSnapshot(p, catEntries, prompt, true)
		if err == errGoBack {
			continue // Re-show category picker.
		}
		if err != nil {
			return "", err
		}
		return snapID, nil
	}
}

// pickSnapshot shows the final snapshot picker with aligned labels.
func pickSnapshot(p interactive.Prompter, entries []snaplib.ManifestEntry, prompt string, allowBack bool) (string, error) {
	_, labels := alignedPickerLabels(entries)

	// Build options: [← Back], data..., [← Back].
	options := make([]string, 0, len(labels)+2)
	if allowBack {
		options = append(options, backOption)
	}
	options = append(options, labels...)
	if allowBack {
		options = append(options, backOption)
	}

	base := strings.TrimRight(prompt, ": ")
	snapPrompt := fmt.Sprintf("%s (%d snapshots):", base, len(entries))

	idx, _, err := p.Select(snapPrompt, options)
	if err != nil {
		return "", wrapInterrupt(err)
	}

	// Back at top or bottom.
	if allowBack && (idx == 0 || idx == len(options)-1) {
		return "", errGoBack
	}

	// Data offset: skip back (if any).
	offset := 0
	if allowBack {
		offset = 1
	}
	return entries[idx-offset].ID, nil
}

// resolveSnapshotID returns the snapshot ID from args or interactive selection.
func resolveSnapshotID(ctx context.Context, args []string, manifest *snaplib.Manifest, prompt string) (string, error) {
	return resolveSnapshotIDWith(ctx, args, manifest, prompt, nil)
}

// resolveSnapshotIDWith returns the snapshot ID from args or interactive selection with filter options.
func resolveSnapshotIDWith(ctx context.Context, args []string, manifest *snaplib.Manifest, prompt string, opts *pickerOpts) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	if interactive.IsNonInteractive(ctx) {
		return "", fmt.Errorf("snapshot_id required in non-interactive mode")
	}

	return selectSnapshot(ctx, manifest, prompt, opts)
}

// ---------------------------------------------------------------------------
// Browse mode: server → operation → snapshot → card → loop back to snapshot
// Used by `list` and `info` interactive mode.
// ---------------------------------------------------------------------------

// browseSnapshots opens a full interactive browser with three-level navigation.
// At the snapshot level, selecting a snapshot shows its info card and loops back
// to the same snapshot list. User navigates with ← Back / ✕ Exit.
func browseSnapshots(cmd *cobra.Command, store *snaplib.Store, manifest *snaplib.Manifest) error {
	entries := manifest.All()
	if len(entries) == 0 {
		return fmt.Errorf("no snapshots found")
	}

	p := interactive.PrompterFromContext(cmd.Context())
	groups := groupByServer(entries)

	for { // Server-level loop.
		var selected serverGroup

		if len(groups) == 1 {
			selected = groups[0]
		} else {
			options := make([]string, 0, len(groups)+1)
			options = append(options, exitOption)
			for _, g := range groups {
				options = append(options, fmt.Sprintf("%s — %d snapshots", serverLabel(g.URL), len(g.Entries)))
			}

			sPrompt := fmt.Sprintf("Select server (%d servers, %d snapshots):", len(groups), len(entries))
			idx, _, err := p.Select(sPrompt, options)
			if err != nil {
				return nil // Ctrl+C → graceful exit
			}
			if idx == 0 {
				return nil // ✕ Exit
			}
			selected = groups[idx-1]
			fmt.Fprintf(cmd.OutOrStdout(), "  ✓ %s\n", serverLabel(selected.URL))
		}

		err := browseByOperation(cmd, store, p, selected.Entries, len(groups) > 1)
		if err == errGoBack {
			if len(groups) == 1 {
				return nil
			}
			continue
		}
		if err == errExit {
			return nil
		}
		if err != nil {
			return nil
		}
	}
}

// browseByOperation shows operation picker, then delegates to browseSnapList.
func browseByOperation(cmd *cobra.Command, store *snaplib.Store, p interactive.Prompter, entries []snaplib.ManifestEntry, allowBack bool) error {
	opGroups := groupByOperation(entries)

	for { // Operation-level loop.
		var opEntries []snaplib.ManifestEntry

		if len(opGroups) == 1 {
			opEntries = opGroups[0].Entries
		} else {
			options := make([]string, 0, len(opGroups)+2)
			if allowBack {
				options = append(options, backOption)
			}
			options = append(options, exitOption)
			for _, g := range opGroups {
				options = append(options, fmt.Sprintf("%-12s — %d snapshots", g.Label, len(g.Entries)))
			}

			qPrompt := fmt.Sprintf("Select operation (%d types, %d snapshots):", len(opGroups), len(entries))
			idx, _, err := p.Select(qPrompt, options)
			if err != nil {
				return errExit // Ctrl+C
			}

			if allowBack && idx == 0 {
				return errGoBack
			}
			exitIdx := 0
			if allowBack {
				exitIdx = 1
			}
			if idx == exitIdx {
				return errExit
			}
			offset := exitIdx + 1
			opEntries = opGroups[idx-offset].Entries
			fmt.Fprintf(cmd.OutOrStdout(), "  ✓ %s\n", opGroups[idx-offset].Label)
		}

		err := browseByCategory(cmd, store, p, opEntries, len(opGroups) > 1 || allowBack)
		if err == errGoBack {
			if len(opGroups) == 1 {
				if allowBack {
					return errGoBack
				}
				return nil
			}
			continue
		}
		if err == errExit {
			return errExit
		}
		if err != nil {
			return nil
		}
	}
}

// browseByCategory shows category picker, then delegates to browseSnapList.
// If only one category exists, the category picker is skipped.
func browseByCategory(cmd *cobra.Command, store *snaplib.Store, p interactive.Prompter, entries []snaplib.ManifestEntry, allowBack bool) error {
	catGroups := groupByCategory(entries)

	// Single category — skip picker.
	if len(catGroups) == 1 {
		return browseSnapList(cmd, store, p, entries, allowBack)
	}

	for { // Category-level loop.
		options := make([]string, 0, len(catGroups)+2)
		if allowBack {
			options = append(options, backOption)
		}
		options = append(options, exitOption)
		for _, g := range catGroups {
			options = append(options, fmt.Sprintf("%-16s — %d snapshots", g.Label, len(g.Entries)))
		}

		qPrompt := fmt.Sprintf("Select resource (%d types, %d snapshots):", len(catGroups), len(entries))
		idx, _, err := p.Select(qPrompt, options)
		if err != nil {
			return errExit // Ctrl+C
		}

		if allowBack && idx == 0 {
			return errGoBack
		}
		exitIdx := 0
		if allowBack {
			exitIdx = 1
		}
		if idx == exitIdx {
			return errExit
		}
		offset := exitIdx + 1
		catEntries := catGroups[idx-offset].Entries
		fmt.Fprintf(cmd.OutOrStdout(), "  ✓ %s\n", catGroups[idx-offset].Label)

		err = browseSnapList(cmd, store, p, catEntries, true)
		if err == errGoBack {
			continue // Re-show category picker.
		}
		if err == errExit {
			return errExit
		}
		if err != nil {
			return nil
		}
	}
}

// browseSnapList shows the snapshot picker and loops: pick → view card → pick again.
func browseSnapList(cmd *cobra.Command, store *snaplib.Store, p interactive.Prompter, entries []snaplib.ManifestEntry, allowBack bool) error {
	_, labels := alignedPickerLabels(entries)

	for { // Snapshot browse loop — after viewing a card, returns here.
		options := make([]string, 0, len(labels)+3)
		if allowBack {
			options = append(options, backOption)
		}
		options = append(options, exitOption)
		options = append(options, labels...)
		if allowBack {
			options = append(options, backOption)
		}

		snapPrompt := fmt.Sprintf("Select snapshot to inspect (%d snapshots):", len(entries))

		idx, _, err := p.Select(snapPrompt, options)
		if err != nil {
			return errExit // Ctrl+C or mock exhausted → exit
		}

		// ← Back at top or bottom.
		if allowBack && (idx == 0 || idx == len(options)-1) {
			return errGoBack
		}

		// ✕ Exit.
		exitIdx := 0
		if allowBack {
			exitIdx = 1
		}
		if idx == exitIdx {
			return errExit
		}

		// Data offset: skip control options.
		offset := exitIdx + 1
		entry := entries[idx-offset]

		meta, err := store.LoadMeta(entry.ID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Error loading snapshot: %v\n\n", err)
			continue
		}

		renderInfoCard(cmd, meta)
		fmt.Fprintln(cmd.OutOrStdout())

		// Post-card action menu.
		action, err := postCardAction(cmd, p, meta)
		if err != nil {
			return errExit
		}
		switch action {
		case postActionBack:
			continue // back to snapshot list
		case postActionExit:
			return errExit
		case postActionRollback:
			executeRollbackFromBrowser(cmd, entry.ID)
			continue // back to snapshot list after rollback
		}
	}
}

// ---------------------------------------------------------------------------
// Post-card action menu
// ---------------------------------------------------------------------------

type postAction int

const (
	postActionBack postAction = iota
	postActionExit
	postActionRollback
)

// postCardAction shows a mini-menu after viewing a snapshot info card.
// Returns the chosen action.
func postCardAction(cmd *cobra.Command, p interactive.Prompter, meta *snaplib.Meta) (postAction, error) {
	out := cmd.OutOrStdout()
	options := []string{backOption, exitOption}

	// Show rollback option only for actionable statuses;
	// for terminal statuses print a hint instead.
	canRollback := meta.Status == snaplib.StatusAvailable || meta.Status == snaplib.StatusRollbackPartial
	if canRollback {
		options = append(options, "↻ Rollback this snapshot")
	} else if meta.Status == snaplib.StatusRolledBack {
		fmt.Fprintln(out, "  ✓ Snapshot already rolled back.")
	} else if meta.Status == snaplib.StatusExpired {
		fmt.Fprintln(out, "  ⚠ Snapshot expired — rollback unavailable.")
	}

	idx, _, err := p.Select("Action:", options)
	if err != nil {
		return postActionExit, err
	}

	switch idx {
	case 0:
		return postActionBack, nil
	case 1:
		return postActionExit, nil
	case 2:
		return postActionRollback, nil
	default:
		return postActionBack, nil
	}
}

// executeRollbackFromBrowser finds and runs the rollback subcommand for the given snapshot.
func executeRollbackFromBrowser(cmd *cobra.Command, snapID string) {
	parent := cmd.Parent()
	if parent == nil {
		fmt.Fprintf(cmd.OutOrStdout(), "💡 Run:  gotr snap rollback %s\n\n", snapID)
		return
	}

	var rollbackCmd *cobra.Command
	for _, c := range parent.Commands() {
		if c.Name() == "rollback" {
			rollbackCmd = c
			break
		}
	}
	if rollbackCmd == nil {
		fmt.Fprintf(cmd.OutOrStdout(), "💡 Run:  gotr snap rollback %s\n\n", snapID)
		return
	}

	rollbackCmd.SetContext(cmd.Context())
	rollbackCmd.SetOut(cmd.OutOrStdout())
	rollbackCmd.SetErr(cmd.ErrOrStderr())
	if err := rollbackCmd.RunE(rollbackCmd, []string{snapID}); err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "\nError: %v\n", err)
	}
	fmt.Fprintln(cmd.OutOrStdout())
}
