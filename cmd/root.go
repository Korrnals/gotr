package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/config"
	"github.com/Korrnals/gotr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Версия утилиты — заполняется при сборке через -ldflags
	Version = "2.5.0" // значение по умолчанию для локальной разработки
	Commit  = "unknown"
	Date    = "unknown"
)

// rootCmd — главная команда: gotr
var rootCmd = &cobra.Command{
	Use:   "gotr",                                 // Название утилиты
	Short: "CLI-клиент для работы с TestRail API", // Краткое описание
	Long: `gotr — удобная утилита для работы с TestRail API v2.
Поддерживает просмотр доступных эндпоинтов, выполнение запросов и многое другое.`,
	// Запускается клиент перед каждой субкомандой
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		utils.DebugPrint("{rootCmd} - Запуск команды: %s", cmd.Use)
		utils.DebugPrint("{rootCmd} - Аргументы: %v", args)
		// Настройка Viper - поддержка env, флагов, конфигов
		viper.AutomaticEnv() // автоматически подтягивать переменные из окружения

		// Маппим переменные окружения в переменные, которые будут использоваться в клиенте
		baseURL := viper.GetString("base_url")
		username := viper.GetString("username")
		password := viper.GetString("password")
		apiKey := viper.GetString("api_key")
		insecure := viper.GetBool("insecure")
		debug := viper.GetBool("debug")

		// [DEBUG] при переданном флаге `--debug` или `-d`
		utils.DebugPrint("{rootCmd} - PersistentPreRunE запущен для команды: %s", cmd.Use)
		utils.DebugPrint("{rootCmd} - baseURL=%s, username=%s", baseURL, username)
		utils.DebugPrint("{rootCmd} - insecure=%v", insecure)

		// Проверяем, что конфиг не пустой и не содержит дефолтных placeholder'ов
		if config.IsDefaultValue(baseURL, config.DefaultBaseURL) ||
			config.IsDefaultValue(username, config.DefaultUsername) ||
			config.IsDefaultValue(apiKey, config.DefaultAPIKey) {
			return fmt.Errorf("конфигурация не задана или содержит дефолтные значения\n" +
				"Запустите 'gotr config init' для создания конфигурации,\n" +
				"затем отредактируйте файл ~/.gotr/config/default.yaml")
		}

		// [DEBUG] при переданном флаге `--debug` или `-d`
		utils.DebugPrint("{rootCmd} - Подключение к %s как %s", baseURL, username)

		// Создаём клиент с опциями
		opts := []client.ClientOption{}
		if insecure {
			opts = append(opts, client.WithSkipTlsVerify(true)) //По-умолчанию, проверка tls - включена
		}

		httpClient, err := client.NewClient(baseURL, username, password, debug, opts...)
		if err != nil {
			return fmt.Errorf("не удалось создать клиент: %w", err)
		}

		// [DEBUG] при переданном флаге `--debug` или `-d`
		utils.DebugPrint("{rootCmd} - Клиент успешно создан и сохранён в контекст")

		// Сохраняем клиент в контекст — будет доступен во всех субкомандах
		ctx := context.WithValue(cmd.Context(), httpClientKey, httpClient)
		cmd.SetContext(ctx)

		return nil
	},
	// Run: func(cmd *cobra.Command, args []string) { } // Можно оставить пустым — будет показывать help
}

// Execute — вызывается из main.go
func Execute() {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// GetClient — удобная функция для получения клиента из контекста
func GetClient(cmd *cobra.Command) *client.HTTPClient {
	val := cmd.Context().Value(httpClientKey)
	if val == nil {
		fmt.Fprintln(os.Stderr, "ОШИБКА: HTTP-клиент не инициализирован. Проверьте --username, --api-key и --url")
		os.Exit(1)
	}
	return val.(*client.HTTPClient)
}

// GetClientInterface — получает клиент как интерфейс (для тестов с mock)
func GetClientInterface(cmd *cobra.Command) client.ClientInterface {
	val := cmd.Context().Value(httpClientKey)
	if val == nil {
		fmt.Fprintln(os.Stderr, "ОШИБКА: HTTP-клиент не инициализирован. Проверьте --username, --api-key и --url")
		os.Exit(1)
	}
	// Поддерживаем как *client.HTTPClient, так и *client.MockClient
	if cli, ok := val.(client.ClientInterface); ok {
		return cli
	}
	return val.(*client.HTTPClient)
}

func initConfig() {
	// 1. Добавляем стандартные пути поиска
	home, err := os.UserHomeDir()
	if err != nil {
		// Не паникуем — просто продолжаем без home-пути
		fmt.Printf("Warning: cannot get user home directory: %v\n", err)
	} else {
		configDir := filepath.Join(home, ".gotr", "config")
		viper.AddConfigPath(configDir) // ~/.gotr/config
	}

	// Удобно для локального тестирования
	viper.AddConfigPath(".") // текущая директория

	// 2. Имя файла без расширения (viper сам попробует .yaml, .json и т.д.)
	viper.SetConfigName("default") // ищем default.yaml (в ~/.gotr/config/)
	viper.SetConfigType("yaml")    // явно указываем yaml

	// 3. Автоматический биндинг env-переменных
	viper.SetEnvPrefix("testrail") // TESTRAIL_BASE_URL, TESTRAIL_USER и т.д.
	viper.AutomaticEnv()

	// 4. Читаем конфиг (если файла нет — просто продолжаем, это не ошибка)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Если ошибка не "файл не найден" — можно залогировать
			fmt.Printf("{rootCmd} - Config file error: %v\n", err)
		}
		// Файл не обязателен — используем defaults
	}
}
