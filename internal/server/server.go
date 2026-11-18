package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	webui "github.com/Arlefel5434/Remotehand/web"
)

// Config defines the public HTTP endpoint and shutdown behavior.
type Config struct {
	ListenAddress   string
	Version         string
	ShutdownTimeout time.Duration
	PingInterval    time.Duration
	PongTimeout     time.Duration
	Sink            CommandSink
	Logger          *log.Logger
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func NewHandler(connectionContext context.Context, config Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok", Version: config.Version})
	})
	gate := &connectionGate{}
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(connectionContext, config, gate, w, r)
	})
	mux.Handle("GET /", http.FileServerFS(webui.Files))
	return mux
}

// Run serves HTTP until the context is cancelled, then shuts down gracefully.
func Run(ctx context.Context, config Config) error {
	if err := validateConfig(config); err != nil {
		return err
	}
	connectionContext, closeConnections := context.WithCancel(context.Background())
	defer closeConnections()

	httpServer := &http.Server{
		Addr:              config.ListenAddress,
		Handler:           NewHandler(connectionContext, config),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
	}

	closeConnections()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down HTTP server: %w", err)
	}

	err := <-serverErr
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP after shutdown: %w", err)
	}
	return nil
}

func validateConfig(config Config) error {
	if config.ListenAddress == "" {
		return errors.New("listen address is required")
	}
	if config.ShutdownTimeout <= 0 || config.PingInterval <= 0 || config.PongTimeout <= 0 {
		return errors.New("shutdown, ping, and pong timeouts must be positive")
	}
	if config.Sink == nil || config.Logger == nil {
		return errors.New("command sink and logger are required")
	}
	return nil
}
