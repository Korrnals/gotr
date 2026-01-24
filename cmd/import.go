package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <src_file> [args...] <project_id>",
	Short: "Выполнить импортирование данных из файла в TestRail",
	Long:  "Выполняет импортирование данных из файла в TestRail",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("IMPORT команда в разработке...")
	},
}
