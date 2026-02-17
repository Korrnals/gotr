// Package progress provides progress bar functionality for long-running operations.
package progress

import (
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
)

// Manager управляет прогресс-барами для операций
type Manager struct {
	output  io.Writer
	quiet   bool
	enabled bool
}

// Option конфигурирует Manager
type Option func(*Manager)

// WithOutput устанавливает вывод для прогресс-баров
func WithOutput(w io.Writer) Option {
	return func(m *Manager) {
		m.output = w
	}
}

// WithQuiet отключает вывод прогресс-баров
func WithQuiet(quiet bool) Option {
	return func(m *Manager) {
		m.quiet = quiet
		m.enabled = !quiet
	}
}

// NewManager создаёт новый Manager
func NewManager(opts ...Option) *Manager {
	m := &Manager{
		output:  os.Stderr,
		enabled: true,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// NewBar создаёт прогресс-бар с общим количеством
func (m *Manager) NewBar(total int64, description string) *progressbar.ProgressBar {
	if m == nil || !m.enabled {
		return nil
	}

	return progressbar.NewOptions64(total,
		progressbar.OptionSetWriter(m.output),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("items"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "|",
			BarEnd:        "|",
		}),
	)
}

// NewSpinner создаёт спиннер для неопределённых операций
func (m *Manager) NewSpinner(description string) *progressbar.ProgressBar {
	if m == nil || !m.enabled {
		return nil
	}

	return progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(m.output),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionClearOnFinish(),
	)
}

// Add увеличивает прогресс-бар
func Add(bar *progressbar.ProgressBar, n int) {
	if bar != nil {
		bar.Add(n)
	}
}

// Finish завершает прогресс-бар
func Finish(bar *progressbar.ProgressBar) {
	if bar != nil {
		bar.Finish()
	}
}

// Describe обновляет описание прогресс-бара
func Describe(bar *progressbar.ProgressBar, description string) {
	if bar != nil {
		bar.Describe(description)
	}
}
