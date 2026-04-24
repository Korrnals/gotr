package bundle

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteZip writes entries to destPath as a standard .zip archive.
func WriteZip(destPath string, entries []Entry) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("zip mkdir parent: %w", err)
	}
	f, err := os.Create(destPath) //nolint:gosec // trusted export directory
	if err != nil {
		return fmt.Errorf("zip create %s: %w", destPath, err)
	}
	defer func() { _ = f.Close() }()

	zw := zip.NewWriter(f)
	for _, e := range entries {
		if err := writeZipEntry(zw, e); err != nil {
			_ = zw.Close()
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return fmt.Errorf("zip close: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("zip close file: %w", err)
	}
	return nil
}

func writeZipEntry(zw *zip.Writer, e Entry) error {
	if err := validateArchivePath(e.ArchivePath); err != nil {
		return err
	}
	hdr := &zip.FileHeader{
		Name:   e.ArchivePath,
		Method: zip.Deflate,
	}
	mode := e.Mode
	if mode == 0 {
		mode = defaultMode
	}
	hdr.SetMode(mode)
	mt := e.ModTime
	if mt.IsZero() {
		mt = defaultTime()
	}
	hdr.Modified = mt

	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return fmt.Errorf("zip header %s: %w", e.ArchivePath, err)
	}

	if e.Content != nil {
		if _, err := w.Write(e.Content); err != nil {
			return fmt.Errorf("zip write %s: %w", e.ArchivePath, err)
		}
		return nil
	}
	if e.SourcePath == "" {
		return fmt.Errorf("zip entry %q: neither Content nor SourcePath set", e.ArchivePath)
	}
	src, err := os.Open(e.SourcePath) //nolint:gosec // trusted local path
	if err != nil {
		return fmt.Errorf("zip open %s: %w", e.SourcePath, err)
	}
	defer func() { _ = src.Close() }()
	if _, err := io.Copy(w, src); err != nil {
		return fmt.Errorf("zip copy %s: %w", e.SourcePath, err)
	}
	return nil
}

// ReadZip extracts srcPath into destDir. Rejects unsafe paths.
func ReadZip(srcPath, destDir string) ([]string, error) {
	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return nil, fmt.Errorf("zip open %s: %w", srcPath, err)
	}
	defer func() { _ = r.Close() }()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, fmt.Errorf("zip extract mkdir %s: %w", destDir, err)
	}

	var names []string
	for _, zf := range r.File {
		if err := validateArchivePath(zf.Name); err != nil {
			return nil, err
		}
		if zf.FileInfo().IsDir() {
			continue
		}
		if err := extractZipFile(zf, destDir); err != nil {
			return nil, err
		}
		names = append(names, zf.Name)
	}
	return names, nil
}

func extractZipFile(zf *zip.File, destDir string) error {
	target := filepath.Join(destDir, filepath.FromSlash(zf.Name))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("zip extract mkdir %s: %w", filepath.Dir(target), err)
	}
	mode := zf.Mode().Perm()
	if mode == 0 {
		mode = defaultMode
	}
	src, err := zf.Open()
	if err != nil {
		return fmt.Errorf("zip open entry %s: %w", zf.Name, err)
	}
	defer func() { _ = src.Close() }()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode) //nolint:gosec // path validated
	if err != nil {
		return fmt.Errorf("zip extract create %s: %w", target, err)
	}
	if _, err := io.Copy(out, src); err != nil { //nolint:gosec // bundles are user-generated
		_ = out.Close()
		return fmt.Errorf("zip extract write %s: %w", target, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("zip extract close %s: %w", target, err)
	}
	return nil
}
