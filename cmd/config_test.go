package cmd

import (
	"os"
	"path/filepath"
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
