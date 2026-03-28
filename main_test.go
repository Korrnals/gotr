package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMain_Success(t *testing.T) {
	initCalled := false
	syncCalled := false
	execCalled := false
	notifyCalled := false
	stopCalled := false

	notify := func(ctx context.Context, sig ...os.Signal) (context.Context, context.CancelFunc) {
		notifyCalled = true
		return ctx, func() {
			stopCalled = true
		}
	}

	err := runMain(
		func() error {
			initCalled = true
			return nil
		},
		func() error {
			syncCalled = true
			return nil
		},
		func(ctx context.Context) {
			execCalled = true
			assert.NotNil(t, ctx)
		},
		notify,
	)

	assert.NoError(t, err)
	assert.True(t, initCalled)
	assert.True(t, notifyCalled)
	assert.True(t, execCalled)
	assert.True(t, syncCalled)
	assert.True(t, stopCalled)
}

func TestRunMain_InitLoggerError(t *testing.T) {
	syncCalled := false
	execCalled := false
	notifyCalled := false

	err := runMain(
		func() error {
			return errors.New("boom")
		},
		func() error {
			syncCalled = true
			return nil
		},
		func(ctx context.Context) {
			execCalled = true
		},
		func(ctx context.Context, sig ...os.Signal) (context.Context, context.CancelFunc) {
			notifyCalled = true
			return ctx, func() {}
		},
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "init logger")
	assert.False(t, syncCalled)
	assert.False(t, execCalled)
	assert.False(t, notifyCalled)
}

func TestMain_NoPanic(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gotr"}

	assert.NotPanics(t, func() {
		main()
	})
}

func TestMain_PanicPath(t *testing.T) {
	if os.Getenv("GOTR_MAIN_PANIC_CHILD") == "1" {
		tmpDir := t.TempDir()
		homeAsFile := tmpDir + "/not-a-dir"
		require.NoError(t, os.WriteFile(homeAsFile, []byte("x"), 0o644))
		require.NoError(t, os.Setenv("HOME", homeAsFile))
		os.Args = []string{"gotr"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_PanicPath", "-test.v")
	cmd.Env = append(os.Environ(), "GOTR_MAIN_PANIC_CHILD=1")
	err := cmd.Run()
	assert.Error(t, err)
}
