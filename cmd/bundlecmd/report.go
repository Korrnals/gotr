package bundlecmd

import (
	"fmt"
	"strings"

	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/reportbundle"
	"github.com/spf13/cobra"
)

func newExportReportCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report <filename|all>",
		Short: "Export migration reports",
		Long: `Export gotr migration reports from ~/.gotr/reports/.
  report <filename>   Copies a single file to ~/.gotr/exports/reports/<basename>.
  report all          Bundles all reports into
                      ~/.gotr/exports/reports/reports_<YYYYMMDD>.zip.`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeExportReportArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("export report: %w", err)
			}
			outPath, _ := cmd.Flags().GetString("out")
			filter, _ := cmd.Flags().GetString("filter")

			target := args[0]
			var res *reportbundle.Result
			switch target {
			case "all":
				res, err = reportbundle.ExportAll(reportsDir, outPath, reportbundle.ExportOptions{
					GotrVersion: version,
					Filter:      filter,
				})
			default:
				if filter != "" {
					return fmt.Errorf("export report: --filter is only valid with 'all'")
				}
				res, err = reportbundle.ExportSingle(reportsDir, target, outPath)
			}
			if err != nil {
				return err
			}
			if target == "all" {
				successf("Exported %d reports → %s", len(res.Files), res.ArchivePath)
			} else {
				successf("Exported report %s → %s", target, res.ArchivePath)
			}
			return nil
		},
	}
	cmd.Flags().String("out", "", "Destination path (default: ~/.gotr/exports/reports/<name>)")
	cmd.Flags().String("filter", "", "When exporting 'all', keep only reports whose name matches this glob")
	return cmd
}

func newImportReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "report <path.zip|path.pdf|path.md|path.json>",
		Short:             "Import reports into ~/.gotr/reports/",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeImportReportArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("import report: %w", err)
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			dry, _ := cmd.Flags().GetBool("dry-run")

			res, err := reportbundle.Import(reportsDir, args[0], reportbundle.ImportOptions{
				Overwrite: overwrite,
				DryRun:    dry,
			})
			if err != nil {
				return err
			}
			verb := "Imported"
			if dry {
				verb = "Would import"
			}
			successf("%s %d report(s): %s", verb, len(res.Copied), strings.Join(res.Copied, ", "))
			return nil
		},
	}
	cmd.Flags().Bool("overwrite", false, "Overwrite existing reports with the same filename")
	cmd.Flags().Bool("dry-run", false, "Validate and show what would be imported; no changes")
	return cmd
}
