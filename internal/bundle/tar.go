package bundle

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// WriteTarGz writes entries into destPath as a gzip-compressed tar archive.
// The parent directory of destPath is created as needed.
func WriteTarGz(destPath string, entries []Entry) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("tar.gz mkdir parent: %w", err)
	}
	f, err := os.Create(destPath) //nolint:gosec // callers pass validated paths under ~/.gotr/exports
	if err != nil {
		return fmt.Errorf("tar.gz create %s: %w", destPath, err)
	}
	defer func() { _ = f.Close() }()

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	for _, e := range entries {
		if err := writeTarEntry(tw, e); err != nil {
			_ = tw.Close()
			_ = gz.Close()
			return err
		}
	}
	if err := tw.Close(); err != nil {
		_ = gz.Close()
		return fmt.Errorf("tar close: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("gzip close: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("tar.gz close file: %w", err)
	}
	return nil
}

func writeTarEntry(tw *tar.Writer, e Entry) error {
	if err := validateArchivePath(e.ArchivePath); err != nil {
		return err
	}
	mode := e.Mode
	if mode == 0 {
		mode = defaultMode
	}
	mt := e.ModTime
	if mt.IsZero() {
		mt = defaultTime()
	}

	var size int64
	switch {
	case e.Content != nil:
		size = int64(len(e.Content))
	case e.SourcePath != "":
		st, err := os.Stat(e.SourcePath)
		if err != nil {
			return fmt.Errorf("tar stat %s: %w", e.SourcePath, err)
		}
		size = st.Size()
	default:
		return fmt.Errorf("tar entry %q: neither Content nor SourcePath set", e.ArchivePath)
	}

	hdr := &tar.Header{
		Name:    e.ArchivePath,
		Mode:    int64(mode.Perm()),
		Size:    size,
		ModTime: mt,
		Format:  tar.FormatPAX,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("tar header %s: %w", e.ArchivePath, err)
	}

	if e.Content != nil {
		if _, err := tw.Write(e.Content); err != nil {
			return fmt.Errorf("tar write %s: %w", e.ArchivePath, err)
		}
		return nil
	}

	src, err := os.Open(e.SourcePath) //nolint:gosec // callers pass trusted local paths
	if err != nil {
		return fmt.Errorf("tar open %s: %w", e.SourcePath, err)
	}
	defer func() { _ = src.Close() }()
	if _, err := io.Copy(tw, src); err != nil {
		return fmt.Errorf("tar copy %s: %w", e.SourcePath, err)
	}
	return nil
}

// ReadTarGz extracts srcPath into destDir. It rejects entries containing
// path traversal (".." segments or absolute paths) and refuses symlinks.
// Returns the list of extracted archive paths (forward-slash) in archive
// order.
func ReadTarGz(srcPath, destDir string) ([]string, error) {
	f, err := os.Open(srcPath) //nolint:gosec // import path is validated by caller
	if err != nil {
		return nil, fmt.Errorf("tar.gz open %s: %w", srcPath, err)
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer func() { _ = gz.Close() }()

	return extractTar(tar.NewReader(gz), destDir)
}

func extractTar(tr *tar.Reader, destDir string) ([]string, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, fmt.Errorf("tar extract mkdir %s: %w", destDir, err)
	}
	var names []string
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar next: %w", err)
		}
		if err := validateArchivePath(hdr.Name); err != nil {
			return nil, err
		}
		switch hdr.Typeflag {
		case tar.TypeReg, tar.TypeRegA: //nolint:staticcheck // TypeRegA deprecated but still appears in older archives
			if err := extractTarRegular(tr, hdr, destDir); err != nil {
				return nil, err
			}
			names = append(names, hdr.Name)
		case tar.TypeDir:
			// Directories are recreated as needed when files are written.
			continue
		default:
			return nil, fmt.Errorf("tar extract: unsupported entry type %c for %q", hdr.Typeflag, hdr.Name)
		}
	}
	return names, nil
}

func extractTarRegular(tr *tar.Reader, hdr *tar.Header, destDir string) error {
	target := filepath.Join(destDir, filepath.FromSlash(hdr.Name))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("tar extract mkdir %s: %w", filepath.Dir(target), err)
	}
	mode := hdr.FileInfo().Mode().Perm()
	if mode == 0 {
		mode = defaultMode
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode) //nolint:gosec // path validated above
	if err != nil {
		return fmt.Errorf("tar extract create %s: %w", target, err)
	}
	if _, err := io.Copy(out, tr); err != nil { //nolint:gosec // size bounded by tar header; gotr bundles are user-generated
		_ = out.Close()
		return fmt.Errorf("tar extract write %s: %w", target, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("tar extract close %s: %w", target, err)
	}
	return nil
}

// validateArchivePath rejects unsafe paths to prevent zip-slip style
// attacks. Entries must be forward-slash, non-empty, and must not escape
// the destination directory.
func validateArchivePath(p string) error {
	if p == "" {
		return errors.New("archive entry has empty path")
	}
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, `\`) {
		return fmt.Errorf("archive entry %q has absolute path", p)
	}
	clean := filepath.ToSlash(filepath.Clean(p))
	if clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return fmt.Errorf("archive entry %q escapes archive root", p)
	}
	return nil
}
