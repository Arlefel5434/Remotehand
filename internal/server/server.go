package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Arlefel5434/Remotehand/internal/input"
	"log"
	"net"
	"net/http"
	"time"

	webui "github.com/Arlefel5434/Remotehand/web"
)

// Config defines the public HTTP endpoint and shutdown behavior.
type Config struct {
	Token           string
	OnListening     func(net.Addr)
	Volume          input.VolumeController
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
	return newHandler(connectionContext, config, &connectionGate{})
}

func newHandler(connectionContext context.Context, config Config, gate *connectionGate) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok", Version: config.Version})
	})
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

	gate := &connectionGate{}
	httpServer := &http.Server{
		Addr:              config.ListenAddress,
		Handler:           newHandler(connectionContext, config, gate),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	if config.OnListening != nil {
		config.OnListening(listener.Addr())
	}
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- httpServer.Serve(listener)
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
	}

	connectionDone := gate.stop()
	closeConnections()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down HTTP server: %w", err)
	}
	if connectionDone != nil {
		select {
		case <-connectionDone:
		case <-shutdownCtx.Done():
			return fmt.Errorf("close controller connection: %w", shutdownCtx.Err())
		}
	}

	err = <-serverErr
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP after shutdown: %w", err)
	}
	return nil
}

func validateConfig(config Config) error {
	if config.Token == "" {
		return errors.New("pairing token required")
	}
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
