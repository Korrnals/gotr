// Package reportbundle implements portable zip export/import of gotr
// migration reports (markdown/PDF/JSON) living in ~/.gotr/reports/.
package reportbundle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/bundle"
	"github.com/Korrnals/gotr/internal/paths"
)

// ExportOptions configures report export.
type ExportOptions struct {
	GotrVersion string
	// Filter is a glob (filepath.Match) applied to basenames. Empty = match all.
	Filter string
}

// Result reports the outcome of an export or import.
type Result struct {
	// ArchivePath is the on-disk path written (zip for `all`, plain copy
	// for a single file). For imports it is the source path.
	ArchivePath string
	// Copied is the list of files copied or extracted (absolute paths).
	Copied []string
	Files  []bundle.File
}

// ExportSingle copies a single report file from ~/.gotr/reports/ into
// destPath. If destPath is empty, the report is copied under
// ~/.gotr/exports/<basename>.
func ExportSingle(reportsDir, reportName, destPath string) (*Result, error) {
	src := filepath.Join(reportsDir, reportName)
	st, err := os.Stat(src)
	if err != nil {
		return nil, fmt.Errorf("reportbundle: %w", err)
	}
	if st.IsDir() {
		return nil, fmt.Errorf("reportbundle: %q is a directory", reportName)
	}
	if destPath == "" {
		out, err := paths.EnsureExportsDirPath()
		if err != nil {
			return nil, err
		}
		destPath = filepath.Join(out, reportName)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return nil, fmt.Errorf("reportbundle: mkdir %s: %w", destPath, err)
	}
	if err := copyFile(src, destPath); err != nil {
		return nil, err
	}
	return &Result{ArchivePath: destPath, Copied: []string{destPath}}, nil
}

// ExportAll bundles every report file under reportsDir (optionally filtered)
// into a zip archive written to destPath. If destPath is empty, the archive
// is written under ~/.gotr/exports/reports_<YYYYMMDD>.zip.
func ExportAll(reportsDir, destPath string, opts ExportOptions) (*Result, error) {
	selected, err := selectReports(reportsDir, opts.Filter)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, errors.New("reportbundle: no reports matched")
	}

	if destPath == "" {
		out, err := paths.EnsureExportsDirPath()
		if err != nil {
			return nil, err
		}
		destPath = filepath.Join(out, fmt.Sprintf("reports_%s.zip", time.Now().UTC().Format("20060102")))
	}

	entries, files, err := buildReportEntries(reportsDir, selected)
	if err != nil {
		return nil, err
	}

	manifest := &bundle.Manifest{
		SchemaVersion: bundle.SchemaVersion,
		Kind:          bundle.KindReports,
		GotrVersion:   opts.GotrVersion,
		GeneratedAt:   time.Now().UTC(),
		Files:         files,
		ReportCount:   len(files),
	}
	manifestJSON, err := manifest.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("reportbundle: marshal manifest: %w", err)
	}
	entries = append(entries,
		bundle.Entry{ArchivePath: bundle.ManifestName, Content: manifestJSON},
		bundle.Entry{ArchivePath: bundle.ChecksumsName, Content: []byte(bundle.FormatSHA256Sums(files))},
		bundle.Entry{ArchivePath: bundle.ReadmeName, Content: []byte(readmeReports(manifest))},
	)

	if err := bundle.WriteZip(destPath, entries); err != nil {
		return nil, fmt.Errorf("reportbundle: write zip: %w", err)
	}
	return &Result{ArchivePath: destPath, Copied: nil, Files: files}, nil
}

func selectReports(reportsDir, filter string) ([]string, error) {
	ents, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil, fmt.Errorf("reportbundle: read %s: %w", reportsDir, err)
	}
	var selected []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if filter != "" {
			match, err := filepath.Match(filter, e.Name())
			if err != nil {
				return nil, fmt.Errorf("reportbundle: bad filter %q: %w", filter, err)
			}
			if !match {
				continue
			}
		}
		selected = append(selected, e.Name())
	}
	sort.Strings(selected)
	return selected, nil
}

func buildReportEntries(reportsDir string, names []string) ([]bundle.Entry, []bundle.File, error) {
	home, _ := os.UserHomeDir()
	var entries []bundle.Entry
	var files []bundle.File
	for _, name := range names {
		src := filepath.Join(reportsDir, name)
		data, err := os.ReadFile(src) //nolint:gosec // reportsDir is controlled
		if err != nil {
			return nil, nil, fmt.Errorf("reportbundle: read %s: %w", src, err)
		}
		archive := "reports/" + name
		entries = append(entries, bundle.Entry{
			ArchivePath: archive,
			RelHome:     relHome(home, src),
			Content:     data,
		})
		files = append(files, bundle.File{
			Path:    archive,
			RelHome: relHome(home, src),
			SHA256:  bundle.SHA256Bytes(data),
			Size:    int64(len(data)),
		})
	}
	return entries, files, nil
}

// ImportOptions controls how a report is imported.
type ImportOptions struct {
	Overwrite bool
	DryRun    bool
}

// Import imports a single file (pdf/md/json) or a zip bundle of reports
// into ~/.gotr/reports/. Returns the list of destination files.
func Import(reportsDir, srcPath string, opts ImportOptions) (*Result, error) {
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		return nil, fmt.Errorf("reportbundle: mkdir reports: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(srcPath))
	if ext == ".zip" {
		return importZip(reportsDir, srcPath, opts)
	}
	return importSingle(reportsDir, srcPath, opts)
}

func importSingle(reportsDir, srcPath string, opts ImportOptions) (*Result, error) {
	base := filepath.Base(srcPath)
	dst := filepath.Join(reportsDir, base)
	if _, err := os.Stat(dst); err == nil && !opts.Overwrite {
		return nil, fmt.Errorf("reportbundle: %s already exists; use --overwrite", base)
	}
	if opts.DryRun {
		return &Result{ArchivePath: srcPath, Copied: []string{dst}}, nil
	}
	if err := copyFile(srcPath, dst); err != nil {
		return nil, err
	}
	return &Result{ArchivePath: srcPath, Copied: []string{dst}}, nil
}

func importZip(reportsDir, srcPath string, opts ImportOptions) (*Result, error) {
	tmp, err := os.MkdirTemp("", "gotr-reports-import-*")
	if err != nil {
		return nil, fmt.Errorf("reportbundle: tmpdir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	if _, err := bundle.ReadZip(srcPath, tmp); err != nil {
		return nil, fmt.Errorf("reportbundle: read zip: %w", err)
	}
	// Validate manifest (tolerate absence for backward-compat).
	manifest, _ := parseManifest(filepath.Join(tmp, bundle.ManifestName))
	if manifest != nil {
		if err := manifest.ValidateSchema(); err != nil {
			return nil, err
		}
		if manifest.Kind != bundle.KindReports && manifest.Kind != "" {
			return nil, fmt.Errorf("reportbundle: unexpected bundle kind %q", manifest.Kind)
		}
	}

	reportsSrc := filepath.Join(tmp, "reports")
	ents, err := os.ReadDir(reportsSrc)
	if err != nil {
		return nil, fmt.Errorf("reportbundle: archive missing reports/: %w", err)
	}

	var copied []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		dst := filepath.Join(reportsDir, e.Name())
		if _, err := os.Stat(dst); err == nil && !opts.Overwrite {
			return nil, fmt.Errorf("reportbundle: %s already exists; use --overwrite", e.Name())
		}
		if opts.DryRun {
			copied = append(copied, dst)
			continue
		}
		if err := copyFile(filepath.Join(reportsSrc, e.Name()), dst); err != nil {
			return nil, err
		}
		copied = append(copied, dst)
	}
	sort.Strings(copied)
	return &Result{ArchivePath: srcPath, Copied: copied}, nil
}

func parseManifest(path string) (*bundle.Manifest, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is tmp dir
	if err != nil {
		return nil, err
	}
	var m bundle.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // trusted local path
	if err != nil {
		return fmt.Errorf("reportbundle: open %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst) //nolint:gosec // trusted local path
	if err != nil {
		return fmt.Errorf("reportbundle: create %s: %w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return fmt.Errorf("reportbundle: copy %s -> %s: %w", src, dst, err)
	}
	return out.Close()
}

func relHome(home, p string) string {
	if home == "" {
		return p
	}
	if rel, err := filepath.Rel(home, p); err == nil && !strings.HasPrefix(rel, "..") {
		return "~/" + filepath.ToSlash(rel)
	}
	return p
}

func readmeReports(m *bundle.Manifest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "gotr reports bundle\n")
	fmt.Fprintf(&b, "generated_at: %s\n", m.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "gotr_version: %s\n", m.GotrVersion)
	fmt.Fprintf(&b, "schema_version: %d\n", m.SchemaVersion)
	fmt.Fprintf(&b, "report_count: %d\n", m.ReportCount)
	fmt.Fprintf(&b, "\nImport with: gotr import report <file>.zip\n")
	return b.String()
}
