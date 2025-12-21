package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
    Use:   "diff <endpoint> [args...]",
    Short: "Выполнить сравнение по критериям",
    Long:  "Отсеивает test-case по совпадающим case_ids",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("DIFF команда в разработке...")
    },
}