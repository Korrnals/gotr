package embed

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

//go:embed jq-linux-amd64
var jqLinux []byte

//go:embed jq-macos-amd64
var jqMac []byte

//go:embed jq-windows-i386.exe
var jqWindows []byte

// RunEmbeddedJQ — запускает встроенный jq с фильтром
func RunEmbeddedJQ(rawBody []byte, filterStr string) error {
	if filterStr == "" {
		filterStr = "."
	}

	// Выбираем бинарник
	var jqBin []byte
	switch runtime.GOOS {
	case "linux":
		jqBin = jqLinux
	case "darwin":
		jqBin = jqMac
	case "windows":
		jqBin = jqWindows
	default:
		return fmt.Errorf("{jq_embed} - платформа %s не поддерживается встроенным jq", runtime.GOOS)
	}

	// Создаём временный файл в текущей директории
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, "jq-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close() // закрываем, чтобы избежать "text file busy"

	// Записываем бинарник
	if err := os.WriteFile(tmpPath, jqBin, 0644); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Явно устанавливаем права на исполнение
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("{jq_embed} - не удалось установить права на исполнение: %w", err)
	}

	// Запускаем jq
	cmd := exec.Command(tmpPath, filterStr)
	cmd.Stdin = bytes.NewReader(rawBody)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("{jq_embed} - ошибка встроенного jq: %w", err)
	}

	// Удаляем файл
	os.Remove(tmpPath)

	return nil
}