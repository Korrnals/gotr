package report

import (
	"errors"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/interactive"
	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ErrNoInteractiveTarget is returned when interactive selection is requested
// but no report files are available to pick from.
var ErrNoInteractiveTarget = errors.New("no reports found to choose from")

// shouldPromptForReport reports whether the command should offer an interactive
// survey prompt in place of a missing positional argument. The decision is
// made based on:
//   - ctx holds a non-nil Prompter and is NOT a *NonInteractivePrompter
//   - stdin is an actual terminal (pipes/CI return false)
//
// It is safe to call even when the context has no prompter attached.
func shouldPromptForReport(cmd *cobra.Command) bool {
	ctx := cmd.Context()
	if interactive.IsNonInteractive(ctx) {
		return false
	}
	if !interactive.HasPrompterInContext(ctx) {
		// When no prompter was injected, assume a terminal prompter is
		// acceptable only if stdin looks like a TTY.
		return term.IsTerminal(int(os.Stdin.Fd()))
	}
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// promptForReport asks the user to pick a report from reportsDir. The first
// option is always "latest" (resolves via ResolveReportPath), followed by all
// reports ordered newest-first.
func promptForReport(cmd *cobra.Command, reportsDir string) (string, error) {
	entries, err := intreport.RecursiveListReports(reportsDir)
	if err != nil {
		return "", fmt.Errorf("list reports: %w", err)
	}
	if len(entries) == 0 {
		return "", ErrNoInteractiveTarget
	}

	options := make([]string, 0, len(entries)+1)
	options = append(options, "latest")
	for _, e := range entries {
		options = append(options, e.Rel)
	}

	p := interactive.PrompterFromContext(cmd.Context())
	_, value, err := p.Select("Select report", options)
	if err != nil {
		return "", err
	}
	return value, nil
}
