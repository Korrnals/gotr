package report

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/paths"
	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/spf13/cobra"
)

func newOrganizeCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "organize",
		Short: "Move legacy flat reports under the categorized ~/.gotr/reports/ hierarchy",
		Long: `Since v3.3.0 gotr stores reports under
~/.gotr/reports/<category>/<label|default>/<YYYY-MM>/.

Run 'gotr report organize --dry-run' to preview the moves, then re-run without
--dry-run to apply them. Already-organized reports are left alone; INDEX.md is
regenerated at the end.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			reportsDir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("report organize: resolve reports dir: %w", err)
			}
			res, err := intreport.MigrateFlatLayout(reportsDir, dryRun)
			if err != nil {
				return fmt.Errorf("report organize: %w", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), res.Summary())
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show planned moves without touching the filesystem")
	return cmd
}
