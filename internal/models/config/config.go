// internal/models/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

type ConfigData struct {
	BaseURL  string `yaml:"base_url"`
	Username string `yaml:"username"`
	APIKey   string `yaml:"api_key"`
	Insecure bool   `yaml:"insecure"`
	JqFormat bool   `yaml:"jq_format"`
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
		BaseURL:  "https://yourcompany.testrail.io/",
		Username: "your-email@example.com",
		APIKey:   "your_api_key_here",
		Insecure: false,
		JqFormat: false,
	}
	return c
}

// Create создаёт файл на диске
func (c *Config) Create() error {
	dir := filepath.Dir(c.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию: %w", err)
	}

	header := []byte("# gotr configuration file\n# Отредактируйте под свои данные\n\n")
	yamlData, _ := yaml.Marshal(c.Data)
	content := append(header, yamlData...)

	if err := os.WriteFile(c.Path, content, 0644); err != nil {
		return fmt.Errorf("не удалось записать файл %s: %w", c.Path, err)
	}

	fmt.Printf("Создан конфиг-файл: %s\n", c.Path)
	return nil
}

// Path возвращает путь (для подкоманды path)
func (c *Config) PathString() string {
	return c.Path
}
