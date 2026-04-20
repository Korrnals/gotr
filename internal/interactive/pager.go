package interactive

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PagerConfig configures the interactive pager.
type PagerConfig struct {
	// Lines are the content lines to paginate.
	Lines []string
	// Header is printed at the top of every page (optional).
	Header string
	// PageSize overrides the number of lines per page.
	// If 0, auto-detects from terminal height.
	PageSize int
}

// Pager displays lines page by page in the terminal.
// Navigation: Enter/Space = next page, b = previous page, q = quit.
// If the output fits in one page, prints directly without pagination.
//nolint:gocyclo // Interactive key handling is intentionally centralized.
func Pager(cfg PagerConfig) error {
	if len(cfg.Lines) == 0 {
		return nil
	}

	pageSize := cfg.PageSize
	if pageSize <= 0 {
		pageSize = detectPageSize()
	}

	// Reserve space for header + status line.
	reservedLines := 2
	if cfg.Header != "" {
		reservedLines += strings.Count(cfg.Header, "\n") + 1
	}
	contentPerPage := pageSize - reservedLines
	if contentPerPage < 5 {
		contentPerPage = 5
	}

	totalPages := (len(cfg.Lines) + contentPerPage - 1) / contentPerPage

	// If everything fits in one page — just print.
	if totalPages <= 1 {
		if cfg.Header != "" {
			fmt.Println(cfg.Header)
		}
		for _, line := range cfg.Lines {
			fmt.Println(line)
		}
		return nil
	}

	// Switch terminal to raw mode for single-key input.
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		// Fallback: print everything.
		return pagerFallback(cfg)
	}
	defer func() {
		_ = term.Restore(fd, oldState)
	}()

	page := 0

	for {
		// Clear screen and render current page.
		fmt.Print("\033[2J\033[H") // clear + move to top

		if cfg.Header != "" {
			fmt.Println(cfg.Header)
		}

		start := page * contentPerPage
		end := start + contentPerPage
		if end > len(cfg.Lines) {
			end = len(cfg.Lines)
		}

		for _, line := range cfg.Lines[start:end] {
			fmt.Println(line)
		}

		// Status bar.
		fmt.Printf("\n\033[7m Page %d/%d │ [Enter/Space] next │ [b] prev │ [q] quit \033[0m", page+1, totalPages)

		// Read single key.
		buf := make([]byte, 3)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		switch {
		case n == 1 && (buf[0] == 'q' || buf[0] == 'Q'):
			fmt.Print("\033[2J\033[H") // clear before exit
			return nil
		case n == 1 && (buf[0] == 'b' || buf[0] == 'B'):
			if page > 0 {
				page--
			}
		case n == 1 && (buf[0] == '\r' || buf[0] == '\n' || buf[0] == ' '):
			if page < totalPages-1 {
				page++
			}
		case n == 1 && buf[0] == 3: // Ctrl+C
			fmt.Print("\033[2J\033[H")
			return nil
		// Arrow keys: ESC [ A/B
		case n == 3 && buf[0] == 27 && buf[1] == '[' && buf[2] == 'B': // Down
			if page < totalPages-1 {
				page++
			}
		case n == 3 && buf[0] == 27 && buf[1] == '[' && buf[2] == 'A': // Up
			if page > 0 {
				page--
			}
		}
	}

	return nil
}

// ShouldPage returns true if output should be paginated
// (stdout is a TTY and there are more lines than fit on screen).
func ShouldPage(lineCount int) bool {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	return lineCount > detectPageSize()-3
}

func detectPageSize() int {
	_, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || h < 10 {
		return 24 // safe default
	}
	return h
}

func pagerFallback(cfg PagerConfig) error {
	if cfg.Header != "" {
		fmt.Println(cfg.Header)
	}
	for _, line := range cfg.Lines {
		fmt.Println(line)
	}
	return nil
}
