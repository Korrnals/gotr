package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/viper"
)

// Вспомогательная функция PrettyYAML - для форматированного ответа
func PrettyJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(true)
	err := enc.Encode(data)
	if err != nil {
		return err
	}

	return nil
}

// DebugPrint печатает сообщение только если --debug включён
func DebugPrint(format string, args ...interface{}) {
	if viper.GetBool("debug") {
		log.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// Вспомогательная функция 'StringPrompt' - для интерактивного ввода
func StringPrompt(q string) bool {
	var s string
	// Считываем строку в буфер
	r := bufio.NewReader(os.Stdin)
	// 3 раза задаем вопрос, и при пустом ответе завершаем функцию
	for i := 0; i < 3; i++ {
		fmt.Fprint(os.Stderr, q + " ")
		s, _ = r.ReadString('\n')
		if s != "" {
			fmt.Println("Выполняется..")
			break
		}
		if i == 2 && s == "" {
			fmt.Println("Выполнение действия - отменено.")
			break
		}
	}
	return true
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
        fmt.Printf("Переменная EDITOR не задана. Используется fallback: %s\n", editor)
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

// UnixToTime конвертирует Unix timestamp в time.Time
func UnixToTime(ts int64) time.Time {
	return time.Unix(ts, 0)
}

// TimeToUnix конвертирует time.Time в Unix timestamp
func TimeToUnix(t time.Time) int64 {
	return t.Unix()
}