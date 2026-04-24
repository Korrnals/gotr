package report

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Korrnals/gotr/internal/paths"
	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "show [filename|latest]",
		Short:             "Open a migration report in the OS default viewer (PDF) or print it (md/json)",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeReportArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("report show: resolve reports dir: %w", err)
			}
			target, err := resolveShowTarget(cmd, args, reportsDir)
			if err != nil {
				return err
			}
			path, err := intreport.ResolveReportPath(reportsDir, target)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return fmt.Errorf("report show: %q not found under %s", target, reportsDir)
				}
				return fmt.Errorf("report show: %w", err)
			}

			ext := strings.ToLower(filepath.Ext(path))
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
	data, err := os.ReadFile(path) //nolint:gosec // path resolved via ResolveReportPath under reports dir
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

// resolveShowTarget returns the user-supplied positional argument or, when
// missing, asks the user to pick a report interactively. When running in
// non-interactive mode / no TTY / no args, it reports a usage error.
func resolveShowTarget(cmd *cobra.Command, args []string, reportsDir string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if !shouldPromptForReport(cmd) {
		return "", fmt.Errorf("report show: a report name or 'latest' is required (pass as argument or run interactively)")
	}
	return promptForReport(cmd, reportsDir)
}
