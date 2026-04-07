package selftest

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Korrnals/gotr/internal/paths"
	"github.com/stretchr/testify/assert"
)

func TestCountMatches(t *testing.T) {
	assert.Equal(t, 0, countMatches("", "a"))
	assert.Equal(t, 2, countMatches("abc--- PASS: x\n--- PASS: y", "--- PASS:"))
	assert.Equal(t, 1, countMatches("FAIL", "FAIL"))
}

func TestParsePackageResults(t *testing.T) {
	out := "ok   github.com/acme/p1 0.01s\nFAIL github.com/acme/p2 0.02s\nok   github.com/acme/p3 0.03s\n"
	parsed := parsePackageResults(out)
	assert.Contains(t, parsed, "✓ p1")
	assert.Contains(t, parsed, "✗ p2")
	assert.Contains(t, parsed, "✓ p3")
}

func TestParsePackageResults_TruncateAndSkipMalformed(t *testing.T) {
	out := "ok\n" +
		"ok   github.com/acme/p1 0.01s\n" +
		"ok   github.com/acme/p2 0.01s\n" +
		"ok   github.com/acme/p3 0.01s\n" +
		"ok   github.com/acme/p4 0.01s\n" +
		"ok   github.com/acme/p5 0.01s\n" +
		"FAIL github.com/acme/p6 0.01s\n"

	parsed := parsePackageResults(out)
	assert.Contains(t, parsed, "+1 more")
	assert.Contains(t, parsed, "✓ p1")
	assert.NotContains(t, parsed, "p6")
}

func TestExtractOverallCoverage(t *testing.T) {
	out := "ok github.com/acme/x coverage: 67.8% of statements\n"
	assert.Equal(t, "67.8%", extractOverallCoverage(out))
	assert.Equal(t, "unknown", extractOverallCoverage("no coverage here"))
}

func TestGetProjectRoot(t *testing.T) {
	oldWD, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldWD) }()

	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	assert.NoError(t, os.MkdirAll(nested, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644))
	assert.NoError(t, os.Chdir(nested))

	got := getProjectRoot()
	assert.Equal(t, root, got)
}

func TestGetProjectRoot_CurrentDirHasGoMod(t *testing.T) {
	oldWD, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldWD) }()

	root := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644))
	assert.NoError(t, os.Chdir(root))

	got := getProjectRoot()
	assert.Equal(t, root, got)
}

func TestBinaryInfoAndGoEnvCheckers(t *testing.T) {
	b := BinaryInfoChecker{Version: "1.2.3", Commit: "abcdef123456", BuildTime: "now"}
	res := b.Check()
	assert.Equal(t, ResultPass, res.Result)
	assert.Contains(t, res.Message, "1.2.3")
	assert.Contains(t, res.Details, "abcdef12")

	g := GoEnvChecker{}
	gres := g.Check()
	assert.Equal(t, ResultPass, gres.Result)
	assert.Contains(t, gres.Message, runtime.Version())
}

// ---------------------------------------------------------------------------
// Name / Category getters for all checker types
// ---------------------------------------------------------------------------

func TestCheckerNameCategory(t *testing.T) {
	checkers := []struct {
		name     string
		checker  Checker
		wantName string
		wantCat  string
	}{
		{
			name:     "ConfigChecker",
			checker:  ConfigChecker{},
			wantName: "Configuration File",
			wantCat:  "Configuration",
		},
		{
			name:     "BaseDirChecker",
			checker:  BaseDirChecker{},
			wantName: "Base Directory Structure",
			wantCat:  "Configuration",
		},
		{
			name:     "BinaryInfoChecker",
			checker:  BinaryInfoChecker{Version: "v", Commit: "c", BuildTime: "t"},
			wantName: "Binary Information",
			wantCat:  "System",
		},
		{
			name:     "GoEnvChecker",
			checker:  GoEnvChecker{},
			wantName: "Go Environment",
			wantCat:  "System",
		},
		{
			name:     "AllTestsChecker",
			checker:  AllTestsChecker{},
			wantName: "All Unit Tests",
			wantCat:  "Tests",
		},
		{
			name:     "CoverageChecker",
			checker:  CoverageChecker{},
			wantName: "Code Coverage",
			wantCat:  "Coverage",
		},
	}

	for _, tt := range checkers {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantName, tt.checker.Name())
			assert.Equal(t, tt.wantCat, tt.checker.Category())
		})
	}
}

func TestConfigAndBaseDirCheckers(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgChecker := ConfigChecker{}
	resMissing := cfgChecker.Check()
	assert.Equal(t, ResultFail, resMissing.Result)
	assert.True(t, resMissing.CanFix)

	cfgPath, err := paths.ConfigFile()
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(filepath.Dir(cfgPath), 0o755))
	assert.NoError(t, os.WriteFile(cfgPath, []byte("base_url: test\n"), 0o644))

	resFound := cfgChecker.Check()
	assert.Equal(t, ResultPass, resFound.Result)
	assert.Contains(t, resFound.Message, "found")

	baseChecker := BaseDirChecker{}
	baseRes := baseChecker.Check()
	assert.Equal(t, ResultPass, baseRes.Result)
	assert.NotEmpty(t, baseRes.Message)
}

func TestConfigChecker_ConfigPathError(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")
	origXdg := os.Getenv("XDG_CONFIG_HOME")

	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	res := ConfigChecker{}.Check()

	if origHome == "" && origUserProfile == "" && origHomeDrive == "" && origHomePath == "" && origXdg == "" {
		if res.Result != ResultFail || res.Error == nil {
			t.Skip("os.UserHomeDir resolved home on this platform/user setup")
		}
		return
	}

	if res.Error == nil {
		t.Skip("os.UserHomeDir resolved home from system user database")
	}
	assert.Equal(t, ResultFail, res.Result)
	assert.Contains(t, res.Message, "Cannot determine config path")
}

func TestBaseDirChecker_AllDirectoriesExist(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	pathsToCreate := []func() (string, error){
		paths.ConfigDirPath,
		paths.LogsDirPath,
		paths.SelftestDirPath,
		paths.CacheDirPath,
		paths.ExportsDirPath,
		paths.TempDirPath,
	}

	for _, fn := range pathsToCreate {
		dir, err := fn()
		assert.NoError(t, err)
		assert.NoError(t, os.MkdirAll(dir, 0o755))
	}

	res := BaseDirChecker{}.Check()
	assert.Equal(t, ResultPass, res.Result)
	assert.Equal(t, "All directories exist", res.Message)
}

func TestBaseDirChecker_FileInsteadOfDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath, err := paths.ConfigDirPath()
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o755))
	assert.NoError(t, os.WriteFile(configPath, []byte("x"), 0o644))

	res := BaseDirChecker{}.Check()
	assert.Equal(t, ResultWarn, res.Result)
	assert.Equal(t, "Some directories could not be created", res.Message)
	assert.Contains(t, res.Details, "config")
}

func TestBaseDirChecker_PathResolutionError(t *testing.T) {
	t.Setenv("HOME", "")

	res := BaseDirChecker{}.Check()
	if res.Error == nil {
		t.Skip("os.UserHomeDir resolved home without HOME; skipping error-branch assertion")
	}

	assert.Equal(t, ResultFail, res.Result)
	assert.Contains(t, res.Message, "Cannot determine")
}

func TestCoverageCheckerHelpers(t *testing.T) {
	c := CoverageChecker{}
	assert.Equal(t, "Code Coverage", c.Name())
	assert.Equal(t, "Coverage", c.Category())
}

func TestAllTestsChecker_Check(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "printf '%s\n' '--- PASS: TestA' '--- FAIL: TestB' 'ok   github.com/acme/p1 0.01s' 'FAIL github.com/acme/p2 0.02s'; exit 1")
	}

	res := AllTestsChecker{}.Check()
	assert.Equal(t, ResultFail, res.Result)
	assert.Contains(t, res.Message, "1 passed, 1 failed, 0 skipped")
	assert.Error(t, res.Error)
}

func TestCoverageChecker_Check(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "printf '%s\n' 'ok github.com/acme/x coverage: 55.5% of statements'")
	}

	res := CoverageChecker{}.Check()
	assert.Equal(t, ResultPass, res.Result)
	assert.Equal(t, "55.5%", res.Message)
}

func TestCoverageChecker_CheckCommandError(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "echo boom >&2; exit 1")
	}

	res := CoverageChecker{}.Check()
	assert.Equal(t, ResultWarn, res.Result)
	assert.Contains(t, res.Message, "Cannot calculate coverage")
	assert.Error(t, res.Error)
}

func TestAllTestsChecker_CheckSuccess(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "printf '%s\n' '--- PASS: TestA' '--- SKIP: TestB' 'ok   github.com/acme/p1 0.01s'")
	}

	res := AllTestsChecker{}.Check()
	assert.Equal(t, ResultPass, res.Result)
	assert.Contains(t, res.Message, "1 passed, 0 failed, 1 skipped")
	assert.NoError(t, res.Error)
}

func TestAllTestsChecker_CheckCommandErrorWithoutFailedCount(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "echo command failed >&2; exit 1")
	}

	res := AllTestsChecker{}.Check()
	assert.Equal(t, ResultFail, res.Result)
	assert.Error(t, res.Error)
	assert.NotContains(t, res.Error.Error(), "test(s) failed")
}

func TestCoverageChecker_CheckWarnOnLowCoverage(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "printf '%s\n' 'ok github.com/acme/x coverage: 40.0% of statements'")
	}

	res := CoverageChecker{}.Check()
	assert.Equal(t, ResultWarn, res.Result)
	assert.Contains(t, res.Details, "Current: 40.0%")
}

func TestCoverageChecker_CheckUnknownCoverage(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "printf '%s\n' 'ok github.com/acme/x no coverage output'")
	}

	res := CoverageChecker{}.Check()
	assert.Equal(t, ResultPass, res.Result)
	assert.Equal(t, "unknown", res.Message)
	assert.Equal(t, "Target: 80%", res.Details)
}

func TestGetProjectRoot_NoGoModFallback(t *testing.T) {
	oldWD, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldWD) }()

	root := t.TempDir()
	nested := filepath.Join(root, "x", "y")
	assert.NoError(t, os.MkdirAll(nested, 0o755))
	assert.NoError(t, os.Chdir(nested))

	got := getProjectRoot()
	assert.Equal(t, ".", got)
}
