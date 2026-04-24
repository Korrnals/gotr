package report

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/Korrnals/gotr/internal/paths"
	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "view [report-file|latest]",
		Short:             "View a migration report",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeReportArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("report view: resolve reports dir: %w", err)
			}
			target, err := resolveViewTarget(cmd, args, reportsDir)
			if err != nil {
				return err
			}
			path, err := intreport.ResolveReportPath(reportsDir, target)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return fmt.Errorf("report view: %q not found under %s", target, reportsDir)
				}
				return fmt.Errorf("report view: %w", err)
			}
			content, err := os.ReadFile(path) //nolint:gosec // path resolved via ResolveReportPath
			if err != nil {
				return fmt.Errorf("report view: read %s: %w", path, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "# %s\n\n", path)
			fmt.Fprint(cmd.OutOrStdout(), string(content))
			return nil
		},
	}
}

// resolveViewTarget mirrors resolveShowTarget for `gotr report view`.
func resolveViewTarget(cmd *cobra.Command, args []string, reportsDir string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if !shouldPromptForReport(cmd) {
		return "", fmt.Errorf("report view: a report name or 'latest' is required (pass as argument or run interactively)")
	}
	return promptForReport(cmd, reportsDir)
}
