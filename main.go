// main.go — util entrypoint
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Korrnals/gotr/cmd"
	"github.com/Korrnals/gotr/internal/log"
)

func main() {
	if err := runMain(log.InitDefault, log.Sync, cmd.Execute, signal.NotifyContext); err != nil {
		panic(err)
	}
}

func runMain(
	initLogger func() error,
	syncLogger func() error,
	execute func(context.Context),
	notifyContext func(context.Context, ...os.Signal) (context.Context, context.CancelFunc),
) error {
	if err := initLogger(); err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() { _ = syncLogger() }()

	ctx, stop := notifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	execute(ctx)

	return nil
}
