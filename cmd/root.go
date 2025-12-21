package cmd

import (
	"context"
	"fmt"
	"gotr/internal/client"
	"gotr/internal/utils"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
    // Версия утилиты — заполняется при сборке через -ldflags
    Version = "dev" // значение по умолчанию для локальной разработки
    Commit  = "unknown"
    Date    = "unknown"
)

// rootCmd — главная команда: gotr
var rootCmd = &cobra.Command{
    Use:   "gotr",                                   // Название утилиты
    Short: "CLI-клиент для работы с TestRail API",            // Краткое описание
    Long: `gotr — удобная утилита для работы с TestRail API v2.
Поддерживает просмотр доступных эндпоинтов, выполнение запросов и многое другое.`,
    // Запускается клиент перед каждой субкомандой
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        utils.DebugPrint("{rootCmd} - Запуск команды: %s", cmd.Use)
        utils.DebugPrint("{rootCmd} - Аргументы: %v", args)
        // Настройка Viper - поддержка env, флаговб конфигов
        viper.AutomaticEnv()            // автоматически поддтягивать переменные из окружения
 
        // Маппим переменные окружения в переменные, которые будут использоваться в клиенте
        baseURL  := viper.GetString("base_url")
        username := viper.GetString("username")
        password := viper.GetString("password")
        apiKey   := viper.GetString("api_key")
        insecure := viper.GetBool("insecure")
        debug    := viper.GetBool("debug")

        // Читаем значения (приоритет: флаги > env > default)
        if baseURL == "" {
            return fmt.Errorf("url обязателен: укажите --url или GOTR_TESTRAIL_BASE_URL (env)")
        }
        
        // [DEBUG] при переданом флаге `--debug` или `-d`
        utils.DebugPrint("{rootCmd} - PersistentPreRunE запущен для команды:", cmd.Use)
        utils.DebugPrint("{rootCmd} - baseURL=%s, username=%s", baseURL, username)
        utils.DebugPrint("{rootCmd} - insecure=%v", insecure)

        // Обработка пустых значений переменных
        if baseURL == "" {
            return fmt.Errorf("base_url обязателен: --url или TESTRAIL_BASE_URL (env)")
        }
        if username == "" {
            return fmt.Errorf("username обязателен: --username или TESTRAIL_USERNAME")
        }
        if apiKey == "" {
            return fmt.Errorf("api_key обязателен: --api-key или TESTRAIL_API_KEY")
        }

        // [DEBUG] при переданом флаге `--debug` или `-d
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

        // [DEBUG] при переданом флаге `--debug` или `-d
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
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

// init — здесь подключаем все субкоманды
func init() {
    // Инициализация конфига
    initConfig()
    
    // Глобальные флаги
	rootCmd.PersistentFlags().String("url", "", "Базовый URL TestRail (например, https://your-instance.testrail.io/")
	rootCmd.PersistentFlags().StringP("username", "u", "", "Email пользователя TestRail")
    rootCmd.PersistentFlags().BoolP("config", "c", false, "Создать дефолтный файл конфигурации (~/.gotr/config.yaml)")
	rootCmd.PersistentFlags().StringP("api-key", "k", "", "API ключ TestRail")
	rootCmd.PersistentFlags().BoolP("insecure", "i", false, "Пропустить проверку TLS сертификата")

    // jq флаги — глобальные, работают во всех подкомандах
    rootCmd.PersistentFlags().BoolP("jq", "j", false, "Выводить через jq (по умолчанию весь JSON как jq .)")
    rootCmd.PersistentFlags().StringP("jq-filter", "f", "", "Явный фильтр jq (переопределяет -j)")
    rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Подавить вывод в консоль (полезно для скриптов и --save)")
 
    // Формат вывода (json по умолчанию — только тело)
	rootCmd.PersistentFlags().StringP("type", "t", "json", "Формат вывода: json (только тело), json-full (полный ответ), table")

    // Скрытый флаг --debug (не показывается в --help)
    rootCmd.PersistentFlags().BoolP("debug", "d", false, "Включить отладочный вывод (скрытый флаг)")
    rootCmd.PersistentFlags().MarkHidden("debug") // ← Эта строка скрывает флаг

	// Привязываем флаги к Viper (чтобы viper знал их значения)
	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure"))
    // Вторичные флаги
    viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
    viper.BindPFlag("jq", rootCmd.PersistentFlags().Lookup("jq"))
    viper.BindPFlag("jq-filter", rootCmd.PersistentFlags().Lookup("jq-filter"))
    viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
    viper.BindPFlag("type", rootCmd.PersistentFlags().Lookup("type"))

	// Defaults (если ничего не задано)
	viper.SetDefault("base_url", "https://yourcompany.testrail.io/")

	// Глобальные ПОДКОМАНДЫ:
    rootCmd.AddCommand(configCmd) // Подключаем субкоманду list
    rootCmd.AddCommand(listCmd) // Подключаем субкоманду list
    rootCmd.AddCommand(getCmd)  // Подключаем субкоманду get
    rootCmd.AddCommand(addCmd)  // Подключаем субкоманду add
    rootCmd.AddCommand(updateCmd) // Подключаем субкоманду update
    rootCmd.AddCommand(deleteCmd) // Подключаем субкоманду delete
    rootCmd.AddCommand(copyCmd)  // Подключаем субкоманду copy
    rootCmd.AddCommand(exportCmd) // Подключаем субкоманду export
    rootCmd.AddCommand(importCmd) // Подключаем субкоманду import

    // Устанавливаем шаблон вывода версии
    rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
    // Флаг -v или --version теперь работает автоматически
    // Не нужно вручную добавлять PersistentFlags().BoolP("version"...
    // Cobra сделает это сам
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

func initConfig() {
	// 1. Добавляем стандартные пути поиска
	home, err := os.UserHomeDir()
	if err != nil {
		// Не паникуем — просто продолжаем без home-пути
		fmt.Printf("Warning: cannot get user home directory: %v\n", err)
	} else {
		configDir := filepath.Join(home, ".gotr")
		viper.AddConfigPath(configDir) // ~/.gotr
	}

	// Удобно для локального тестирования
	viper.AddConfigPath(".") // текущая директория

	// 2. Имя файла без расширения (viper сам попробует .yaml, .json и т.д.)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml") // явно указываем yaml

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