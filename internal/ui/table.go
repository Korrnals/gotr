// Package ui — см. display.go.
// Файл table.go — статический вывод финальных results (таблицы + JSON).
//
// Концепция:
//
//	display.go — live progress (ANSI, refresh loop, Task reporter)
//	table.go   — статический вывод: Table(), JSON(), NewTable()
//
// Использование:
//
//	// Создать и заполнить таблицу:
//	t := ui.NewTable(cmd)
//	t.AppendHeader(table.Row{"ID", "Имя"})
//	t.AppendRow(table.Row{1, "foo"})
//	ui.Table(cmd, t)    // вывод в формате --format (table/json/csv/md)
//
//	// Вывод произвольного значения как JSON:
//	if err := ui.JSON(cmd, myStruct); err != nil { return err }
package ui

import (
	"encoding/json"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// OutputFormat — допустимые значения флага --format.
type OutputFormat string

const (
	FormatTable    OutputFormat = "table"
	FormatJSON     OutputFormat = "json"
	FormatCSV      OutputFormat = "csv"
	FormatMarkdown OutputFormat = "md"
	FormatHTML     OutputFormat = "html"
)

// NewTable создаёт go-pretty table.Writer с базовой конфигурацией:
//   - вывод направлен в cmd.OutOrStdout()
//   - стиль: StyleRounded (рамки с закруглёнными углами)
//
// Вызов Table(cmd, t) затем выводит таблицу в нужном формате.
func NewTable(cmd *cobra.Command) table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	t.SetStyle(table.StyleRounded)
	return t
}

// Table выводит таблицу в формате, указанном флагом --format команды.
// Если флага нет или значение неизвестно — выводит как таблицу (StyleRounded).
//
// Поддерживаемые форматы: table (умолч.), csv, md, html.
// Для вывода в JSON — используй ui.JSON(cmd, yourSlice).
func Table(cmd *cobra.Command, t table.Writer) {
	t.SetOutputMirror(cmd.OutOrStdout())

	format := getFormat(cmd)
	switch format {
	case FormatCSV:
		fmt.Fprintln(cmd.OutOrStdout(), t.RenderCSV())
	case FormatMarkdown:
		fmt.Fprintln(cmd.OutOrStdout(), t.RenderMarkdown())
	case FormatHTML:
		fmt.Fprintln(cmd.OutOrStdout(), t.RenderHTML())
	default:
		// table, json или неизвестный формат — выводим как таблицу
		fmt.Fprintln(cmd.OutOrStdout(), t.Render())
	}
}

// JSON сериализует v в JSON с отступами и печатает в cmd.OutOrStdout().
// Используется для не-табличных данных (объекты, массивы, raw API-ответы).
func JSON(cmd *cobra.Command, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("ui.JSON: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

// IsJSON возвращает true если команда запрошена с --format json.
// Удобно для условной логики в команде:
//
//	if ui.IsJSON(cmd) { /* вывод raw */ } else { /* таблица */ }
func IsJSON(cmd *cobra.Command) bool {
	return getFormat(cmd) == FormatJSON
}

// IsQuiet возвращает true если установлен флаг --quiet.
func IsQuiet(cmd *cobra.Command) bool {
	q, _ := cmd.Flags().GetBool("quiet")
	return q
}

// getFormat читает --format из флагов команды.
// Ищет сначала в локальных флагах, затем в inherited (PersistentFlags родителей).
func getFormat(cmd *cobra.Command) OutputFormat {
	if f := cmd.Flags().Lookup("format"); f != nil {
		return OutputFormat(f.Value.String())
	}
	if f := cmd.InheritedFlags().Lookup("format"); f != nil {
		return OutputFormat(f.Value.String())
	}
	return FormatTable
}
