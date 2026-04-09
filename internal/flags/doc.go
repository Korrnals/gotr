// Package flags provides lightweight helpers for parsing IDs and reading
// Cobra command flags consistently across the CLI.
//
// It supports positional argument validation (for interactive use) and
// flag-based retrieval with sensible defaults. All functions return
// structured errors suitable for CLI output. Used throughout cmd/ to
// eliminate repetitive flag-reading boilerplate and enforce consistent
// validation when resource IDs are required.
//
// Key functions: [ParseID], [ParseIDFromArgs], [ValidateRequiredID],
// [GetFlagInt64], [GetFlagString], [GetFlagBool].
package flags
