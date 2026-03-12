package cmd

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/models/config"
	"github.com/Korrnals/gotr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	defaultConfig = "$HOME/.gotr/config/default.yaml"
)

// configCmd — родительская команда "config"
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Управление конфигурацией gotr",
	Long:  `Команды для работы с конфигурационным файлом утилиты gotr.`,
	// ОТКЛЮЧАЕМ PersistentPreRunE для всей ветки config
	// Это предотвращает создание клиента и проверку обязательных флагов
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Ничего не делаем — переопределяем родительский PersistentPreRunE
	},
}

// configInitCmd — подкоманда "init" - создает дефолтный файл-конфигурации
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Создать дефолтный файл конфигурации",
	Long: `Создаёт дефолтный файл конфигурации в ~/.gotr/config/default.yaml.

Пример:
	gotr config init			# Создать дефолтный конфиг

Default (config):
	$HOME/.gotr/config/default.yaml		# Путь до дефолтного файла конфигурации

Примечание:
	После создания обязательно отредактируйте файл, указав свои данные TestRail.`,

	// ОТКЛЮЧАЕМ PersistentPreRunE для всей ветки config
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Ничего не делаем — переопределяем родительский PersistentPreRunE
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Default()
		if err != nil {
			return err
		}
		return cfg.WithDefaults().Create()
	},
}

// configPathCmd - подкоманда "path" - для получения пути текущего файла-конфигурации
var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Показать путь к текущему конфиг-файлу",
	Long:  `Выводит путь к файлу конфигурации, который используется в данный момент.`,

	// ОТКЛЮЧАЕМ PersistentPreRunE для всей ветки config
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Ничего не делаем — переопределяем родительский PersistentPreRunE
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Default()
		if err != nil {
			return fmt.Errorf("failed to determine config path: %w", err)
		}

		used := viper.ConfigFileUsed()
		if used == "" {
			fmt.Printf("Config file not found.\nExpected location: %s\n", cfg.PathString())
			fmt.Println("Create it with: gotr config init")
		} else {
			fmt.Printf("Current config file: %s\n", used)
		}
		return nil
	},
}

// configViewCmd - подкоманда "view" - для быстрого просмотра содержимого конфига
var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Показать содержимое текущего конфиг-файла",
	Long:  "Выводит содержимое конфигурационного файла в читаемом виде.",

	// ОТКЛЮЧАЕМ PersistentPreRunE для всей ветки config
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Ничего не делаем — переопределяем родительский PersistentPreRunE
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		used := viper.ConfigFileUsed()
		if used == "" {
			cfg, _ := config.Default()
			fmt.Printf("Config file not found: %s\n", cfg.PathString())
			fmt.Println("Create it with: gotr config init")
			return nil
		}

		data, err := os.ReadFile(used)
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		fmt.Printf("Config file contents %s:\n\n%s\n", used, string(data))
		return nil
	},
}

// configEditCmd - подкоманда "edit" - для быстрого редактирования содержимого конфига
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Открыть конфиг-файл в редакторе по умолчанию",
	Long: `Открывает текущий конфигурационный файл в редакторе, указанном в переменной окружения EDITOR.
Если EDITOR не задан — используется fallback (vi/nano на Linux, notepad на Windows).

Примеры:
	export EDITOR=code    # VS Code
	export EDITOR=nano
	gotr config edit      # откроет в указанном редакторе`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Отключаем родительский PreRun (клиент не нужен)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		// Определяем путь к конфигу
		used := viper.ConfigFileUsed()
		if used == "" {
			cfg, err := config.Default()
			if err != nil {
				return fmt.Errorf("failed to determine config path: %w", err)
			}
			fmt.Printf("Config file not found: %s\n", cfg.PathString())
			fmt.Println("Create it with: gotr config init")
			return nil
		}

		// Запускаем редактор
		if err := utils.OpenEditor(used); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		fmt.Printf("Config file opened in editor: %s\n", used)
		return nil
	},
}
