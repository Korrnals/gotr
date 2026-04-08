// Package output centralizes all response output logic for gotr commands,
// supporting multiple formats (JSON, CSV, YAML, plain text) and optional
// file export to ~/.gotr/exports/.
//
// Commands add the --save flag via [AddFlag] and call [Output] or
// [OutputGetResult] to render responses consistently. [OutputResultWithFlags]
// handles root-level --output and --quiet flags for direct JSON output
// with optional file saving.
//
// The [DryRunPrinter] provides structured output for --dry-run operations,
// showing what would be changed without executing.
//
// This design ensures format changes, export logic, and save-path prompting
// remain in one place rather than scattered across command files.
package output
