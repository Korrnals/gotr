package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRootCmd_Properties проверяет базовые свойства root команды
func TestRootCmd_Properties(t *testing.T) {
	assert.Equal(t, "gotr", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)
	assert.NotEmpty(t, Version)
}

// TestVersion_Properties проверяет версию
func TestVersion_Properties(t *testing.T) {
	assert.NotEmpty(t, Version)
	assert.NotEmpty(t, Commit)
	assert.NotEmpty(t, Date)
}

// TestGetClient_NotNilContext проверяет что GetClient требует контекст
func TestGetClient_NotNilContext(t *testing.T) {
	// GetClient требует контекст с clientом
	// Если контекст пустой, функция вызывает panic
	// Проверяем только что функция существует
	assert.NotNil(t, GetClient)
}

// TestGetClientInterface_NotNilContext проверяет что GetClient требует контекст
func TestGetClientInterface_NotNilContext(t *testing.T) {
	// GetClient требует контекст с clientом
	// Если контекст пустой, функция вызывает panic
	// Проверяем только что функция существует
	assert.NotNil(t, GetClient)
}

// TestRootCmd_NonInteractiveFlagRegistered проверяет наличие флага --non-interactive
func TestRootCmd_NonInteractiveFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("non-interactive")
	assert.NotNil(t, flag, "--non-interactive flag must be registered")
	assert.Equal(t, "false", flag.DefValue)
}

func TestGetClient_PanicWithoutClient(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	assert.Panics(t, func() {
		_ = GetClient(cmd)
	})
}

func TestGetClient_Success(t *testing.T) {
	httpClient := &client.HTTPClient{}
	cmd := &cobra.Command{}
	cmd.SetContext(context.WithValue(context.Background(), httpClientKey, httpClient))

	got := GetClient(cmd)
	assert.Equal(t, httpClient, got)
}

func TestGetClient_PanicOnUnexpectedType(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.WithValue(context.Background(), httpClientKey, struct{}{}))

	assert.Panics(t, func() {
		_ = GetClient(cmd)
	})
}

// TestGetClient_WithMock проверяет что GetClient работает с mock клиентом
func TestGetClient_WithMock(t *testing.T) {
	mock := &client.MockClient{}
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	cmd.SetContext(ctx)

	got := GetClient(cmd)
	assert.Equal(t, mock, got)
}

func TestGetClientInterface_PanicWithoutClient(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	assert.Panics(t, func() {
		_ = GetClient(cmd)
	})
}

func TestGetClientInterface_PanicOnUnexpectedType(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.WithValue(context.Background(), httpClientKey, struct{}{}))

	assert.Panics(t, func() {
		_ = GetClient(cmd)
	})
}

func TestExecute_SuccessPath(t *testing.T) {
	originalRoot := rootCmd
	defer func() { rootCmd = originalRoot }()

	rootCmd = &cobra.Command{Use: "gotr-test"}
	assert.NotPanics(t, func() {
		Execute(context.Background())
	})
}

func TestExecute_ErrorExitCode1(t *testing.T) {
	if os.Getenv("GOTR_EXECUTE_CHILD_MODE") == "error" {
		rootCmd = &cobra.Command{
			Use: "gotr-test",
			RunE: func(cmd *cobra.Command, args []string) error {
				return errors.New("boom")
			},
		}
		Execute(context.Background())
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExecute_ErrorExitCode1", "-test.v")
	cmd.Env = append(os.Environ(), "GOTR_EXECUTE_CHILD_MODE=error")
	err := cmd.Run()

	var exitErr *exec.ExitError
	if assert.Error(t, err) && errors.As(err, &exitErr) {
		assert.Equal(t, 1, exitErr.ExitCode())
	}
}

func TestExecute_CanceledExitCode130(t *testing.T) {
	if os.Getenv("GOTR_EXECUTE_CHILD_MODE") == "canceled" {
		rootCmd = &cobra.Command{
			Use: "gotr-test",
			RunE: func(cmd *cobra.Command, args []string) error {
				return context.Canceled
			},
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		Execute(ctx)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExecute_CanceledExitCode130", "-test.v")
	cmd.Env = append(os.Environ(), "GOTR_EXECUTE_CHILD_MODE=canceled")
	err := cmd.Run()

	var exitErr *exec.ExitError
	if assert.Error(t, err) && errors.As(err, &exitErr) {
		assert.Equal(t, 130, exitErr.ExitCode())
	}
}

func TestInitConfig_WithInvalidYAML(t *testing.T) {
	viper.Reset()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
		viper.Reset()
	})

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	invalidConfig := filepath.Join(tmpDir, "default.yaml")
	require.NoError(t, os.WriteFile(invalidConfig, []byte("base_url: ["), 0o600))

	assert.NotPanics(t, func() {
		initConfig()
	})
}

func TestInitConfig_ConfigNotFound(t *testing.T) {
	viper.Reset()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
		viper.Reset()
	})

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	require.NoError(t, os.Chdir(tmpDir))

	assert.NotPanics(t, func() {
		initConfig()
	})
}

func TestInitConfig_UserHomeDirError(t *testing.T) {
	viper.Reset()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	origUserHomeDir := userHomeDir
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
		userHomeDir = origUserHomeDir
		viper.Reset()
	})

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	userHomeDir = func() (string, error) {
		return "", errors.New("boom")
	}

	assert.NotPanics(t, func() {
		initConfig()
	})
}

func TestRootPersistentPreRunE_NonInteractivePrompterInjected(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("base_url", "https://example.com")
	viper.Set("username", "qa@example.com")
	viper.Set("api_key", "api-key")

	cmd := &cobra.Command{Use: "test-cmd"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("non-interactive", "true"))
	cmd.SetContext(context.Background())

	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.NoError(t, err)

	assert.True(t, interactive.IsNonInteractive(cmd.Context()))
	_, ok := interactive.PrompterFromContext(cmd.Context()).(*interactive.NonInteractivePrompter)
	assert.True(t, ok)
	assert.NotNil(t, cmd.Context().Value(httpClientKey))
}

func TestRootPersistentPreRunE_PasswordHasPriorityOverDefaultAPIKey(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("base_url", "https://example.com")
	viper.Set("username", "qa@example.com")
	viper.Set("password", "password-auth")
	viper.Set("api_key", config.DefaultAPIKey)

	cmd := &cobra.Command{Use: "test-cmd"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.SetContext(context.Background())

	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.NoError(t, err)
	assert.False(t, interactive.IsNonInteractive(cmd.Context()))
	_, ok := interactive.PrompterFromContext(cmd.Context()).(*interactive.TerminalPrompter)
	assert.True(t, ok)
}

func TestRootPersistentPreRunE_FallsBackToAPIKeyWhenPasswordEmpty(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("base_url", "https://example.com")
	viper.Set("username", "qa@example.com")
	viper.Set("password", "")
	viper.Set("api_key", "api-key-fallback")

	cmd := &cobra.Command{Use: "test-cmd"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.SetContext(context.Background())

	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.NoError(t, err)
	assert.NotNil(t, cmd.Context().Value(httpClientKey))
}

func TestRootPersistentPreRunE_RejectsDefaultConfigValues(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("base_url", config.DefaultBaseURL)
	viper.Set("username", config.DefaultUsername)
	viper.Set("api_key", config.DefaultAPIKey)

	cmd := &cobra.Command{Use: "test-cmd"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.SetContext(context.Background())

	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not set or contains default values")
}

func TestRootPersistentPreRunE_ReturnsClientCreationError(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("base_url", "://broken-url")
	viper.Set("username", "qa@example.com")
	viper.Set("api_key", "api-key")

	cmd := &cobra.Command{Use: "test-cmd"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.SetContext(context.Background())

	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create client")
}

func TestRootPersistentPreRunE_InsecureOptionEnabled(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("base_url", "https://example.com")
	viper.Set("username", "qa@example.com")
	viper.Set("api_key", "api-key")
	viper.Set("insecure", true)

	cmd := &cobra.Command{Use: "test-cmd"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.SetContext(context.Background())

	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.NoError(t, err)
	assert.NotNil(t, cmd.Context().Value(httpClientKey))
}

func TestInitConfig_NonNotFoundReadError(t *testing.T) {
	viper.Reset()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
		viper.Reset()
	})

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	require.NoError(t, os.Chdir(tmpDir))

	configDir := filepath.Join(tmpDir, ".gotr", "config")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "default.yaml"), []byte("base_url: ["), 0o600))

	assert.NotPanics(t, func() {
		initConfig()
	})
}

func TestInitConfig_FallbackToCurrentDirectoryWhenHomeUnavailable(t *testing.T) {
	viper.Reset()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	origUserHomeDir := userHomeDir
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
		userHomeDir = origUserHomeDir
		viper.Reset()
	})

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte("base_url: from-current-dir\n"), 0o600))

	userHomeDir = func() (string, error) {
		return "", errors.New("home unavailable")
	}

	initConfig()
	assert.Equal(t, filepath.Join(tmpDir, "default.yaml"), viper.ConfigFileUsed())
}

func TestInitConfig_ReadsConfigFromHomePath(t *testing.T) {
	viper.Reset()
	origUserHomeDir := userHomeDir
	t.Cleanup(func() {
		userHomeDir = origUserHomeDir
		viper.Reset()
	})

	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".gotr", "config")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	configPath := filepath.Join(configDir, "default.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("base_url: home-config\n"), 0o600))

	userHomeDir = func() (string, error) {
		return tmpHome, nil
	}

	initConfig()
	assert.Equal(t, configPath, viper.ConfigFileUsed())
}

func TestInitConfig_ReadsConfigFromCurrentDirWhenHomeHasNoConfig(t *testing.T) {
	viper.Reset()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	origUserHomeDir := userHomeDir
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
		userHomeDir = origUserHomeDir
		viper.Reset()
	})

	tmpHome := t.TempDir()
	tmpWD := t.TempDir()
	wdConfig := filepath.Join(tmpWD, "default.yaml")
	require.NoError(t, os.WriteFile(wdConfig, []byte("base_url: wd-config\n"), 0o600))
	require.NoError(t, os.Chdir(tmpWD))

	userHomeDir = func() (string, error) {
		return tmpHome, nil
	}

	initConfig()
	assert.Equal(t, wdConfig, viper.ConfigFileUsed())
}
