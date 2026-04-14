package snap

import (
	"fmt"
	"sort"

	"github.com/Korrnals/gotr/internal/interactive"
	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Browse and inspect snapshots",
		Long: `Interactive snapshot browser with three-level navigation:

  1. Server — pick the TestRail instance
  2. Operation — filter by action (add, update, delete, …)
  3. Snapshot — choose a specific snapshot to inspect

After selecting a snapshot, its details are shown (same as 'gotr snap info').
Use "← Back" to navigate between levels.

With --format or in non-interactive mode: flat table with all columns.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snaplib.NewStore()
			if err != nil {
				return fmt.Errorf("snap store: %w", err)
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("snap manifest: %w", err)
			}

			entries := manifest.All()
			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No snapshots found.")
				return nil
			}

			// Non-interactive or explicit format → table/json output.
			if interactive.IsNonInteractive(cmd.Context()) || hasExplicitFormat(cmd) {
				return listTable(cmd, entries)
			}

			// Interactive: full three-level browser.
			return listInteractive(cmd, store, manifest)
		},
	}
}

// hasExplicitFormat returns true if --format was explicitly set by the user.
func hasExplicitFormat(cmd *cobra.Command) bool {
	if f := cmd.Flags().Lookup("format"); f != nil && f.Changed {
		return true
	}
	if f := cmd.InheritedFlags().Lookup("format"); f != nil && f.Changed {
		return true
	}
	return false
}

// listTable renders snapshot entries as a table or JSON.
func listTable(cmd *cobra.Command, entries []snaplib.ManifestEntry) error {
	if ui.IsJSON(cmd) {
		return ui.JSON(cmd, entries)
	}

	t := ui.NewTable(cmd)
	t.AppendHeader(table.Row{"#", "ID", "SERVER", "OP", "ENTITY", "IDS", "CATEGORY", "STATUS", "TIMESTAMP"})
	for i, e := range entries {
		t.AppendRow(table.Row{
			i + 1, e.ID, serverLabel(e.ServerURL),
			e.Operation, e.EntityType, entityIDsLabel(e.EntityIDs),
			e.Category, e.Status,
			e.Timestamp.Format("2006-01-02 15:04:05"),
		})
	}
	ui.Table(cmd, t)
	return nil
}

// serverGroup pairs a server URL with its snapshot entries.
type serverGroup struct {
	URL     string
	Entries []snaplib.ManifestEntry
}

// groupByServer groups entries by ServerURL in sorted order.
func groupByServer(entries []snaplib.ManifestEntry) []serverGroup {
	m := make(map[string][]snaplib.ManifestEntry)
	for _, e := range entries {
		m[e.ServerURL] = append(m[e.ServerURL], e)
	}

	groups := make([]serverGroup, 0, len(m))
	for url, ents := range m {
		groups = append(groups, serverGroup{URL: url, Entries: ents})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].URL < groups[j].URL
	})
	return groups
}

// listInteractive opens three-level browser: server → operation → snapshot → info card.
// After viewing a card the user is returned to the snapshot list.
func listInteractive(cmd *cobra.Command, store *snaplib.Store, manifest *snaplib.Manifest) error {
	return browseSnapshots(cmd, store, manifest)
}
