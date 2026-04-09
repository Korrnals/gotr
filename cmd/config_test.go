package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestConfigInitCmd_CreatesDefaultConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	viper.Reset()

	err := configInitCmd.RunE(configInitCmd, nil)
	assert.NoError(t, err)

	cfgPath := filepath.Join(home, ".gotr", "config", "default.yaml")
	_, statErr := os.Stat(cfgPath)
	assert.NoError(t, statErr)
}

func TestConfigPathCmd(t *testing.T) {
	t.Run("without used config", func(t *testing.T) {
		viper.Reset()
		err := configPathCmd.RunE(configPathCmd, nil)
		assert.NoError(t, err)
	})

	t.Run("with used config", func(t *testing.T) {
		viper.Reset()
		dir := t.TempDir()
		cfg := filepath.Join(dir, "cfg.yaml")
		assert.NoError(t, os.WriteFile(cfg, []byte("base_url: test\n"), 0o644))

		viper.SetConfigFile(cfg)
		assert.NoError(t, viper.ReadInConfig())

		err := configPathCmd.RunE(configPathCmd, nil)
		assert.NoError(t, err)
	})
}

func TestConfigViewCmd(t *testing.T) {
	t.Run("without used config", func(t *testing.T) {
		viper.Reset()
		err := configViewCmd.RunE(configViewCmd, nil)
		assert.NoError(t, err)
	})

	t.Run("with used config", func(t *testing.T) {
		viper.Reset()
		dir := t.TempDir()
		cfg := filepath.Join(dir, "cfg.yaml")
		assert.NoError(t, os.WriteFile(cfg, []byte("base_url: test\n"), 0o644))

		viper.SetConfigFile(cfg)
		assert.NoError(t, viper.ReadInConfig())

		err := configViewCmd.RunE(configViewCmd, nil)
		assert.NoError(t, err)
	})
}

func TestConfigEditCmd_NoConfigUsed(t *testing.T) {
	viper.Reset()
	err := configEditCmd.RunE(configEditCmd, nil)
	assert.NoError(t, err)
}

// TestConfigViewCmd_ReadFileError covers the os.ReadFile error path
// when the config file is recorded by viper but no longer exists.
func TestConfigViewCmd_ReadFileError(t *testing.T) {
	viper.Reset()
	dir := t.TempDir()
	cfg := filepath.Join(dir, "cfg.yaml")
	assert.NoError(t, os.WriteFile(cfg, []byte("base_url: test\n"), 0o644))

	viper.SetConfigFile(cfg)
	assert.NoError(t, viper.ReadInConfig())

	// Delete the file after viper has read it
	os.Remove(cfg)

	err := configViewCmd.RunE(configViewCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config")
}

// TestConfigEditCmd_WithConfig_EditorSuccess covers the OpenEditor success path.
func TestConfigEditCmd_WithConfig_EditorSuccess(t *testing.T) {
	t.Setenv("EDITOR", "/bin/true")
	viper.Reset()
	dir := t.TempDir()
	cfg := filepath.Join(dir, "cfg.yaml")
	assert.NoError(t, os.WriteFile(cfg, []byte("base_url: test\n"), 0o644))

	viper.SetConfigFile(cfg)
	assert.NoError(t, viper.ReadInConfig())

	err := configEditCmd.RunE(configEditCmd, nil)
	assert.NoError(t, err)
}

// TestConfigEditCmd_WithConfig_EditorError covers the OpenEditor error path.
func TestConfigEditCmd_WithConfig_EditorError(t *testing.T) {
	t.Setenv("EDITOR", "/nonexistent-binary-12345")
	viper.Reset()
	dir := t.TempDir()
	cfg := filepath.Join(dir, "cfg.yaml")
	assert.NoError(t, os.WriteFile(cfg, []byte("base_url: test\n"), 0o644))

	viper.SetConfigFile(cfg)
	assert.NoError(t, viper.ReadInConfig())

	err := configEditCmd.RunE(configEditCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "editor")
}

func TestRedactSensitiveConfig(t *testing.T) {
	input := strings.Join([]string{
		"base_url: \"https://example.testrail.io\"",
		"api_key: \"super-secret\"",
		"password: plain-password",
		"token: abc123",
		"authorization: Bearer token-value",
		"username: \"qa@example.com\"",
	}, "\n")

	out := redactSensitiveConfig(input)

	assert.NotContains(t, out, "super-secret")
	assert.NotContains(t, out, "plain-password")
	assert.NotContains(t, out, "abc123")
	assert.NotContains(t, out, "token-value")
	assert.Contains(t, out, "api_key: \"***\"")
	assert.Contains(t, out, "password: \"***\"")
	assert.Contains(t, out, "token: \"***\"")
	assert.Contains(t, out, "authorization: \"***\"")
	assert.Contains(t, out, "base_url: \"https://example.testrail.io\"")
	assert.Contains(t, out, "username: \"qa@example.com\"")
}
