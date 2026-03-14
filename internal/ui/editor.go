package ui

import (
	"os"
	"os/exec"
	"runtime"
)

// OpenEditor opens a file in the configured editor (EDITOR env var).
func OpenEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vi"
		}
		Warningf(os.Stdout, "EDITOR not set. Using fallback: %s", editor)
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
