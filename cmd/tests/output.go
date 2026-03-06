package tests

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// printJSON outputs data as JSON to stdout.
// start parameter is kept for backward compatibility but unused.
func printJSON(cmd *cobra.Command, data interface{}, start time.Time) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes))
	return nil
}
