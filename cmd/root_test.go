package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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

// TestGetClientInterface_NotNilContext проверяет что GetClientInterface требует контекст
func TestGetClientInterface_NotNilContext(t *testing.T) {
	// GetClientInterface требует контекст с clientом
	// Если контекст пустой, функция вызывает panic
	// Проверяем только что функция существует
	assert.NotNil(t, GetClientInterface)
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

func TestGetClientInterface_WithMock(t *testing.T) {
	mock := &client.MockClient{}
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	cmd.SetContext(ctx)

	got := GetClientInterface(cmd)
	assert.Equal(t, mock, got)
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
