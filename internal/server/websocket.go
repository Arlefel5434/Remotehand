package server

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Arlefel5434/Remotehand/internal/protocol"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type CommandSink interface {
	Handle(context.Context, protocol.Command) error
}

type LoggingSink struct{ Logger *log.Logger }

func (s LoggingSink) Handle(_ context.Context, command protocol.Command) error {
	s.Logger.Printf("command type=%s", command.Type)
	return nil
}

type connectionGate struct {
	mu     sync.Mutex
	active bool
}

func (g *connectionGate) acquire() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.active {
		return false
	}
	g.active = true
	return true
}

func (g *connectionGate) release() {
	g.mu.Lock()
	g.active = false
	g.mu.Unlock()
}

func handleWebSocket(parent context.Context, config Config, gate *connectionGate, w http.ResponseWriter, r *http.Request) {
	if config.Token == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Sec-WebSocket-Protocol")), []byte("remotehand."+config.Token)) != 1 {
		http.Error(w, "pairing required", http.StatusUnauthorized)
		return
	}
	if !gate.acquire() {
		http.Error(w, "a controller is already connected", http.StatusConflict)
		return
	}
	defer gate.release()

	connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{CompressionMode: websocket.CompressionDisabled, Subprotocols: []string{"remotehand." + config.Token}})
	if err != nil {
		config.Logger.Print("websocket accept failed")
		return
	}
	defer connection.CloseNow()
	connection.SetReadLimit(protocol.MaxMessageSize)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pingErr := make(chan error, 1)
	go heartbeat(ctx, connection, config.PingInterval, config.PongTimeout, pingErr)
	readErr := make(chan error, 1)
	go func() { readErr <- readCommands(ctx, connection, config) }()

	config.Logger.Printf("connection opened remote=%s", r.RemoteAddr)
	select {
	case <-parent.Done():
		_ = connection.Close(websocket.StatusGoingAway, "agent is stopping")
		cancel()
		<-readErr
	case err = <-readErr:
		cancel()
		if err != nil && websocket.CloseStatus(err) != websocket.StatusNormalClosure {
			config.Logger.Print("connection closed")
		}
	case heartbeatError := <-pingErr:
		cancel()
		if heartbeatError != nil {
			config.Logger.Print("connection heartbeat failed")
		}
		_ = connection.CloseNow()
		<-readErr
	}
}

func heartbeat(ctx context.Context, connection *websocket.Conn, interval, timeout time.Duration, result chan<- error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			result <- nil
			return
		case <-ticker.C:
			pingCtx, pingCancel := context.WithTimeout(ctx, timeout)
			err := connection.Ping(pingCtx)
			pingCancel()
			if err != nil {
				result <- err
				return
			}
		}
	}
}

func readCommands(ctx context.Context, connection *websocket.Conn, config Config) error {
	publishVolume := func(id string, delta int, level *int, muted *bool) error {
		state, err := config.Volume.UpdateVolume(delta, level, muted)
		if err != nil {
			return writeError(ctx, connection, id, "audio_unavailable", "Аудиоустройство недоступно")
		}
		writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return wsjson.Write(writeCtx, connection, map[string]any{"version": protocol.Version, "type": "volume_state", "payload": state})
	}
	if config.Volume != nil {
		if err := publishVolume("", 0, nil, nil); err != nil {
			return err
		}
	}
	limit := newCommandBudget(time.Now())
	for {
		messageType, data, err := connection.Read(ctx)
		if err != nil {
			return err
		}
		if messageType != websocket.MessageText {
			_ = connection.Close(websocket.StatusInvalidFramePayloadData, "text messages are required")
			return errors.New("received non-text WebSocket message")
		}
		if !limit.allow(time.Now()) {
			_ = connection.Close(websocket.StatusPolicyViolation, "command rate exceeded")
			return errors.New("command rate exceeded")
		}

		command, violation := protocol.Parse(data)
		if violation != nil {
			if violation.Fatal {
				_ = connection.Close(websocket.StatusCode(violation.CloseStatus), violation.Code)
				return violation
			}
			if err := writeError(ctx, connection, command.ID, violation.Code, violation.Message); err != nil {
				return fmt.Errorf("write protocol error: %w", err)
			}
			continue
		}

		if config.Volume != nil {
			switch p := command.Payload.(type) {
			case protocol.VolumePayload:
				if err := publishVolume(command.ID, p.Delta, p.Level, nil); err != nil {
					return err
				}
				continue
			case protocol.MutePayload:
				if err := publishVolume(command.ID, 0, nil, &p.Muted); err != nil {
					return err
				}
				continue
			}
		}
		if err := config.Sink.Handle(ctx, command); err != nil {
			if writeErr := writeError(ctx, connection, command.ID, "command_failed", "command could not be processed"); writeErr != nil {
				return fmt.Errorf("write command error: %w", writeErr)
			}
		}
	}
}

func writeError(ctx context.Context, connection *websocket.Conn, id, code, message string) error {
	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return wsjson.Write(writeCtx, connection, protocol.NewErrorEnvelope(id, code, message))
}

type commandBudget struct {
	credit float64
	last   time.Time
}

const commandsPerSecond = 120

func newCommandBudget(now time.Time) commandBudget { return commandBudget{commandsPerSecond, now} }
func (b *commandBudget) allow(now time.Time) bool {
	b.credit = min(float64(commandsPerSecond), b.credit+now.Sub(b.last).Seconds()*commandsPerSecond)
	b.last = now
	if b.credit < 1 {
		return false
	}
	b.credit--
	return true
}
