// main.go — util entrypoint
package main

import (
	"github.com/Korrnals/gotr/cmd"
	"github.com/Korrnals/gotr/internal/log"
)

func main() {
	// Инициализация логгера
	if err := log.InitDefault(); err != nil {
		panic(err)
	}
	defer log.Sync()

	cmd.Execute()
}
