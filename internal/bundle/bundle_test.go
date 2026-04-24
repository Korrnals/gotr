package bundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSHA256Bytes_Deterministic(t *testing.T) {
	h1 := SHA256Bytes([]byte("hello world"))
	h2 := SHA256Bytes([]byte("hello world"))
	if h1 != h2 {
		t.Fatalf("expected identical hashes, got %s vs %s", h1, h2)
	}
	if len(h1) != 64 {
		t.Fatalf("expected 64-char hex hash, got len=%d", len(h1))
	}
}

func TestFormatAndParseSHA256Sums_Roundtrip(t *testing.T) {
	files := []File{
		{Path: "b.txt", SHA256: strings.Repeat("b", 64)},
		{Path: "a.txt", SHA256: strings.Repeat("a", 64)},
	}
	text := FormatSHA256Sums(files)
	// Sorted by path.
	if !strings.HasPrefix(text, strings.Repeat("a", 64)+"  a.txt\n") {
		t.Fatalf("unexpected output order: %q", text)
	}
	parsed, err := ParseSHA256Sums(text)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed["a.txt"] != strings.Repeat("a", 64) || parsed["b.txt"] != strings.Repeat("b", 64) {
		t.Fatalf("roundtrip mismatch: %+v", parsed)
	}
}

func TestManifest_ValidateSchema(t *testing.T) {
	m := &Manifest{SchemaVersion: SchemaVersion}
	if err := m.ValidateSchema(); err != nil {
		t.Fatalf("current schema must validate: %v", err)
	}
	m.SchemaVersion = 0
	if err := m.ValidateSchema(); err == nil {
		t.Fatal("missing schema_version must fail")
	}
	m.SchemaVersion = SchemaVersion + 99
	if err := m.ValidateSchema(); err == nil {
		t.Fatal("future schema_version must fail")
	}
}

func TestWriteAndReadTarGz_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	if err := os.WriteFile(src, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "out.tar.gz")
	entries := []Entry{
		{ArchivePath: ManifestName, Content: []byte(`{"k":"v"}`)},
		{ArchivePath: "nested/dir/file.txt", SourcePath: src},
	}
	if err := WriteTarGz(archive, entries); err != nil {
		t.Fatalf("write tar.gz: %v", err)
	}
	st, err := os.Stat(archive)
	if err != nil || st.Size() == 0 {
		t.Fatalf("archive not created: %v size=%d", err, st.Size())
	}

	ext := filepath.Join(dir, "ext")
	names, err := ReadTarGz(archive, ext)
	if err != nil {
		t.Fatalf("read tar.gz: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("want 2 entries, got %d: %v", len(names), names)
	}
	got, err := os.ReadFile(filepath.Join(ext, "nested/dir/file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "payload" {
		t.Fatalf("payload mismatch: %q", got)
	}
}

func TestWriteAndReadZip_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	if err := os.WriteFile(src, []byte("zip-payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "out.zip")
	entries := []Entry{
		{ArchivePath: ManifestName, Content: []byte(`{}`)},
		{ArchivePath: "reports/a.md", SourcePath: src},
	}
	if err := WriteZip(archive, entries); err != nil {
		t.Fatalf("write zip: %v", err)
	}
	ext := filepath.Join(dir, "ext")
	if _, err := ReadZip(archive, ext); err != nil {
		t.Fatalf("read zip: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(ext, "reports/a.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "zip-payload" {
		t.Fatalf("payload mismatch: %q", got)
	}
}

func TestValidateArchivePath_RejectsTraversal(t *testing.T) {
	cases := []string{
		"",
		"/abs/path",
		"../escape",
		"nested/../../etc/passwd",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			if err := validateArchivePath(p); err == nil {
				t.Fatalf("path %q should be rejected", p)
			}
		})
	}
	ok := []string{"manifest.json", "snaps/abc/meta.json", "nested/dir/file.txt"}
	for _, p := range ok {
		if err := validateArchivePath(p); err != nil {
			t.Fatalf("path %q should be accepted: %v", p, err)
		}
	}
}
