package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Arlefel5434/Remotehand/internal/protocol"
)

func TestHealth(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	NewHandler(context.Background(), testConfig("test-version", discardSink{})).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	var body healthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := healthResponse{Status: "ok", Version: "test-version"}
	if body != want {
		t.Fatalf("body = %#v, want %#v", body, want)
	}
}

func TestHealthRejectsOtherMethods(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodPost, "/health", nil)
	response := httptest.NewRecorder()

	NewHandler(context.Background(), testConfig("dev", discardSink{})).ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

func TestUnknownPath(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	response := httptest.NewRecorder()

	NewHandler(context.Background(), testConfig("dev", discardSink{})).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

type discardSink struct{}

func (discardSink) Handle(context.Context, protocol.Command) error { return nil }

func testConfig(version string, sink CommandSink) Config {
	return Config{
		Version:      version,
		PingInterval: time.Hour,
		PongTimeout:  time.Second,
		Sink:         sink,
		Logger:       log.New(io.Discard, "", 0),
	}
}
