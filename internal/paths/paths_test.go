package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBaseAndDerivedPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	base, err := BaseDir()
	if err != nil {
		t.Fatalf("BaseDir error: %v", err)
	}
	wantBase := filepath.Join(home, DirName)
	if base != wantBase {
		t.Fatalf("BaseDir = %q, want %q", base, wantBase)
	}

	cfg, err := ConfigDirPath()
	if err != nil {
		t.Fatalf("ConfigDirPath error: %v", err)
	}
	if cfg != filepath.Join(wantBase, ConfigDir) {
		t.Fatalf("ConfigDirPath = %q", cfg)
	}

	logs, err := LogsDirPath()
	if err != nil {
		t.Fatalf("LogsDirPath error: %v", err)
	}
	if logs != filepath.Join(wantBase, LogsDir) {
		t.Fatalf("LogsDirPath = %q", logs)
	}

	selftest, err := SelftestDirPath()
	if err != nil {
		t.Fatalf("SelftestDirPath error: %v", err)
	}
	if selftest != filepath.Join(wantBase, SelftestDir) {
		t.Fatalf("SelftestDirPath = %q", selftest)
	}

	cache, err := CacheDirPath()
	if err != nil {
		t.Fatalf("CacheDirPath error: %v", err)
	}
	if cache != filepath.Join(wantBase, CacheDir) {
		t.Fatalf("CacheDirPath = %q", cache)
	}

	exports, err := ExportsDirPath()
	if err != nil {
		t.Fatalf("ExportsDirPath error: %v", err)
	}
	if exports != filepath.Join(wantBase, ExportsDir) {
		t.Fatalf("ExportsDirPath = %q", exports)
	}

	tmp, err := TempDirPath()
	if err != nil {
		t.Fatalf("TempDirPath error: %v", err)
	}
	if tmp != filepath.Join(wantBase, TempDir) {
		t.Fatalf("TempDirPath = %q", tmp)
	}

	cfgFile, err := ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile error: %v", err)
	}
	if cfgFile != filepath.Join(wantBase, ConfigDir, "default.yaml") {
		t.Fatalf("ConfigFile = %q", cfgFile)
	}
}

func TestEnsureLogsDirPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := EnsureLogsDirPath()
	if err != nil {
		t.Fatalf("EnsureLogsDirPath error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("logs dir not created: %v", err)
	}
}

func TestEnsureAllDirsAndEnsureDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := EnsureAllDirs(); err != nil {
		t.Fatalf("EnsureAllDirs error: %v", err)
	}

	check := []func() (string, error){
		ConfigDirPath,
		LogsDirPath,
		SelftestDirPath,
		CacheDirPath,
		ExportsDirPath,
		TempDirPath,
	}
	for _, fn := range check {
		p, err := fn()
		if err != nil {
			t.Fatalf("path fn error: %v", err)
		}
		if st, err := os.Stat(p); err != nil || !st.IsDir() {
			t.Fatalf("expected dir %q to exist", p)
		}
	}

	custom := filepath.Join(home, DirName, "custom")
	err := EnsureDir(func() (string, error) { return custom, nil })
	if err != nil {
		t.Fatalf("EnsureDir error: %v", err)
	}
	if st, err := os.Stat(custom); err != nil || !st.IsDir() {
		t.Fatalf("expected custom dir %q to exist", custom)
	}
}
