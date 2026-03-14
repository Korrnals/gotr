// main.go — util entrypoint
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Korrnals/gotr/cmd"
	"github.com/Korrnals/gotr/internal/log"
)

func main() {
	// Инициализация логгера
	if err := log.InitDefault(); err != nil {
		panic(err)
	}
	defer log.Sync()

	// Подключаем OS-сигналы — Ctrl+C теперь отменяет контекст и все in-flight HTTP запросы
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd.Execute(ctx)
}
