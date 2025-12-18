package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gotr/internal/models/data"
	"log"
	"os"

	"github.com/spf13/viper"
)

// Вспомогательная функция PrettyYAML - для форматированного вывода конфигурации Kubernetes
func PrettyJSON(cfg *data.TestCase) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(true)
	err := enc.Encode(cfg)
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

// Вспомогательная функция '' - для интерактивного ввода
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
