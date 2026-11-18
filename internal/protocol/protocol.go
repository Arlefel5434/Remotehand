package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	Version        = 1
	MaxMessageSize = 8 * 1024
	MaxIDLength    = 64
	MaxTextBytes   = 2048
)

type Type string

const (
	TypeMove   Type = "move"
	TypeScroll Type = "scroll"
	TypeClick  Type = "click"
	TypeText   Type = "text"
	TypeKey    Type = "key"
	TypeVolume Type = "volume"
	TypeMute   Type = "mute"
)

type Command struct {
	Type    Type
	ID      string
	Payload any
}

type MovePayload struct{ DX, DY int }
type ScrollPayload struct{ Steps int }
type ClickPayload struct{ Button string }
type TextPayload struct{ Value string }
type KeyPayload struct{ Key string }
type VolumePayload struct {
	Delta int
	Level *int
}
type MutePayload struct{ Muted bool }

type Violation struct {
	Code        string
	Message     string
	Fatal       bool
	CloseStatus int
}

func (v *Violation) Error() string { return v.Message }

type ErrorEnvelope struct {
	Version int          `json:"version"`
	Type    string       `json:"type"`
	ID      string       `json:"id,omitempty"`
	Payload ErrorPayload `json:"payload"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewErrorEnvelope(id, code, message string) ErrorEnvelope {
	return ErrorEnvelope{
		Version: Version,
		Type:    "error",
		ID:      id,
		Payload: ErrorPayload{Code: code, Message: message},
	}
}

type rawEnvelope struct {
	Version *int            `json:"version"`
	Type    *string         `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

func Parse(data []byte) (Command, *Violation) {
	if !utf8.Valid(data) {
		return Command{}, fatal("invalid_envelope", "message must be valid UTF-8", 1007)
	}

	var envelope rawEnvelope
	if err := decodeStrict(data, &envelope); err != nil {
		return Command{}, fatal("invalid_envelope", "message must be one valid JSON object", 1007)
	}
	if envelope.Version == nil || envelope.Type == nil || len(envelope.Payload) == 0 {
		return Command{}, fatal("invalid_envelope", "version, type, and payload are required", 1007)
	}
	if *envelope.Version != Version {
		return Command{}, fatal("unsupported_version", fmt.Sprintf("protocol version %d is unsupported", *envelope.Version), 1002)
	}
	if utf8.RuneCountInString(envelope.ID) > MaxIDLength || strings.IndexFunc(envelope.ID, unicode.IsControl) >= 0 {
		return Command{}, fatal("invalid_envelope", "id must be at most 64 characters without control characters", 1007)
	}

	command := Command{Type: Type(*envelope.Type), ID: envelope.ID}
	var violation *Violation
	switch command.Type {
	case TypeScroll:
		var v struct {
			Steps int `json:"steps"`
		}
		if err := decodeStrict(envelope.Payload, &v); err != nil || v.Steps == 0 || v.Steps < -8 || v.Steps > 8 {
			violation = invalidPayload("steps must be a nonzero integer between -8 and 8")
		} else {
			command.Payload = ScrollPayload{Steps: v.Steps}
		}
	case TypeMove:
		command.Payload, violation = parseMove(envelope.Payload)
	case TypeClick:
		command.Payload, violation = parseClick(envelope.Payload)
	case TypeText:
		command.Payload, violation = parseText(envelope.Payload)
	case TypeKey:
		command.Payload, violation = parseKey(envelope.Payload)
	case TypeVolume:
		command.Payload, violation = parseVolume(envelope.Payload)
	case TypeMute:
		command.Payload, violation = parseMute(envelope.Payload)
	default:
		return command, nonfatal("unknown_command", fmt.Sprintf("command type %q is unsupported", command.Type))
	}
	if violation != nil {
		return command, violation
	}
	return command, nil
}

func parseMove(data []byte) (any, *Violation) {
	var value struct {
		DX *int `json:"dx"`
		DY *int `json:"dy"`
	}
	if err := decodeStrict(data, &value); err != nil || value.DX == nil || value.DY == nil {
		return nil, invalidPayload("move requires integer dx and dy")
	}
	if *value.DX < -200 || *value.DX > 200 || *value.DY < -200 || *value.DY > 200 {
		return nil, invalidPayload("dx and dy must be between -200 and 200")
	}
	return MovePayload{DX: *value.DX, DY: *value.DY}, nil
}

func parseClick(data []byte) (any, *Violation) {
	var value struct {
		Button *string `json:"button"`
	}
	if err := decodeStrict(data, &value); err != nil || value.Button == nil {
		return nil, invalidPayload("click requires button")
	}
	if *value.Button != "left" && *value.Button != "right" {
		return nil, invalidPayload("button must be left or right")
	}
	return ClickPayload{Button: *value.Button}, nil
}

func parseText(data []byte) (any, *Violation) {
	var value struct {
		Value *string `json:"value"`
	}
	if err := decodeStrict(data, &value); err != nil || value.Value == nil {
		return nil, invalidPayload("text requires value")
	}
	if !utf8.ValidString(*value.Value) || len(*value.Value) > MaxTextBytes {
		return nil, invalidPayload("text must be valid UTF-8 and at most 2048 bytes")
	}
	return TextPayload{Value: *value.Value}, nil
}

func parseKey(data []byte) (any, *Violation) {
	var value struct {
		Key *string `json:"key"`
	}
	if err := decodeStrict(data, &value); err != nil || value.Key == nil {
		return nil, invalidPayload("key requires key")
	}
	switch *value.Key {
	case "enter", "backspace", "escape":
		return KeyPayload{Key: *value.Key}, nil
	default:
		return nil, invalidPayload("key must be enter, backspace, or escape")
	}
}

func parseVolume(data []byte) (any, *Violation) {
	var value struct {
		Delta *int `json:"delta"`
		Level *int `json:"level"`
	}
	if err := decodeStrict(data, &value); err != nil || (value.Delta == nil) == (value.Level == nil) {
		return nil, invalidPayload("volume requires either delta or level")
	}
	if value.Level != nil {
		if *value.Level < 0 || *value.Level > 100 {
			return nil, invalidPayload("level must be between 0 and 100")
		}
		return VolumePayload{Level: value.Level}, nil
	}
	if *value.Delta == 0 || *value.Delta < -10 || *value.Delta > 10 {
		return nil, invalidPayload("delta must be between -10 and 10 and cannot be zero")
	}
	return VolumePayload{Delta: *value.Delta}, nil
}

func parseMute(data []byte) (any, *Violation) {
	var value struct {
		Muted *bool `json:"muted"`
	}
	if err := decodeStrict(data, &value); err != nil || value.Muted == nil {
		return nil, invalidPayload("mute requires boolean muted")
	}
	return MutePayload{Muted: *value.Muted}, nil
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON value")
	}
	return nil
}

func invalidPayload(message string) *Violation {
	return nonfatal("invalid_payload", message)
}

func nonfatal(code, message string) *Violation {
	return &Violation{Code: code, Message: message}
}

func fatal(code, message string, closeStatus int) *Violation {
	return &Violation{Code: code, Message: message, Fatal: true, CloseStatus: closeStatus}
}
