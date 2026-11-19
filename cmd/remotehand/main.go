package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
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
	sensitivity, err := pointerSensitivity(os.Getenv("REMOTEHAND_POINTER_SENSITIVITY"))
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config := server.Config{
		Token:           rand.Text(),
		Volume:          adapter,
		ListenAddress:   envOrDefault("REMOTEHAND_LISTEN_ADDR", ":8080"),
		Version:         version,
		ShutdownTimeout: 5 * time.Second,
		PingInterval:    20 * time.Second,
		PongTimeout:     5 * time.Second,
		Logger:          log.Default(),
	}
	config.Sink = server.CommandDispatcher{
		Pointer:     adapter,
		Keyboard:    adapter,
		Fallback:    server.LoggingSink{Logger: config.Logger},
		Sensitivity: sensitivity,
	}
	if runtime.GOOS == "windows" {
		config.OnListening = func(addr net.Addr) {
			if info, err := os.Stdout.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
				go watchPhoneLinks(ctx, addr.(*net.TCPAddr), config.Token)
				fmt.Println("Введите disconnect и нажмите Enter, чтобы завершить сеанс и остановить агент.")
			}
			go func() {
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					if strings.EqualFold(strings.TrimSpace(scanner.Text()), "disconnect") {
						stop()
						return
					}
				}
			}()
		}
	} else if path := os.Getenv("REMOTEHAND_PAIR_FILE"); path != "" {
		if err := os.WriteFile(path, []byte(config.Token), 0600); err != nil {
			return err
		}
		defer os.Remove(path)
	}
	log.Printf("starting Remotehand version=%s address=%s input_adapter=%s sensitivity=%.2f", config.Version, config.ListenAddress, adapter.Name(), sensitivity)

	return server.Run(ctx, config)
}

func pointerSensitivity(value string) (float64, error) {
	if value == "" {
		return 1, nil
	}
	sensitivity, err := strconv.ParseFloat(value, 64)
	if err != nil || sensitivity < 0.1 || sensitivity > 3 {
		return 0, fmt.Errorf("REMOTEHAND_POINTER_SENSITIVITY must be a number between 0.1 and 3.0")
	}
	return sensitivity, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
