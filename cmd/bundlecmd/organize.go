package bundlecmd

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/exportsorg"
	"github.com/Korrnals/gotr/internal/paths"
	"github.com/spf13/cobra"
)

// newExportOrganizeCmd builds `gotr export organize`, which migrates the
// pre-v3.3.0 flat ~/.gotr/exports/ layout into the categorized layout
// (snaps/, reports/, api/).
func newExportOrganizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "organize",
		Short: "Reorganize ~/.gotr/exports/ into snaps/, reports/, api/ subdirectories",
		Long: `Scan ~/.gotr/exports/ for entries written by pre-v3.3.0 gotr and move them
into the new categorized layout:

  ~/.gotr/exports/snaps/    <- *.tar.gz, *.tgz
  ~/.gotr/exports/reports/  <- *.zip, *.pdf, *.md, *.json
  ~/.gotr/exports/api/      <- resource directories (cases/, suites/, ...)

Existing destinations are preserved and never overwritten.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			base, err := paths.ExportsDirPath()
			if err != nil {
				return fmt.Errorf("export organize: %w", err)
			}
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			res, err := exportsorg.MigrateExportsLayout(base, dryRun)
			if err != nil {
				return fmt.Errorf("export organize: %w", err)
			}
			if len(res.Plans) == 0 {
				infof("Exports layout already organized under %s", base)
				return nil
			}
			verb := "Would move"
			if !dryRun {
				verb = "Moved"
			}
			for _, p := range res.Plans {
				infof("%s %s -> %s", verb, p.RelSrc, p.RelDest)
			}
			if dryRun {
				infof("Dry-run: %d entr(y|ies) pending under %s", len(res.Plans), base)
			} else {
				infof("Organized: moved=%d skipped=%d (base=%s)", res.Moved, res.Skipped, base)
			}
			return nil
		},
	}
	cmd.Flags().Bool("dry-run", false, "Show what would be moved without touching the filesystem")
	return cmd
}
