// Package bundle provides portable tar.gz and zip archive helpers with a
// common manifest.json and SHA256SUMS layout used by `gotr export/import
// snap` and `gotr export/import report`.
//
// A bundle contains:
//   - manifest.json   — machine-readable metadata (schema version, gotr
//     version, kind, file list with SHA-256 and relative-to-home paths).
//   - SHA256SUMS      — GNU-format checksum file for external verification.
//   - README.txt      — human-readable summary.
//   - payload files   — archived with forward-slash paths and deterministic
//     modes so bundles are reproducible across machines.
package bundle
