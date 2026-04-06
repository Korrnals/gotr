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

var selectEmbeddedJQBinaryFunc = selectEmbeddedJQBinary
var writeEmbeddedBinaryFile = os.WriteFile

func selectEmbeddedJQBinary(goos string) ([]byte, error) {
	switch goos {
	case "linux":
		return jqLinux, nil
	case "darwin":
		return jqMac, nil
	case "windows":
		return jqWindows, nil
	default:
		return nil, fmt.Errorf("{jq_embed} - платформа %s не поддерживается встроенным jq", goos)
	}
}

// RunEmbeddedJQ runs the embedded jq binary with the given filter.
func RunEmbeddedJQ(rawBody []byte, filterStr string) error {
	if filterStr == "" {
		filterStr = "."
	}

	jqBin, err := selectEmbeddedJQBinaryFunc(runtime.GOOS)
	if err != nil {
		return err
	}

	// Create a temp file in the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, "jq-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close() // close before writing to avoid "text file busy"

	// Write the embedded binary to the temp file
	if err := writeEmbeddedBinaryFile(tmpPath, jqBin, 0644); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Explicitly set the executable permission
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("{jq_embed} - не удалось установить права на исполнение: %w", err)
	}

	// Run jq
	cmd := exec.Command(tmpPath, filterStr)
	cmd.Stdin = bytes.NewReader(rawBody)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("{jq_embed} - ошибка встроенного jq: %w", err)
	}

	// Clean up the temp binary
	os.Remove(tmpPath)

	return nil
}
