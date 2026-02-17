package compare

import "github.com/spf13/cobra"

// addPersistentFlagsForTests adds persistent flags that are normally registered
// in Register() function. Use this in tests that create commands directly.
func addPersistentFlagsForTests(cmd *cobra.Command) {
	cmd.Flags().StringP("pid1", "1", "", "ID первого проекта")
	cmd.Flags().StringP("pid2", "2", "", "ID второго проекта")
	cmd.Flags().StringP("format", "f", "table", "Формат вывода")
	cmd.Flags().Bool("save", false, "Сохранить результат")
	cmd.Flags().String("save-to", "", "Сохранить в указанный файл")
}
