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
		Short: "List all snapshots",
		Long: `Displays snapshots grouped by server.

In interactive mode: two-level picker (server → snapshot).
With --format or in non-interactive mode: table with SERVER column.`,
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

			// Interactive mode: two-level picker.
			return listInteractive(cmd, entries)
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
	t.AppendHeader(table.Row{"#", "ID", "SERVER", "OP", "ENTITY", "CATEGORY", "STATUS", "TIMESTAMP"})
	for i, e := range entries {
		server := e.ServerURL
		if server == "" {
			server = "(unknown)"
		}
		t.AppendRow(table.Row{
			i + 1, e.ID, server, e.Operation, e.EntityType,
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

// formatEntryLabel creates a display label for a snapshot entry in select picker.
func formatEntryLabel(idx int, e snaplib.ManifestEntry) string {
	name := ""
	if e.Name != "" {
		name = fmt.Sprintf(" %q", e.Name)
	}
	return fmt.Sprintf("[%d] %s %s%s | %s | T%d | %s",
		idx, e.Operation, e.EntityType, name, e.Status,
		e.RollbackTier, e.Timestamp.Format("2006-01-02 15:04"))
}

// listInteractive performs two-level interactive selection: server → snapshot.
func listInteractive(cmd *cobra.Command, entries []snaplib.ManifestEntry) error {
	p := interactive.PrompterFromContext(cmd.Context())
	groups := groupByServer(entries)

	var selected serverGroup
	if len(groups) == 1 {
		selected = groups[0]
	} else {
		options := make([]string, len(groups))
		for i, g := range groups {
			label := g.URL
			if label == "" {
				label = "(unknown server)"
			}
			options[i] = fmt.Sprintf("%s — %d snapshots", label, len(g.Entries))
		}

		idx, _, err := p.Select("Select server:", options)
		if err != nil {
			return err
		}
		selected = groups[idx]
	}

	// Show snapshot picker within selected server.
	options := make([]string, len(selected.Entries))
	for i, e := range selected.Entries {
		options[i] = formatEntryLabel(i+1, e)
	}

	idx, _, err := p.Select("Select snapshot:", options)
	if err != nil {
		return err
	}

	snapID := selected.Entries[idx].ID
	fmt.Fprintf(cmd.OutOrStdout(), "\nSelected: %s\n", snapID)
	fmt.Fprintf(cmd.OutOrStdout(), "Run: gotr snap info %s\n", snapID)
	return nil
}
