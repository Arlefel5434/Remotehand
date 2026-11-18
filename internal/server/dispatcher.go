package server

import (
	"context"
	"fmt"
	"math"

	"github.com/Arlefel5434/Remotehand/internal/input"
	"github.com/Arlefel5434/Remotehand/internal/protocol"
)

type CommandDispatcher struct {
	Pointer     input.PointerController
	Fallback    CommandSink
	Sensitivity float64
}

func (d CommandDispatcher) Handle(ctx context.Context, command protocol.Command) error {
	if d.Pointer == nil || d.Fallback == nil {
		return fmt.Errorf("dispatcher dependencies are required")
	}
	if d.Sensitivity < 0.1 || d.Sensitivity > 3 {
		return fmt.Errorf("pointer sensitivity must be between 0.1 and 3.0")
	}

	switch command.Type {
	case protocol.TypeMove:
		payload, ok := command.Payload.(protocol.MovePayload)
		if !ok {
			return fmt.Errorf("move command has unexpected payload %T", command.Payload)
		}
		return d.Pointer.Move(scaleDelta(payload.DX, d.Sensitivity), scaleDelta(payload.DY, d.Sensitivity))
	case protocol.TypeClick:
		payload, ok := command.Payload.(protocol.ClickPayload)
		if !ok {
			return fmt.Errorf("click command has unexpected payload %T", command.Payload)
		}
		return d.Pointer.Click(input.Button(payload.Button))
	default:
		return d.Fallback.Handle(ctx, command)
	}
}

func scaleDelta(value int, sensitivity float64) int {
	scaled := int(math.Round(float64(value) * sensitivity))
	if scaled < -200 {
		return -200
	}
	if scaled > 200 {
		return 200
	}
	return scaled
}
