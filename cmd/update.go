package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// updateCmd — команда для обновления ресурсов
var updateCmd = &cobra.Command{
	Use:   "update <endpoint> [args...]",
	Short: "Выполнить POST-запрос (обновление ресурса)",
	Long:  "Обновляет существующий объект (case, run, project и т.д.).",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("UPDATE команда в разработке...")
	},
}
