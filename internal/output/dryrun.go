// Package output provides utilities for output formatting including dry-run mode.
package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// DryRunPrinter handles dry-run output formatting.
type DryRunPrinter struct {
	Command string
}

// NewDryRunPrinter creates a new dry-run printer for the given command.
func NewDryRunPrinter(command string) *DryRunPrinter {
	return &DryRunPrinter{Command: command}
}

// PrintOperation displays what operation would be performed.
func (p *DryRunPrinter) PrintOperation(operation, method, url string, body interface{}) {
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "                    DRY RUN MODE")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintf(os.Stderr, "Command:    %s\n", p.Command)
	fmt.Fprintf(os.Stderr, "Operation:  %s\n", operation)
	fmt.Fprintf(os.Stderr, "HTTP Method: %s\n", method)
	fmt.Fprintf(os.Stderr, "Endpoint:   %s\n", url)
	fmt.Fprintln(os.Stderr, "───────────────────────────────────────────────────────────")
	fmt.Fprintln(os.Stderr, "Request Body:")

	if body != nil {
		jsonBytes, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error marshaling: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, string(jsonBytes))
		}
	} else {
		fmt.Fprintln(os.Stderr, "  (no body)")
	}

	fmt.Fprintln(os.Stderr, "───────────────────────────────────────────────────────────")
	fmt.Fprintln(os.Stderr, "Result:     No changes made (dry-run mode)")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
}

// PrintSimple displays a simple dry-run message for operations without body.
func (p *DryRunPrinter) PrintSimple(operation, description string) {
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "                    DRY RUN MODE")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintf(os.Stderr, "Command:     %s\n", p.Command)
	fmt.Fprintf(os.Stderr, "Operation:   %s\n", operation)
	fmt.Fprintf(os.Stderr, "Description: %s\n", description)
	fmt.Fprintln(os.Stderr, "───────────────────────────────────────────────────────────")
	fmt.Fprintln(os.Stderr, "Result:      No changes made (dry-run mode)")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
}

// PrintBatch displays batch operations (like sync).
func (p *DryRunPrinter) PrintBatch(operation string, items []string) {
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "                    DRY RUN MODE")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintf(os.Stderr, "Command:   %s\n", p.Command)
	fmt.Fprintf(os.Stderr, "Operation: %s\n", operation)
	fmt.Fprintln(os.Stderr, "───────────────────────────────────────────────────────────")
	fmt.Fprintf(os.Stderr, "Would process %d items:\n", len(items))
	for i, item := range items {
		if i >= 10 {
			fmt.Fprintf(os.Stderr, "  ... and %d more\n", len(items)-10)
			break
		}
		fmt.Fprintf(os.Stderr, "  • %s\n", item)
	}
	fmt.Fprintln(os.Stderr, "───────────────────────────────────────────────────────────")
	fmt.Fprintln(os.Stderr, "Result:    No changes made (dry-run mode)")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
}

// FormatBodyForDisplay returns a formatted JSON string for display.
func FormatBodyForDisplay(body interface{}) string {
	if body == nil {
		return "(no body)"
	}
	jsonBytes, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return fmt.Sprintf("(error: %v)", err)
	}
	return string(jsonBytes)
}

// PrintSummary displays a summary of what would happen.
func (p *DryRunPrinter) PrintSummary(actions []string) {
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "                    DRY RUN SUMMARY")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintf(os.Stderr, "Command: %s\n", p.Command)
	fmt.Fprintln(os.Stderr, "───────────────────────────────────────────────────────────")
	fmt.Fprintf(os.Stderr, "Planned actions (%d):\n", len(actions))
	for i, action := range actions {
		fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, action)
	}
	fmt.Fprintln(os.Stderr, "───────────────────────────────────────────────────────────")
	fmt.Fprintln(os.Stderr, "Status: NO CHANGES MADE (dry-run mode)")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
}

// PrintValidationError displays validation errors in dry-run mode.
func (p *DryRunPrinter) PrintValidationError(err error) {
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "                    DRY RUN VALIDATION ERROR")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintf(os.Stderr, "Command: %s\n", p.Command)
	fmt.Fprintf(os.Stderr, "Error:   %v\n", err)
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
}
