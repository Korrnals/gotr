package snap

import (
	"fmt"
	"os"
	"text/tabwriter"

	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all snapshots",
		Long:  "Displays a table of all snapshots from the manifest index.",
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
				fmt.Fprintln(os.Stdout, "No snapshots found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tOPERATION\tENTITY\tCATEGORY\tSTATUS\tTIMESTAMP")
			for _, e := range entries {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					e.ID, e.Operation, e.EntityType, e.Category, e.Status,
					e.Timestamp.Format("2006-01-02 15:04:05"))
			}
			return w.Flush()
		},
	}
}
