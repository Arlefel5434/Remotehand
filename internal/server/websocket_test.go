package server

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Arlefel5434/Remotehand/internal/protocol"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type recordingSink struct{ commands chan protocol.Command }

func (s recordingSink) Handle(_ context.Context, command protocol.Command) error {
	s.commands <- command
	return nil
}

func TestWebSocketReportsNonFatalErrorAndKeepsConnection(t *testing.T) {
	t.Parallel()

	sink := recordingSink{commands: make(chan protocol.Command, 1)}
	testServer := httptest.NewServer(NewHandler(context.Background(), testConfig("dev", sink)))
	defer testServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, wsURL(testServer.URL), &websocket.DialOptions{Subprotocols: []string{"remotehand.test-token"}})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer connection.CloseNow()

	unknown := map[string]any{"version": 1, "type": "dance", "id": "bad-1", "payload": map[string]any{}}
	if err := wsjson.Write(ctx, connection, unknown); err != nil {
		t.Fatalf("write unknown command: %v", err)
	}
	var response protocol.ErrorEnvelope
	if err := wsjson.Read(ctx, connection, &response); err != nil {
		t.Fatalf("read error response: %v", err)
	}
	if response.ID != "bad-1" || response.Payload.Code != "unknown_command" {
		t.Fatalf("response = %#v", response)
	}

	move := map[string]any{"version": 1, "type": "move", "payload": map[string]any{"dx": 2, "dy": -3}}
	if err := wsjson.Write(ctx, connection, move); err != nil {
		t.Fatalf("write move: %v", err)
	}
	select {
	case command := <-sink.commands:
		if command.Type != protocol.TypeMove {
			t.Fatalf("command type = %q", command.Type)
		}
	case <-ctx.Done():
		t.Fatal("move command was not dispatched")
	}
}

func TestWebSocketRejectsSecondConnection(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(NewHandler(context.Background(), testConfig("dev", discardSink{})))
	defer testServer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	first, _, err := websocket.Dial(ctx, wsURL(testServer.URL), &websocket.DialOptions{Subprotocols: []string{"remotehand.test-token"}})
	if err != nil {
		t.Fatalf("dial first connection: %v", err)
	}
	defer first.CloseNow()

	second, response, err := websocket.Dial(ctx, wsURL(testServer.URL), &websocket.DialOptions{Subprotocols: []string{"remotehand.test-token"}})
	if second != nil {
		second.CloseNow()
	}
	if err == nil {
		t.Fatal("second connection unexpectedly succeeded")
	}
	if response == nil || response.StatusCode != http.StatusConflict {
		t.Fatalf("second response = %#v, want 409", response)
	}
}

func TestWebSocketRejectsForeignOrigin(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(NewHandler(context.Background(), testConfig("dev", discardSink{})))
	defer testServer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	header := http.Header{"Origin": []string{"http://evil.example"}}
	connection, response, err := websocket.Dial(ctx, wsURL(testServer.URL), &websocket.DialOptions{HTTPHeader: header, Subprotocols: []string{"remotehand.test-token"}})
	if connection != nil {
		connection.CloseNow()
	}
	if err == nil {
		t.Fatal("foreign origin unexpectedly succeeded")
	}
	if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("response = %#v, want 403", response)
	}
}

func TestWebSocketClosesOnMalformedEnvelope(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(NewHandler(context.Background(), testConfig("dev", discardSink{})))
	defer testServer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, wsURL(testServer.URL), &websocket.DialOptions{Subprotocols: []string{"remotehand.test-token"}})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer connection.CloseNow()

	if err := connection.Write(ctx, websocket.MessageText, []byte(`{`)); err != nil {
		t.Fatalf("write malformed message: %v", err)
	}
	_, _, err = connection.Read(ctx)
	if websocket.CloseStatus(err) != websocket.StatusInvalidFramePayloadData {
		t.Fatalf("close status = %d, error = %v", websocket.CloseStatus(err), err)
	}
}

func TestConnectionStaysOpenAcrossHeartbeats(t *testing.T) {
	t.Parallel()

	sink := recordingSink{commands: make(chan protocol.Command, 1)}
	config := testConfig("dev", sink)
	config.PingInterval = 20 * time.Millisecond
	config.PongTimeout = 200 * time.Millisecond
	testServer := httptest.NewServer(NewHandler(context.Background(), config))
	defer testServer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, wsURL(testServer.URL), &websocket.DialOptions{Subprotocols: []string{"remotehand.test-token"}})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer connection.CloseNow()

	readDone := make(chan error, 1)
	go func() {
		_, _, readErr := connection.Read(ctx)
		readDone <- readErr
	}()
	time.Sleep(80 * time.Millisecond)

	move := map[string]any{"version": 1, "type": "move", "payload": map[string]any{"dx": 1, "dy": 1}}
	if err := wsjson.Write(ctx, connection, move); err != nil {
		t.Fatalf("write after heartbeats: %v", err)
	}
	select {
	case <-sink.commands:
	case err := <-readDone:
		t.Fatalf("connection closed during heartbeat: %v", err)
	case <-ctx.Done():
		t.Fatal("command was not received after heartbeats")
	}
}

func TestWebSocketClosesOversizedMessage(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(NewHandler(context.Background(), testConfig("dev", discardSink{})))
	defer testServer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, wsURL(testServer.URL), &websocket.DialOptions{Subprotocols: []string{"remotehand.test-token"}})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer connection.CloseNow()

	oversized := bytes.Repeat([]byte("x"), protocol.MaxMessageSize+1)
	if err := connection.Write(ctx, websocket.MessageText, oversized); err != nil {
		t.Fatalf("write oversized message: %v", err)
	}
	_, _, err = connection.Read(ctx)
	if websocket.CloseStatus(err) != websocket.StatusMessageTooBig {
		t.Fatalf("close status = %d, error = %v", websocket.CloseStatus(err), err)
	}
}

func TestAgentShutdownClosesConnection(t *testing.T) {
	t.Parallel()

	connectionContext, stopConnections := context.WithCancel(context.Background())
	testServer := httptest.NewServer(NewHandler(connectionContext, testConfig("dev", discardSink{})))
	defer testServer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, wsURL(testServer.URL), &websocket.DialOptions{Subprotocols: []string{"remotehand.test-token"}})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer connection.CloseNow()

	stopConnections()
	_, _, err = connection.Read(ctx)
	if websocket.CloseStatus(err) != websocket.StatusGoingAway {
		t.Fatalf("close status = %d, error = %v", websocket.CloseStatus(err), err)
	}
}

func TestLoggingSinkDoesNotLogTextPayload(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	sink := LoggingSink{Logger: log.New(&output, "", 0)}
	command := protocol.Command{
		Type:    protocol.TypeText,
		ID:      "text-1",
		Payload: protocol.TextPayload{Value: "do-not-log-this"},
	}
	if err := sink.Handle(context.Background(), command); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if strings.Contains(output.String(), "do-not-log-this") {
		t.Fatalf("log contains text payload: %q", output.String())
	}
}

func wsURL(httpURL string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http") + "/ws"
}

func TestPairingRequiredBeforeUpgrade(t *testing.T) {
	srv := httptest.NewServer(NewHandler(context.Background(), testConfig("dev", discardSink{})))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, offered := range [][]string{nil, {"remotehand.previous-run"}} {
		conn, response, err := websocket.Dial(ctx, wsURL(srv.URL), &websocket.DialOptions{Subprotocols: offered})
		if conn != nil {
			conn.CloseNow()
		}
		if err == nil || response == nil || response.StatusCode != http.StatusUnauthorized {
			t.Fatal("expected HTTP 401 before upgrade")
		}
	}
	conn, _, err := websocket.Dial(ctx, wsURL(srv.URL), &websocket.DialOptions{Subprotocols: []string{"remotehand.test-token"}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	if conn.Subprotocol() != "remotehand.test-token" {
		t.Fatal("browser subprotocol negotiation failed")
	}
}

func TestCommandBudget(t *testing.T) {
	now := time.Now()
	b := newCommandBudget(now)
	for range commandsPerSecond {
		if !b.allow(now) {
			t.Fatal("early rejection")
		}
	}
	if b.allow(now) {
		t.Fatal("exceeded burst")
	}
	if !b.allow(now.Add(time.Second)) {
		t.Fatal("budget did not refill")
	}
}
