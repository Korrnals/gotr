package tests

import (
	"time"

	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// printJSON outputs data as JSON to stdout.
// start parameter is kept for backward compatibility but unused.
func printJSON(cmd *cobra.Command, data interface{}, start time.Time) error {
	return ui.JSON(cmd, data)
}
