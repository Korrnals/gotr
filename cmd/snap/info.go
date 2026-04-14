package snap

import (
	"fmt"
	"io"
	"strings"

	"github.com/Korrnals/gotr/internal/interactive"
	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [snapshot_id]",
		Short: "Show snapshot details",
		Long: `Displays snapshot metadata as a formatted card.

If snapshot_id is omitted, shows an interactive picker.
Use --format json for machine-readable JSON output.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snaplib.NewStore()
			if err != nil {
				return err
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return err
			}

			// Non-interactive or explicit ID: show once and exit.
			if len(args) > 0 || interactive.IsNonInteractive(cmd.Context()) {
				snapID, err := resolveSnapshotID(cmd.Context(), args, manifest, "Select snapshot to inspect:")
				if err != nil {
					return err
				}

				meta, err := store.LoadMeta(snapID)
				if err != nil {
					return fmt.Errorf("snapshot %q not found: %w", snapID, err)
				}

				if ui.IsJSON(cmd) {
					return ui.JSON(cmd, meta)
				}

				renderInfoCard(cmd, meta)
				return nil
			}

			// Interactive browse: pick → view → back to list.
			return browseSnapshots(cmd, store, manifest)
		},
	}
}

// tierLabel returns a human-readable tier description.
func tierLabel(t snaplib.Tier) string {
	switch t {
	case snaplib.Tier1:
		return "T1 (full rollback)"
	case snaplib.Tier2:
		return "T2 (new ID on rollback)"
	case snaplib.Tier3:
		return "T3 (info only)"
	default:
		return fmt.Sprintf("T%d", t)
	}
}

// humanSize formats bytes into a readable string.
func humanSize(b int64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// renderInfoCard prints a structured snapshot info card.
func renderInfoCard(cmd *cobra.Command, meta *snaplib.Meta) {
	out := cmd.OutOrStdout()

	entityIDs := "-"
	if len(meta.EntityIDs) > 0 {
		ids := make([]string, len(meta.EntityIDs))
		for i, id := range meta.EntityIDs {
			ids[i] = fmt.Sprintf("%d", id)
		}
		entityIDs = strings.Join(ids, ", ")
	}

	// Main info table.
	t := ui.NewTable(cmd)
	t.SetTitle("Snapshot Info")
	t.AppendRows([]table.Row{
		{"ID", meta.ID},
		{"Server", serverLabel(meta.ServerURL)},
		{"Operation", fmt.Sprintf("%s %s", meta.Operation, meta.EntityType)},
		{"Category", meta.Category},
		{"Tier", tierLabel(meta.RollbackTier)},
		{"Status", meta.Status},
		{"Entity IDs", entityIDs},
		{"Project", meta.ProjectID},
		{"Suite", meta.SuiteID},
		{"CLI Command", meta.CLICommand},
		{"Created", meta.Timestamp.Format("2006-01-02 15:04:05 UTC")},
		{"Data Size", humanSize(meta.DataSizeBytes)},
	})
	if meta.Name != "" {
		t.AppendRow(table.Row{"Name", meta.Name})
	}
	fmt.Fprintln(out, t.Render())

	// Entities sub-table.
	if len(meta.Entities) > 0 {
		et := ui.NewTable(cmd)
		et.SetTitle(fmt.Sprintf("Entities (%d)", len(meta.Entities)))
		et.AppendHeader(table.Row{"#", "TYPE", "ID", "PARENT"})
		for i, e := range meta.Entities {
			parent := "-"
			if e.ParentID != 0 {
				parent = fmt.Sprintf("%d", e.ParentID)
			}
			et.AppendRow(table.Row{i + 1, e.Type, e.ID, parent})
		}
		fmt.Fprintln(out, et.Render())
	}

	// Rollback log sub-table.
	if len(meta.RollbackLog) > 0 {
		rt := ui.NewTable(cmd)
		rt.SetTitle(fmt.Sprintf("Rollback Log (%d)", len(meta.RollbackLog)))
		rt.AppendHeader(table.Row{"#", "TYPE", "ID", "STATUS", "ERROR"})
		for i, r := range meta.RollbackLog {
			errMsg := "-"
			if r.Error != "" {
				errMsg = r.Error
			}
			rt.AppendRow(table.Row{i + 1, r.Type, r.ID, r.Status, errMsg})
		}
		fmt.Fprintln(out, rt.Render())
	}

	// Status summary.
	renderStatusSummary(out, meta)
}

// renderStatusSummary prints a human-readable interpretation of the snapshot state.
func renderStatusSummary(out io.Writer, meta *snaplib.Meta) {
	switch meta.Status {
	case snaplib.StatusRolledBack:
		fmt.Fprintln(out, "✓ Snapshot has been fully rolled back.")
	case snaplib.StatusRollbackPartial:
		failed := 0
		for _, r := range meta.RollbackLog {
			if r.Status == snaplib.RBFailed {
				failed++
			}
		}
		fmt.Fprintf(out, "⚠ Partial rollback: %d of %d entities failed. Run `gotr snap rollback %s` to retry.\n",
			failed, len(meta.RollbackLog), meta.ID)
	case snaplib.StatusExpired:
		fmt.Fprintln(out, "Snapshot expired — data may have been cleaned up.")
	case snaplib.StatusAvailable:
		fmt.Fprintln(out, "✓ Snapshot is available for rollback.")
	}
}
