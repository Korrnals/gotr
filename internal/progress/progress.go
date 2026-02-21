// Package progress provides progress bar functionality for long-running operations.
package progress

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// Manager управляет прогресс-барами для операций
type Manager struct {
	progress *mpb.Progress
	output   io.Writer
	quiet    bool
	enabled  bool
	wg       sync.WaitGroup
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

	if m.enabled {
		m.progress = mpb.New(
			mpb.WithOutput(m.output),
			mpb.WithRefreshRate(150*time.Millisecond),
		)
	}

	return m
}

// Bar представляет прогресс-бар
type Bar struct {
	bar    *mpb.Bar
	total  int64
	mgr    *Manager
}

// NewBar создаёт прогресс-бар с общим количеством
func (m *Manager) NewBar(total int64, description string) *Bar {
	if m == nil || !m.enabled || m.progress == nil {
		return nil
	}

	m.wg.Add(1)

	mpbBar := m.progress.AddBar(total,
		mpb.PrependDecorators(
			// Название операции
			decor.Name(description, decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			// Процент выполнения
			decor.Percentage(decor.WCSyncSpace),
			// Прогресс (текущее/всего)
			decor.CountersNoUnit("(%d/%d)", decor.WCSyncWidth),
			// Скорость
			decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 60, decor.WCSyncSpace),
			// ETA
			decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 60, decor.WCSyncWidth), ""),
		),
	)

	return &Bar{
		bar:   mpbBar,
		total: total,
		mgr:   m,
	}
}

// NewSpinner создаёт спиннер для неопределённых операций
func (m *Manager) NewSpinner(description string) *Bar {
	if m == nil || !m.enabled || m.progress == nil {
		return nil
	}

	m.wg.Add(1)

	// Для спиннера используем bar с индетерминированным прогрессом
	// Используем заполняющийся бар как спиннер
	mpbBar := m.progress.AddBar(100,
		mpb.BarFillerClearOnComplete(),
		mpb.PrependDecorators(
			decor.Name(description, decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			// Простой декоратор для спиннера
			decor.Name("...", decor.WCSyncWidth),
		),
	)

	return &Bar{
		bar:   mpbBar,
		total: 100,
		mgr:   m,
	}
}

// Add увеличивает прогресс-бар на n
func (b *Bar) Add(n int) {
	if b != nil && b.bar != nil {
		b.bar.IncrBy(n)
	}
}

// Increment увеличивает прогресс-бар на 1
func (b *Bar) Increment() {
	if b != nil && b.bar != nil {
		b.bar.Increment()
	}
}

// Finish завершает прогресс-бар
func (b *Bar) Finish() {
	if b != nil && b.bar != nil {
		if b.total > 0 {
			b.bar.SetTotal(b.total, true)
		} else {
			// Для спиннера просто помечаем как завершённый
			b.bar.Abort(false)
		}
		if b.mgr != nil {
			b.mgr.wg.Done()
		}
	}
}

// Describe обновляет описание прогресс-бара
func (b *Bar) Describe(description string) {
	if b != nil && b.bar != nil {
		// mpb не поддерживает динамическое изменение описания напрямую
		// но мы можем использовать decor.Name с динамическим обновлением
	}
}

// Wait ожидает завершения всех прогресс-баров
func (m *Manager) Wait() {
	if m != nil && m.progress != nil {
		m.wg.Wait()
		m.progress.Wait()
	}
}

// Add увеличивает прогресс-бар (устаревшая функция для совместимости)
func Add(bar *Bar, n int) {
	if bar != nil {
		bar.Add(n)
	}
}

// Finish завершает прогресс-бар (устаревшая функция для совместимости)
func Finish(bar *Bar) {
	if bar != nil {
		bar.Finish()
	}
}

// Describe обновляет описание прогресс-бара (устаревшая функция для совместимости)
func Describe(bar *Bar, description string) {
	if bar != nil {
		bar.Describe(description)
	}
}
