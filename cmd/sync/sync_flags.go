package sync

import "github.com/spf13/cobra"

// addSyncFlags — общий набор флагов для команд sync/*
func addSyncFlags(c *cobra.Command) {
	c.Flags().Int64("src-project", 0, "Source project ID")
	c.Flags().Int64("src-suite", 0, "Source suite ID")
	c.Flags().Int64("dst-project", 0, "Destination project ID")
	c.Flags().Int64("dst-suite", 0, "Destination suite ID")
	c.Flags().String("compare-field", "title", "Поле для дубликатов")
	c.Flags().Bool("dry-run", false, "Просмотр без импорта")
	c.Flags().BoolP("approve", "y", false, "Автоматическое подтверждение")
	c.Flags().BoolP("save-mapping", "m", false, "Автоматически сохранить mapping")
	c.Flags().String("mapping-file", "", "Mapping файл для замены shared_step_id")
	c.Flags().String("output", "", "Дополнительный JSON")
}
