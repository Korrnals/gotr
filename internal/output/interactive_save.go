package output

import (
	"fmt"
	"strings"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

const defaultSavePathMarker = "__DEFAULT__"

// ResolveSavePathFromFlags resolves save target from explicit save flags.
// Returns:
//   - savePath ("", "__DEFAULT__", or custom path)
//   - explicit=true when user explicitly set --save or --save-to
func ResolveSavePathFromFlags(cmd *cobra.Command) (savePath string, explicit bool, err error) {
	if cmd.Flags().Changed("save-to") {
		savePath, err = cmd.Flags().GetString("save-to")
		if err != nil {
			return "", false, fmt.Errorf("failed to read --save-to flag: %w", err)
		}
		return savePath, true, nil
	}

	if cmd.Flags().Changed("save") {
		return defaultSavePathMarker, true, nil
	}

	return "", false, nil
}

// PromptSavePath asks whether to save command result and where to save it.
// Returns "" (do not save), "__DEFAULT__" (default exports path), or custom path.
func PromptSavePath(p interactive.Prompter, subject string) (string, error) {
	if subject == "" {
		subject = "result"
	}

	saveToFile, err := p.Confirm(fmt.Sprintf("Save %s to file?", subject), false)
	if err != nil {
		return "", fmt.Errorf("interactive save selection failed: %w", err)
	}

	if !saveToFile {
		return "", nil
	}

	useDefaultPath, err := p.Confirm("Use default export path (~/.gotr/exports)?", true)
	if err != nil {
		return "", fmt.Errorf("interactive save path selection failed: %w", err)
	}

	if useDefaultPath {
		return defaultSavePathMarker, nil
	}

	customPath, err := p.Input(fmt.Sprintf("Enter output file path for %s (e.g. compare_result.json):", subject), "")
	if err != nil {
		return "", fmt.Errorf("interactive output path input failed: %w", err)
	}

	customPath = strings.TrimSpace(customPath)
	if customPath == "" {
		return "", fmt.Errorf("interactive output path is empty")
	}

	return customPath, nil
}