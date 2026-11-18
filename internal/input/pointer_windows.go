//go:build windows

package input

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	inputMouse          = 0
	mouseEventMove      = 0x0001
	mouseEventLeftDown  = 0x0002
	mouseEventLeftUp    = 0x0004
	mouseEventRightDown = 0x0008
	mouseEventRightUp   = 0x0010
	windowsInputSize    = 40
)

var sendInput = windows.NewLazySystemDLL("user32.dll").NewProc("SendInput")

type windowsPointer struct{}

type mouseInput struct {
	DX        int32
	DY        int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	Padding   uint32
	ExtraInfo uintptr
}

type windowsInput struct {
	Type    uint32
	Padding uint32
	Data    [32]byte
}

func newWindowsPointer() (Adapter, error) {
	if unsafe.Sizeof(windowsInput{}) != windowsInputSize {
		return nil, fmt.Errorf("unexpected Windows INPUT size: %d", unsafe.Sizeof(windowsInput{}))
	}
	return windowsPointer{}, nil
}

func (windowsPointer) Name() string { return "windows" }

func (windowsPointer) Move(dx, dy int) error {
	return sendWindowsInputs([]windowsInput{newMouseInput(int32(dx), int32(dy), mouseEventMove)})
}

func (windowsPointer) Click(button Button) error {
	var down, up uint32
	switch button {
	case ButtonLeft:
		down, up = mouseEventLeftDown, mouseEventLeftUp
	case ButtonRight:
		down, up = mouseEventRightDown, mouseEventRightUp
	default:
		return fmt.Errorf("unsupported pointer button %q", button)
	}
	return sendWindowsInputs([]windowsInput{
		newMouseInput(0, 0, down),
		newMouseInput(0, 0, up),
	})
}

func newMouseInput(dx, dy int32, flags uint32) windowsInput {
	value := windowsInput{Type: inputMouse}
	*(*mouseInput)(unsafe.Pointer(&value.Data[0])) = mouseInput{DX: dx, DY: dy, Flags: flags}
	return value
}

func sendWindowsInputs(inputs []windowsInput) error {
	if len(inputs) == 0 {
		return nil
	}
	inserted, _, callErr := sendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	if inserted == uintptr(len(inputs)) {
		return nil
	}
	if callErr != nil && !errors.Is(callErr, windows.ERROR_SUCCESS) {
		return fmt.Errorf("SendInput inserted %d of %d events: %w", inserted, len(inputs), callErr)
	}
	return fmt.Errorf("SendInput inserted %d of %d events", inserted, len(inputs))
}
