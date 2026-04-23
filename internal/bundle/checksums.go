package bundle

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// SHA256File returns the hex-encoded SHA-256 digest of the file at path.
func SHA256File(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // path comes from trusted callers enumerating local files
	if err != nil {
		return "", fmt.Errorf("sha256 open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("sha256 read %s: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// SHA256Bytes returns the hex-encoded SHA-256 digest of b.
func SHA256Bytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// FormatSHA256Sums produces GNU-compatible SHA256SUMS text for files,
// stable-sorted by archive path for reproducibility.
func FormatSHA256Sums(files []File) string {
	cp := make([]File, len(files))
	copy(cp, files)
	sort.Slice(cp, func(i, j int) bool { return cp[i].Path < cp[j].Path })

	var b strings.Builder
	for _, f := range cp {
		// Two spaces between hash and filename is the canonical GNU
		// coreutils sha256sum format (binary mode uses a single space+asterisk
		// instead; we stick with portable text mode).
		fmt.Fprintf(&b, "%s  %s\n", f.SHA256, f.Path)
	}
	return b.String()
}

// ParseSHA256Sums parses GNU-format SHA256SUMS text into a map keyed by
// archive path.
func ParseSHA256Sums(text string) (map[string]string, error) {
	out := map[string]string{}
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		// Format: "<hex>  <path>" (two spaces) or "<hex> *<path>" (binary).
		var hash, path string
		switch {
		case strings.Contains(line, "  "):
			parts := strings.SplitN(line, "  ", 2)
			hash, path = parts[0], parts[1]
		case strings.Contains(line, " *"):
			parts := strings.SplitN(line, " *", 2)
			hash, path = parts[0], parts[1]
		default:
			return nil, fmt.Errorf("sha256sums: malformed line %q", line)
		}
		if len(hash) != 64 {
			return nil, fmt.Errorf("sha256sums: bad hash length in %q", line)
		}
		out[path] = hash
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("sha256sums: scan: %w", err)
	}
	return out, nil
}
