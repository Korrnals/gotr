package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DebugPrint печатает сообщение только если --debug включён
func DebugPrint(format string, args ...interface{}) {
	if viper.GetBool("debug") {
		log.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// OpenEditor открывает файл в редакторе по умолчанию
func OpenEditor(filepath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Fallback в зависимости от ОС
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vi" // или "nano" — vim более универсальный
		}
		fmt.Printf("EDITOR not set. Using fallback: %s\n", editor)
	}

	// Создаём команду
	cmd := exec.Command(editor, filepath)

	// Привязываем stdin/stdout/stderr к терминалу
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Запускаем
	return cmd.Run()
}

// GetFieldValue возвращает строковое представление значения field структуры по имени.
// Поддерживает case-insensitive поиск (title/Title — найдёт оба варианта).
// Если поле не найдено — возвращает пустую строку.
//
// Пример:
//
//	val := GetFieldValue(someCase, "title")  // → "Мой тест-кейс"
//	val := GetFieldValue(someCase, "PriorityID") // → "3"
func GetFieldValue(obj interface{}, field string) string {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.IsValid() {
		return ""
	}

	// Прямой поиск по точному имени
	f := v.FieldByName(field)
	if f.IsValid() {
		return fmt.Sprintf("%v", f.Interface())
	}

	// Case-insensitive поиск
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		if strings.EqualFold(fieldName, field) {
			f = v.Field(i)
			if f.IsValid() {
				return fmt.Sprintf("%v", f.Interface())
			}
		}
	}

	return ""
}

// LoadMapping загружает mapping из файла. Поддерживает два формата:
// 1) простая JSON map[string]int64 — { "123": 456, ... }
// 2) структура SharedStepMapping (файл с полем pairs) — будет прочитан и преобразован в map[source_id]=target_id
func LoadMapping(file string) (map[int64]int64, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Попробуем сначала простой map[string]int64
	var m1 map[string]int64
	if err := json.Unmarshal(data, &m1); err == nil && len(m1) > 0 {
		res := make(map[int64]int64, len(m1))
		for k, v := range m1 {
			id, err := strconv.ParseInt(k, 10, 64)
			if err != nil {
				// если ключ не число — пропускаем
				continue
			}
			res[id] = v
		}
		return res, nil
	}

	// Попробуем формат с pairs
	var wrapper struct {
		Pairs []struct {
			SourceID int64 `json:"source_id"`
			TargetID int64 `json:"target_id"`
		} `json:"pairs"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Pairs) > 0 {
		res := make(map[int64]int64, len(wrapper.Pairs))
		for _, p := range wrapper.Pairs {
			res[p.SourceID] = p.TargetID
		}
		return res, nil
	}

	// Формат не распознан — вернуть пустой map и nil err
	return make(map[int64]int64), nil
}

// getProjectRoot — возвращает корневую директорию проекта (где лежит go.mod)
func getProjectRoot() string {
	// 1. Получаем путь к текущему файлу test
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get test file path")
	}

	// 2. Идем вверх по дереву папок, пока не найдем go.mod
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir // Корень найден
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			panic("project root (go.mod) not found")
		}
		dir = parent
	}
}

// LogDir — директория для логов (создается, если не существует)
func LogDir() string {
	// Теперь путь всегда вычисляется от корня, где лежит go.mod
	rootDir := getProjectRoot()

	logPath := filepath.Join(rootDir, ".testrail", "logs")

	if err := os.MkdirAll(logPath, 0755); err != nil {
		panic(err)
	}
	return logPath
}

// ParseID парсит строку в int64 (для ID из аргументов команд)
func ParseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// SaveToFile сохраняет данные в JSON файл с форматированием
func SaveToFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("serialization error: %w", err)
	}
	return os.WriteFile(filename, jsonData, 0644)
}

// OutputResult выводит данные в JSON и сохраняет в файл (если указан флаг --output)
// Используется в CLI командах для стандартизации вывода
func OutputResult(cmd *cobra.Command, data interface{}) error {
	quiet, _ := cmd.Flags().GetBool("quiet")
	output, _ := cmd.Flags().GetString("output")

	// Сохранение в файл если указан флаг
	if output != "" {
		if err := SaveToFile(output, data); err != nil {
			return err
		}
		if !quiet {
			fmt.Printf("Response saved to %s\n", output)
		}
	}

	// Вывод в консоль если не quiet режим
	if !quiet {
		pretty, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("JSON formatting error: %w", err)
		}
		fmt.Println(string(pretty))
	}

	return nil
}

// PrintSuccess выводит сообщение об успехе (если не quiet режим)
func PrintSuccess(cmd *cobra.Command, format string, args ...interface{}) {
	quiet, _ := cmd.Flags().GetBool("quiet")
	if !quiet {
		fmt.Printf(format+"\n", args...)
	}
}
