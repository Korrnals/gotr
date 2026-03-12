// internal/models/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/ui"
)

// DefaultConfigValues — дефолтные placeholder'ы в шаблоне конфигурации.
// Эти значения используются как creating конфига, так и для проверки валидности.
const (
	DefaultBaseURL  = "https://yourcompany.testrail.io/"
	DefaultUsername = "your-email@example.com"
	DefaultAPIKey   = "your_api_key_here"
)

type ConfigData struct {
	BaseURL  string `yaml:"base_url"`
	Username string `yaml:"username"`
	APIKey   string `yaml:"api_key"`
	Insecure bool   `yaml:"insecure"`
	JqFormat bool   `yaml:"jq_format"`
	Debug    bool   `yaml:"debug"`
}

// Config — представляет один конфиг-файл
type Config struct {
	Path string      // полный путь к файлу
	Data *ConfigData // данные (можно расширять)
}

// New создаёт экземпляр конфига по произвольному пути
func New(path string) *Config {
	return &Config{
		Path: path,
		Data: &ConfigData{}, // пустой или с дефолтами
	}
}

// Default возвращает конфиг по стандартному пути (~/.gotr/config/default.yaml)
func Default() (*Config, error) {
	// Используем centralized paths
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".gotr", "config", "default.yaml")
	return New(path), nil
}

// WithDefaults заполняет дефолтными значениями
func (c *Config) WithDefaults() *Config {
	c.Data = &ConfigData{
		BaseURL:  DefaultBaseURL,
		Username: DefaultUsername,
		APIKey:   DefaultAPIKey,
		Insecure: false,
		JqFormat: false,
		Debug:    false,
	}
	return c
}

// Create создаёт файл на диске
func (c *Config) Create() error {
	dir := filepath.Dir(c.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	content := []byte(c.renderTemplate())

	if err := os.WriteFile(c.Path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", c.Path, err)
	}

	ui.Infof(os.Stdout, "Config file created: %s", c.Path)
	return nil
}

func (c *Config) renderTemplate() string {
	data := c.Data
	if data == nil {
		data = (&Config{}).WithDefaults().Data
	}

	return fmt.Sprintf(`# gotr configuration file
#
# Приоритет источников:
#   1) CLI flags
#   2) Environment variables (TESTRAIL_*)
#   3) Этот файл

# Базовый URL TestRail.
# Пример cloud:  https://yourcompany.testrail.io
# Пример server: https://testrail.example.local
base_url: %q

# Логин (обычно email пользователя TestRail).
username: %q

# API key пользователя TestRail.
api_key: %q

# true  -> пропустить проверку TLS сертификата (небезопасно, только для внутренних стендов)
# false -> стандартная безопасная проверка TLS
insecure: %v

# Включить jq-форматирование вывода (если встроенный jq доступен в системе).
jq_format: %v

# Включить отладочный вывод gotr.
debug: %v

compare:
  # Режим окружения для compare-запросов:
  #   auto   - попытка определить по URL (cloud/server)
  #   cloud  - принудительно cloud-профиль
  #   server - принудительно server-профиль
  deployment: "auto"

  # Для cloud-профиля: professional|enterprise
  cloud_tier: "professional"

  # Глобальный лимит запросов в минуту для compare.
  #   -1 -> автоматически по профилю (cloud/server)
  #    0 -> лимит выключен
  #   >0 -> фиксированное значение req/min
  rate_limit: -1

  # Дефолт для cloud, если rate_limit=-1.
  # professional: 180, enterprise: 300
  cloud_rate_limit: 300

  # Дефолт для server, если rate_limit=-1.
  # Обычно 0 (без лимита).
  server_rate_limit: 0

  cases:
    # Параллельность по сьютам (между сьютами).
    parallel_suites: 10

    # Параллельность страниц внутри одного сьюта.
    parallel_pages: 6

    # Количество retry для каждой страницы в основном этапе загрузки compare cases.
    page_retries: 5

    # Таймаут полной операции compare cases.
    timeout: "30m"

    retry:
      # Попытки дозабора одной failed-страницы.
      attempts: 5

      # Количество параллельных воркеров дозабора.
      workers: 12

      # Пауза между попытками дозабора одной страницы.
      delay: "200ms"

    # Всегда пытаться автоматически дозабирать failed pages после основного compare cases.
    auto_retry_failed_pages: true
`, data.BaseURL, data.Username, data.APIKey, data.Insecure, data.JqFormat, data.Debug)
}

// Path возвращает путь (для подкоманды path)
func (c *Config) PathString() string {
	return c.Path
}

// IsValid проверяет, что конфиг содержит реальные данные, а не дефолтные placeholder'ы
func (c *Config) IsValid() bool {
	if c.Data == nil {
		return false
	}
	return c.Data.BaseURL != "" && c.Data.BaseURL != DefaultBaseURL &&
		c.Data.Username != "" && c.Data.Username != DefaultUsername &&
		c.Data.APIKey != "" && c.Data.APIKey != DefaultAPIKey
}

// IsDefaultValue проверяет, является ли значение дефолтным placeholder'ом
func IsDefaultValue(value, defaultValue string) bool {
	return value == "" || value == defaultValue
}
