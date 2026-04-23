package report

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Korrnals/gotr/internal/paths"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <filename|latest>",
		Short: "Open a migration report in the OS default viewer (PDF) or print it (md/json)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("report show: resolve reports dir: %w", err)
			}
			target := args[0]
			if target == "latest" {
				latest, err := resolveLatestReport(reportsDir)
				if err != nil {
					return fmt.Errorf("report show: %w", err)
				}
				target = latest
			}
			path := filepath.Join(reportsDir, target)
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("report show: %w", err)
			}

			ext := strings.ToLower(filepath.Ext(target))
			switch ext {
			case ".pdf":
				if err := openWithOS(path); err != nil {
					return fmt.Errorf("report show: open PDF: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Opened %s\n", path)
				return nil
			case ".md", ".json", ".txt", "":
				return catFile(cmd.OutOrStdout(), path)
			default:
				// Best-effort: fall back to OS opener for unknown extensions.
				if err := openWithOS(path); err != nil {
					return fmt.Errorf("report show: open %s: %w", ext, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Opened %s\n", path)
				return nil
			}
		},
	}
}

func catFile(w io.Writer, path string) error {
	data, err := os.ReadFile(path) //nolint:gosec // path is inside reports dir resolved from paths helper
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// openWithOS invokes the OS default file opener: xdg-open on Linux, open on
// macOS, rundll32 on Windows. Returns an error if no opener is available.
func openWithOS(path string) error {
	var (
		bin  string
		args []string
	)
	switch runtime.GOOS {
	case "linux":
		bin = "xdg-open"
		args = []string{path}
	case "darwin":
		bin = "open"
		args = []string{path}
	case "windows":
		bin = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", path}
	default:
		return errors.New("unsupported platform for `report show` — cat the file manually")
	}
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("opener %q not found in PATH", bin)
	}
	return exec.Command(bin, args...).Start() //nolint:gosec // bin is a fixed per-OS constant
}
