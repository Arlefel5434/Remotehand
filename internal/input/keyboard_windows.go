//go:build windows

package input

import (
	"fmt"
	"unicode/utf16"
	"unsafe"
)

const (
	inputKeyboard       = 1
	keyEventKeyUp       = 0x0002
	keyEventUnicode     = 0x0004
	virtualKeyEnter     = 0x0D
	virtualKeyBackspace = 0x08
	virtualKeyEscape    = 0x1B
)

type keyboardInput struct {
	VirtualKey uint16
	ScanCode   uint16
	Flags      uint32
	Time       uint32
	Padding    uint32
	ExtraInfo  uintptr
}

func (windowsPointer) SendText(text string) error {
	codeUnits := utf16.Encode([]rune(text))
	inputs := make([]windowsInput, 0, len(codeUnits)*2)
	for _, codeUnit := range codeUnits {
		inputs = append(inputs,
			newKeyboardInput(0, codeUnit, keyEventUnicode),
			newKeyboardInput(0, codeUnit, keyEventUnicode|keyEventKeyUp),
		)
	}
	return sendWindowsInputs(inputs)
}

func (windowsPointer) PressKey(key Key) error {
	var virtualKey uint16
	switch key {
	case KeyEnter:
		virtualKey = virtualKeyEnter
	case KeyBackspace:
		virtualKey = virtualKeyBackspace
	case KeyEscape:
		virtualKey = virtualKeyEscape
	default:
		return fmt.Errorf("unsupported key %q", key)
	}
	return sendWindowsInputs([]windowsInput{
		newKeyboardInput(virtualKey, 0, 0),
		newKeyboardInput(virtualKey, 0, keyEventKeyUp),
	})
}

func newKeyboardInput(virtualKey, scanCode uint16, flags uint32) windowsInput {
	value := windowsInput{Type: inputKeyboard}
	*(*keyboardInput)(unsafe.Pointer(&value.Data[0])) = keyboardInput{
		VirtualKey: virtualKey,
		ScanCode:   scanCode,
		Flags:      flags,
	}
	return value
}
