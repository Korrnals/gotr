package reporter

// Пример использования go-pretty/v6 для отрисовки статистики.
// Этот файл добавлен для сравнения с кастомным рендером (reporter.go).
// Пока НЕ подключён к основному коду — только для тестирования.
//
// Использование:
//
//	reporter.NewPretty("cases").
//	    Section("Общая статистика").
//	    Stat("⏱️", "Время выполнения", elapsed).
//	    PrintPretty()

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// PrettyReport — аналог Report, но рендерит через go-pretty.
type PrettyReport struct {
	title    string
	elements []element
	w        io.Writer
}

// NewPretty создаёт отчёт с go-pretty рендером.
func NewPretty(title string) *PrettyReport {
	return &PrettyReport{
		title: title,
		w:     os.Stdout,
	}
}

// Writer устанавливает вывод (по умолчанию os.Stdout).
func (r *PrettyReport) Writer(w io.Writer) *PrettyReport {
	r.w = w
	return r
}

// Section добавляет заголовок секции.
func (r *PrettyReport) Section(name string) *PrettyReport {
	if len(r.elements) > 0 {
		r.elements = append(r.elements, element{typ: elemBlank})
	}
	r.elements = append(r.elements, element{typ: elemSection, label: name})
	return r
}

// Stat добавляет строку статистики.
func (r *PrettyReport) Stat(icon, label string, value interface{}) *PrettyReport {
	r.elements = append(r.elements, element{
		typ:   elemStat,
		icon:  icon,
		label: label,
		value: fmt.Sprintf("%v", value),
	})
	return r
}

// StatIf добавляет строку только при выполнении условия.
func (r *PrettyReport) StatIf(cond bool, icon, label string, value interface{}) *PrettyReport {
	if cond {
		return r.Stat(icon, label, value)
	}
	return r
}

// StatFmt добавляет строку с форматированным значением.
func (r *PrettyReport) StatFmt(icon, label, format string, args ...interface{}) *PrettyReport {
	r.elements = append(r.elements, element{
		typ:   elemStat,
		icon:  icon,
		label: label,
		value: fmt.Sprintf(format, args...),
	})
	return r
}

// Blank добавляет пустую строку.
func (r *PrettyReport) Blank() *PrettyReport {
	r.elements = append(r.elements, element{typ: elemBlank})
	return r
}

// Separator добавляет разделитель.
func (r *PrettyReport) Separator() *PrettyReport {
	r.elements = append(r.elements, element{typ: elemSeparator})
	return r
}

// Print рендерит отчёт через go-pretty.
func (r *PrettyReport) Print() {
	fmt.Fprint(r.w, r.String())
}

// String рендерит отчёт в строку через go-pretty table.
func (r *PrettyReport) String() string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)

	// Заголовок таблицы.
	tw.SetTitle(fmt.Sprintf("📊 СТАТИСТИКА: %s", r.title))

	// Настройка: одна колонка, выравнивание по левому краю.
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignLeft, WidthMin: 55},
	})

	for _, e := range r.elements {
		switch e.typ {
		case elemSection:
			tw.AppendSeparator()
			tw.AppendRow(table.Row{fmt.Sprintf("◆ %s", e.label)})
		case elemStat:
			tw.AppendRow(table.Row{fmt.Sprintf("%s %s: %s", e.icon, e.label, e.value)})
		case elemSeparator:
			tw.AppendSeparator()
		case elemBlank:
			tw.AppendRow(table.Row{""})
		}
	}

	return "\n" + tw.Render() + "\n"
}

// CompareStatsPretty — аналог CompareStats для go-pretty.
func CompareStatsPretty(resource string, pid1, pid2 int64, onlyFirst, onlySecond, common int, elapsed time.Duration) *PrettyReport {
	total := onlyFirst + onlySecond + common
	return NewPretty(resource).
		Section("Общая статистика").
		Stat("⏱️", "Время выполнения", elapsed.Round(time.Millisecond)).
		Stat("📦", "Всего обработано", total).
		Section("Результат сравнения").
		Stat("✅", fmt.Sprintf("Только в проекте %d", pid1), onlyFirst).
		Stat("✅", fmt.Sprintf("Только в проекте %d", pid2), onlySecond).
		Stat("🔗", "Общих", common)
}
