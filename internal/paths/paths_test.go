package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBaseDirAndDerivedPaths_HomeUnset(t *testing.T) {
	t.Setenv("HOME", "")

	if _, err := BaseDir(); err == nil {
		t.Skip("os.UserHomeDir resolved home without HOME; skipping error-branch assertions")
	}

	checks := []struct {
		name string
		fn   func() (string, error)
	}{
		{name: "ConfigDirPath", fn: ConfigDirPath},
		{name: "LogsDirPath", fn: LogsDirPath},
		{name: "SelftestDirPath", fn: SelftestDirPath},
		{name: "CacheDirPath", fn: CacheDirPath},
		{name: "ExportsDirPath", fn: ExportsDirPath},
		{name: "TempDirPath", fn: TempDirPath},
		{name: "ConfigFile", fn: ConfigFile},
		{name: "EnsureLogsDirPath", fn: EnsureLogsDirPath},
	}

	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.fn(); err == nil {
				t.Fatalf("expected error for %s when home is unresolved", tc.name)
			}
		})
	}
}

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

func TestEnsureLogsDirPath_AlreadyExists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	first, err := EnsureLogsDirPath()
	if err != nil {
		t.Fatalf("first EnsureLogsDirPath error: %v", err)
	}
	second, err := EnsureLogsDirPath()
	if err != nil {
		t.Fatalf("second EnsureLogsDirPath error: %v", err)
	}
	if first != second {
		t.Fatalf("paths differ: first=%q second=%q", first, second)
	}
}

func TestEnsureLogsDirPath_MkdirError(t *testing.T) {
	parent := t.TempDir()
	homeFile := filepath.Join(parent, "not-a-dir")
	if err := os.WriteFile(homeFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("write home file: %v", err)
	}
	t.Setenv("HOME", homeFile)

	_, err := EnsureLogsDirPath()
	if err == nil {
		t.Fatal("expected error when home path is a file")
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

func TestEnsureAllDirs_MkdirError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	baseFile := filepath.Join(home, DirName)
	if err := os.WriteFile(baseFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("write base file: %v", err)
	}

	if err := EnsureAllDirs(); err == nil {
		t.Fatal("expected EnsureAllDirs to fail when base path is a file")
	}
}

func TestEnsureAllDirs_PathResolutionError(t *testing.T) {
	t.Setenv("HOME", "")

	err := EnsureAllDirs()
	if err == nil {
		t.Skip("os.UserHomeDir resolved home without HOME; skipping error-branch assertion")
	}
}

func TestEnsureDir_ErrorPaths(t *testing.T) {
	errExpected := os.ErrInvalid
	err := EnsureDir(func() (string, error) { return "", errExpected })
	if err == nil {
		t.Fatal("expected EnsureDir to return dirFunc error")
	}

	home := t.TempDir()
	filePath := filepath.Join(home, "not-a-dir")
	if err := os.WriteFile(filePath, []byte("x"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err = EnsureDir(func() (string, error) {
		return filepath.Join(filePath, "child"), nil
	})
	if err == nil {
		t.Fatal("expected EnsureDir to fail when parent is a file")
	}
}
