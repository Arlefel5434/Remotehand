package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arlefel5434/Remotehand/internal/input"
	"github.com/Arlefel5434/Remotehand/internal/server"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		log.Printf("remotehand stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	adapter, err := input.New(os.Getenv("REMOTEHAND_INPUT_MODE"))
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config := server.Config{
		ListenAddress:   envOrDefault("REMOTEHAND_LISTEN_ADDR", ":8080"),
		Version:         version,
		ShutdownTimeout: 5 * time.Second,
	}
	log.Printf("starting Remotehand version=%s address=%s input_adapter=%s", config.Version, config.ListenAddress, adapter.Name())

	return server.Run(ctx, config)
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
