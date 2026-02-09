package milestones

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

// outputResult выводит результат в зависимости от формата
func outputResult(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")

	switch output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	default:
		return nil
	}
}

// outputList выводит список результатов в зависимости от формата
func outputList(cmd *cobra.Command, data interface{}) error {
	return outputResult(cmd, data)
}
