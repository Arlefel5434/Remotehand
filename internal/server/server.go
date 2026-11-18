package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Config defines the public HTTP endpoint and shutdown behavior.
type Config struct {
	ListenAddress   string
	Version         string
	ShutdownTimeout time.Duration
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// NewHandler returns the minimal HTTP surface for the current MVP stage.
func NewHandler(version string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok", Version: version})
	})
	return mux
}

// Run serves HTTP until the context is cancelled, then shuts down gracefully.
func Run(ctx context.Context, config Config) error {
	if config.ListenAddress == "" {
		return errors.New("listen address is required")
	}
	if config.ShutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be positive")
	}

	httpServer := &http.Server{
		Addr:              config.ListenAddress,
		Handler:           NewHandler(config.Version),
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
