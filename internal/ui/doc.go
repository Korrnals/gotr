// Package ui provides a unified display layer for gotr, consolidating
// all terminal I/O into one package with consistent styling and behavior.
//
// The [Display] type manages a live-updating terminal area with real-time
// task statistics (cases received, speed, elapsed time), refreshing via
// ANSI escape codes without scrolling. Tasks report progress through
// ProgressReporter methods (OnCasesReceived, OnPageFetched, OnSuiteComplete).
//
// Static helpers ([Table], [JSON], [Infof], [Successf], [Warningf]) format
// final output in the user's preferred format (table/csv/markdown/html).
// Editor integration via [OpenEditor] enables inline file editing for
// workflow commands.
package ui
