package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
    Use:   "delete <endpoint> [args...]",
    Short: "Выполнить POST-запрос (удаление ресурса)",
    Long:  "Удаляет существующий объект (case, run, project и т.д.).",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("DELETE команда в разработке...")
    },
}