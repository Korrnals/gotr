package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
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
	original := executeMain
	defer func() { executeMain = original }()

	executeMain = func() error { return nil }

	assert.NotPanics(t, func() {
		main()
	})
}

func TestMain_PanicsOnExecuteError(t *testing.T) {
	if os.Getenv("GOTR_MAIN_EXIT_CHILD") == "1" {
		executeMain = func() error { return errors.New("boom") }
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_PanicsOnExecuteError", "-test.v")
	cmd.Env = append(os.Environ(), "GOTR_MAIN_EXIT_CHILD=1")
	out, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(out), "fatal: boom")
}

func TestMain_PanicPath(t *testing.T) {
	if os.Getenv("GOTR_MAIN_PANIC_CHILD") == "1" {
		executeMain = func() error { return errors.New("boom") }
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_PanicPath", "-test.v")
	cmd.Env = append(os.Environ(), "GOTR_MAIN_PANIC_CHILD=1")
	out, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(out), "fatal: boom")
}

func TestExecuteMain_DefaultClosure_HelpPath(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"gotr", "--help"}
	err := executeMain()
	assert.NoError(t, err)
}
