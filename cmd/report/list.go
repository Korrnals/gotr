package report

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Korrnals/gotr/internal/paths"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local migration reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("report list: resolve reports dir: %w", err)
			}

			entries, err := os.ReadDir(reportsDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintln(cmd.OutOrStdout(), "No reports found")
					return nil
				}
				return fmt.Errorf("report list: read reports dir: %w", err)
			}

			reports := make([]string, 0)
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				if strings.HasPrefix(e.Name(), "migration-") && strings.HasSuffix(e.Name(), ".md") {
					reports = append(reports, e.Name())
				}
			}
			if len(reports) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No reports found")
				return nil
			}

			sort.Sort(sort.Reverse(sort.StringSlice(reports)))
			limit, _ := cmd.Flags().GetInt("limit")
			if limit <= 0 || limit > len(reports) {
				limit = len(reports)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Migration reports (%d):\n", len(reports))
			for i := 0; i < limit; i++ {
				fmt.Fprintf(cmd.OutOrStdout(), "  %d. %s\n", i+1, reports[i])
			}
			return nil
		},
	}

	cmd.Flags().Int("limit", 20, "Max number of reports to show")
	return cmd
}
