package sync

import (
	"context"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/spf13/cobra"
)

// addSyncFlags defines a common set of flags for sync/* commands.
func addSyncFlags(c *cobra.Command) {
	c.Flags().Int64("src-project", 0, "Source project ID")
	c.Flags().Int64("src-suite", 0, "Source suite ID")
	c.Flags().Int64("dst-project", 0, "Destination project ID")
	c.Flags().Int64("dst-suite", 0, "Destination suite ID")
	c.Flags().String("compare-field", "title", "Field for duplicate detection")
	c.Flags().Bool("dry-run", false, "Preview without importing")
	c.Flags().BoolP("approve", "y", false, "Auto-approve confirmation")
	c.Flags().BoolP("save-mapping", "m", false, "Save mapping automatically")
	c.Flags().String("mapping-file", "", "Mapping file for shared_step_id replacement")
	c.Flags().String("output", "", "Additional JSON output")
	snap.RegisterFlags(c)
}

// applySessionFallback fills zero project/suite IDs from the work session if available.
// This enables parameter inheritance: compare → sync → snap.
func applySessionFallback(ctx context.Context, srcProject, dstProject, srcSuite, dstSuite *int64) {
	s := interactive.SessionFromContext(ctx)
	if s == nil {
		return
	}
	sessSrc, sessDst := s.Projects()
	if *srcProject == 0 {
		*srcProject = sessSrc
	}
	if *dstProject == 0 {
		*dstProject = sessDst
	}
	sessSrcSuite, sessDstSuite := s.Suites()
	if *srcSuite == 0 {
		*srcSuite = sessSrcSuite
	}
	if *dstSuite == 0 {
		*dstSuite = sessDstSuite
	}
}
