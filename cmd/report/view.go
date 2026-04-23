package report

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Korrnals/gotr/internal/paths"
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view [report-file|latest]",
		Short: "View a migration report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("report view: resolve reports dir: %w", err)
			}

			target := args[0]
			if target == "latest" {
				latest, err := resolveLatestReport(reportsDir)
				if err != nil {
					return fmt.Errorf("report view: %w", err)
				}
				target = latest
			}

			reportPath := filepath.Join(reportsDir, target)
			content, err := os.ReadFile(reportPath)
			if err != nil {
				return fmt.Errorf("report view: read report %s: %w", target, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "# %s\n\n", target)
			fmt.Fprint(cmd.OutOrStdout(), string(content))
			return nil
		},
	}
}

func resolveLatestReport(reportsDir string) (string, error) {
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no reports found")
		}
		return "", err
	}

	reports := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if reportLike(e.Name()) {
			reports = append(reports, e.Name())
		}
	}
	if len(reports) == 0 {
		return "", fmt.Errorf("no reports found")
	}

	sort.Sort(sort.Reverse(sort.StringSlice(reports)))
	return reports[0], nil
}
