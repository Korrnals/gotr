package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// copyCmd — команда для копирования ресурсов
var copyCmd = &cobra.Command{
	Use:   "copy <src_project_id> [args... <resource>] <dest_project_id>",
	Short: "Выполнить копирование из одного ресу",
	Long:  "Выполнить экспортирование данных из одного проекта в другой",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("COPY команда в разработке...")
	},
}
