package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <endpoint> [args...]",
	Short: "Выполнить POST-запрос (создание ресурса)",
	Long:  "Создаёт новый объект (case, run, project и т.д.).",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ADD команда в разработке...")
	},
}
