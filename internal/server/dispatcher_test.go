package server

import (
	"context"
	"errors"
	"testing"

	"github.com/Arlefel5434/Remotehand/internal/input"
	"github.com/Arlefel5434/Remotehand/internal/protocol"
)

type fakePointer struct {
	dx, dy  int
	button  input.Button
	moveErr error
}

func (p *fakePointer) Move(dx, dy int) error {
	p.dx, p.dy = dx, dy
	return p.moveErr
}

func (p *fakePointer) Click(button input.Button) error {
	p.button = button
	return nil
}

type countingSink struct{ calls int }

func (s *countingSink) Handle(context.Context, protocol.Command) error {
	s.calls++
	return nil
}

func TestCommandDispatcherMovesWithSensitivityAndClamp(t *testing.T) {
	t.Parallel()
	pointer := &fakePointer{}
	dispatcher := CommandDispatcher{Pointer: pointer, Fallback: &countingSink{}, Sensitivity: 1.5}

	err := dispatcher.Handle(context.Background(), protocol.Command{
		Type:    protocol.TypeMove,
		Payload: protocol.MovePayload{DX: 150, DY: -10},
	})
	if err != nil {
		t.Fatalf("Handle(): %v", err)
	}
	if pointer.dx != 200 || pointer.dy != -15 {
		t.Fatalf("move = (%d,%d), want (200,-15)", pointer.dx, pointer.dy)
	}
}

func TestCommandDispatcherClicks(t *testing.T) {
	t.Parallel()
	pointer := &fakePointer{}
	dispatcher := CommandDispatcher{Pointer: pointer, Fallback: &countingSink{}, Sensitivity: 1}

	err := dispatcher.Handle(context.Background(), protocol.Command{
		Type:    protocol.TypeClick,
		Payload: protocol.ClickPayload{Button: "right"},
	})
	if err != nil {
		t.Fatalf("Handle(): %v", err)
	}
	if pointer.button != input.ButtonRight {
		t.Fatalf("button = %q, want right", pointer.button)
	}
}

func TestCommandDispatcherForwardsOtherCommands(t *testing.T) {
	t.Parallel()
	fallback := &countingSink{}
	dispatcher := CommandDispatcher{Pointer: &fakePointer{}, Fallback: fallback, Sensitivity: 1}

	if err := dispatcher.Handle(context.Background(), protocol.Command{Type: protocol.TypeText}); err != nil {
		t.Fatalf("Handle(): %v", err)
	}
	if fallback.calls != 1 {
		t.Fatalf("fallback calls = %d, want 1", fallback.calls)
	}
}

func TestCommandDispatcherReturnsPointerError(t *testing.T) {
	t.Parallel()
	want := errors.New("send input failed")
	dispatcher := CommandDispatcher{
		Pointer:     &fakePointer{moveErr: want},
		Fallback:    &countingSink{},
		Sensitivity: 1,
	}
	err := dispatcher.Handle(context.Background(), protocol.Command{
		Type:    protocol.TypeMove,
		Payload: protocol.MovePayload{DX: 1, DY: 1},
	})
	if !errors.Is(err, want) {
		t.Fatalf("Handle() error = %v, want %v", err, want)
	}
}
