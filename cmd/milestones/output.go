package milestones

import (
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// outputList outputs list data as JSON or saves to file.
// Kept as package-level alias for backward compatibility with tests.
func outputList(cmd *cobra.Command, data interface{}) error {
	return output.OutputResult(cmd, data, "milestones")
}
