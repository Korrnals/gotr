package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/debug"
	"github.com/Korrnals/gotr/internal/models/config"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Версия утилиты — заполняется при сборке через -ldflags
	Version = "3.0.0-dev" // значение по умолчанию для локальной разработки
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
		debug.DebugPrint("{rootCmd} - Running command: %s", cmd.Use)
		debug.DebugPrint("{rootCmd} - Arguments: %v", args)
		// Настройка Viper - поддержка env, флагов, конфигов
		viper.AutomaticEnv() // автоматически подтягивать переменные из окружения

		// Маппим переменные окружения в переменные, которые будут использоваться в клиенте
		baseURL := viper.GetString("base_url")
		username := viper.GetString("username")
		insecure := viper.GetBool("insecure")
		debugMode := viper.GetBool("debug")

		// Поддержка password (Basic Auth) и api_key (API Key Auth).
		// password имеет приоритет — для обратной совместимости с TESTRAIL_PASSWORD.
		apiKey := viper.GetString("password")
		if apiKey == "" {
			apiKey = viper.GetString("api_key")
		}

		// [DEBUG] при переданном флаге `--debug` или `-d`
		debug.DebugPrint("{rootCmd} - PersistentPreRunE running for command: %s", cmd.Use)
		debug.DebugPrint("{rootCmd} - baseURL=%s, username=%s", baseURL, username)
		debug.DebugPrint("{rootCmd} - insecure=%v", insecure)

		// Проверяем, что конфиг не пустой и не содержит дефолтных placeholder'ов
		if config.IsDefaultValue(baseURL, config.DefaultBaseURL) ||
			config.IsDefaultValue(username, config.DefaultUsername) ||
			config.IsDefaultValue(apiKey, config.DefaultAPIKey) {
			return fmt.Errorf("configuration not set or contains default values\n" +
				"Run 'gotr config init' to create configuration,\n" +
				"then edit the file ~/.gotr/config/default.yaml")
		}

		// [DEBUG] при переданном флаге `--debug` или `-d`
		debug.DebugPrint("{rootCmd} - Connecting to %s as %s", baseURL, username)

		// Создаём клиент с опциями
		opts := []client.ClientOption{}
		if insecure {
			opts = append(opts, client.WithSkipTlsVerify(true)) //По-умолчанию, проверка tls - включена
		}

		httpClient, err := client.NewClient(baseURL, username, apiKey, debugMode, opts...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// [DEBUG] при переданном флаге `--debug` или `-d`
		debug.DebugPrint("{rootCmd} - Client created and stored in context")

		// Сохраняем клиент в контекст — будет доступен во всех субкомандах
		ctx := context.WithValue(cmd.Context(), httpClientKey, httpClient)
		cmd.SetContext(ctx)

		return nil
	},
	// Run: func(cmd *cobra.Command, args []string) { } // Можно оставить пустым — будет показывать help
}

// Execute — вызывается из main.go с контекстом (поддерживает signal.NotifyContext)
func Execute(ctx context.Context) {
	rootCmd.SilenceUsage = true  // не печатать usage при ошибке
	rootCmd.SilenceErrors = true // сами обработаем вывод ошибки
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "\nПрервано.")
			os.Exit(130)
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// GetClient — удобная функция для получения клиента из контекста
func GetClient(cmd *cobra.Command) *client.HTTPClient {
	val := cmd.Context().Value(httpClientKey)
	if val == nil {
		panic("gotr: HTTP client not initialized. Check --username, --api-key and --url")
	}
	return val.(*client.HTTPClient)
}

// GetClientInterface — получает клиент как интерфейс (для тестов с mock)
func GetClientInterface(cmd *cobra.Command) client.ClientInterface {
	val := cmd.Context().Value(httpClientKey)
	if val == nil {
		panic("gotr: HTTP client not initialized. Check --username, --api-key and --url")
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
		ui.Warningf(os.Stderr, "cannot get user home directory: %v", err)
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
			ui.Warningf(os.Stderr, "Config file error: %v", err)
		}
		// Файл не обязателен — используем defaults
	}
}
