package cmd

import (
	"github.com/Korrnals/gotr/cmd/attachments"
	"github.com/Korrnals/gotr/cmd/cases"
	"github.com/Korrnals/gotr/cmd/get"
	"github.com/Korrnals/gotr/cmd/labels"
	"github.com/Korrnals/gotr/cmd/plans"
	"github.com/Korrnals/gotr/cmd/result"
	"github.com/Korrnals/gotr/cmd/run"
	"github.com/Korrnals/gotr/cmd/sync"
	"github.com/Korrnals/gotr/cmd/test"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// init регистрирует все команды при старте
func init() {
	// Сначала инициализируем конфиг
	initConfig()

	// Глобальные флаги root команды
	initGlobalFlags()

	// Регистрация всех команд
	registerConfigCmd()
	registerListCmd()
	registerAddCmd()
	registerDeleteCmd()
	registerUpdateCmd()
	registerCopyCmd()
	registerExportCmd()
	registerImportCmd()
	registerCompareCmd()
	registerCompletionCmd()

	// Регистрация команд из подпакетов (передаем GetClient)
	attachments.Register(rootCmd, GetClient)
	cases.Register(rootCmd, GetClientInterface)
	get.Register(rootCmd, GetClient)
	labels.Register(rootCmd, GetClientInterface)
	plans.Register(rootCmd, GetClientInterface)
	run.Register(rootCmd, GetClient)
	result.Register(rootCmd, GetClient)
	sync.Register(rootCmd, GetClient)
	test.Register(rootCmd, GetClient)
}

// initGlobalFlags инициализирует глобальные флаги для rootCmd
func initGlobalFlags() {
	// Глобальные флаги — ТОЛЬКО подключение и базовая настройка
	rootCmd.PersistentFlags().String("url", "", "Базовый URL TestRail")
	rootCmd.PersistentFlags().StringP("username", "u", "", "Email пользователя TestRail")
	rootCmd.PersistentFlags().StringP("api-key", "k", "", "API ключ TestRail")
	rootCmd.PersistentFlags().Bool("insecure", false, "Пропустить проверку TLS сертификата")
	rootCmd.PersistentFlags().BoolP("config", "c", false, "Создать дефолтный файл конфигурации")

	// Скрытый debug
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Включить отладочный вывод")
	rootCmd.PersistentFlags().MarkHidden("debug")

	// Биндим к Viper
	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// ============================================
// Конфигурация
// ============================================

func registerConfigCmd() {
	rootCmd.AddCommand(configCmd)

	// Подкоманды config
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configEditCmd)

	// Автодополнение для "gotr config "
	configCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"init", "path", "view", "edit"}, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// ============================================
// List
// ============================================

func registerListCmd() {
	rootCmd.AddCommand(listCmd)

	// Флаги только для list
	listCmd.Flags().Bool("json", false, "Вывести в формате JSON")
	listCmd.Flags().Bool("short", false, "Краткий вывод (только URI)")

	// Автодополнение для первого аргумента (ресурса)
	listCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return ValidResources, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveDefault
	}
}

// ============================================
// Add
// ============================================

func registerAddCmd() {
	rootCmd.AddCommand(addCmd)
}

// ============================================
// Delete
// ============================================

func registerDeleteCmd() {
	rootCmd.AddCommand(deleteCmd)
}

// ============================================
// Update
// ============================================

func registerUpdateCmd() {
	rootCmd.AddCommand(updateCmd)
}

// ============================================
// Copy
// ============================================

func registerCopyCmd() {
	rootCmd.AddCommand(copyCmd)
}

// ============================================
// Export
// ============================================

func registerExportCmd() {
	rootCmd.AddCommand(exportCmd)

	// Специфичные флаги для export
	exportCmd.Flags().StringP("project-id", "p", "", "ID проекта (для эндпоинтов с {project_id})")
	exportCmd.Flags().StringP("suite-id", "s", "", "ID тест-сюиты (для get_cases)")
	exportCmd.Flags().String("section-id", "", "ID секции (для get_cases)")
	exportCmd.Flags().String("milestone-id", "", "ID milestone (для get_runs)")
	exportCmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (если указан)")

	// Автодополнение
	exportCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return ValidResources, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			endpoints, _ := getResourceEndpoints(args[0], "list")
			return endpoints, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveDefault
	}
}

// ============================================
// Import
// ============================================

func registerImportCmd() {
	rootCmd.AddCommand(importCmd)
}

// ============================================
// Compare
// ============================================

func registerCompareCmd() {
	rootCmd.AddCommand(compareCmd)

	// Флаги для compare
	compareCmd.Flags().StringP("pid1", "1", "", "ID первого проекта (обязательно)")
	compareCmd.Flags().StringP("pid2", "2", "", "ID второго проекта (обязательно)")
	compareCmd.Flags().String("field", "title", "Поле для сравнения (title, priority_id, custom_preconds и т.д.)")
}

// ============================================
// Completion
// ============================================

func registerCompletionCmd() {
	rootCmd.AddCommand(completionCmd)
}
