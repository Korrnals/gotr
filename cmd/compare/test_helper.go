package compare

import (
	"time"

	"github.com/spf13/cobra"
)

// addPersistentFlagsForTests adds persistent flags that are normally registered
// in Register() function. Use this in tests that create commands directly.
func addPersistentFlagsForTests(cmd *cobra.Command) {
	cmd.Flags().StringP("pid1", "1", "", "First project ID")
	cmd.Flags().StringP("pid2", "2", "", "Second project ID")
	cmd.Flags().StringP("format", "f", "table", "Output format")
	cmd.Flags().BoolP("quiet", "q", false, "Suppress informational output")
	cmd.Flags().Bool("save", false, "Save result")
	cmd.Flags().String("save-to", "", "Save to a specific file")
	cmd.Flags().Int("rate-limit", -1, "")
	cmd.Flags().Int("parallel-suites", 8, "")
	cmd.Flags().Int("parallel-pages", 10, "")
	cmd.Flags().Int("page-retries", 5, "")
	cmd.Flags().Duration("timeout", 30*time.Minute, "")
	cmd.Flags().Int("retry-attempts", 3, "")
	cmd.Flags().Int("retry-workers", 6, "")
	cmd.Flags().Duration("retry-delay", 500*time.Millisecond, "")
}
