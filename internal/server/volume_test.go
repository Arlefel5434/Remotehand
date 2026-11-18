package server

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Arlefel5434/Remotehand/internal/input"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestControllerVolume(t *testing.T) {
	config := testConfig("dev", discardSink{})
	adapter, err := input.New("mock")
	if err != nil {
		t.Fatal(err)
	}
	config.Volume = adapter
	srv := httptest.NewServer(NewHandler(context.Background(), config))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, wsURL(srv.URL), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	check := func(level int, muted bool) {
		t.Helper()
		var msg struct {
			Type    string
			Payload input.VolumeState
		}
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			t.Fatal(err)
		}
		if msg.Type != "volume_state" || msg.Payload.Level != level || msg.Payload.Muted != muted {
			t.Fatalf("unexpected state: %+v", msg)
		}
	}
	check(50, false)
	for n := 1; n <= 6; n++ {
		if err := conn.Write(ctx, websocket.MessageText, []byte(`{"version":1,"type":"volume","payload":{"delta":10}}`)); err != nil {
			t.Fatal(err)
		}
		check(min(100, 50+n*10), false)
	}
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"version":1,"type":"mute","payload":{"muted":true}}`)); err != nil {
		t.Fatal(err)
	}
	check(100, true)
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"version":1,"type":"volume","payload":{"level":37}}`)); err != nil {
		t.Fatal(err)
	}
	check(37, true)
}
