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
	const ctrl, shift, tab = 0x11, 0x10, 0x09
	var keys []uint16
	switch key {
	case KeyEnter:
		keys = []uint16{virtualKeyEnter}
	case KeyBackspace:
		keys = []uint16{virtualKeyBackspace}
	case KeyEscape:
		keys = []uint16{virtualKeyEscape}
	case "restore_tab":
		keys = []uint16{ctrl, shift, 'T'}
	case "new_tab":
		keys = []uint16{ctrl, 'T'}
	case "next_tab":
		keys = []uint16{ctrl, tab}
	case "find":
		keys = []uint16{ctrl, 'F'}
	default:
		return fmt.Errorf("unsupported key %q", key)
	}
	events := make([]windowsInput, 0, 2*len(keys))
	for _, vk := range keys {
		events = append(events, newKeyboardInput(vk, 0, 0))
	}
	for i := len(keys) - 1; i >= 0; i-- {
		events = append(events, newKeyboardInput(keys[i], 0, keyEventKeyUp))
	}
	return sendWindowsInputs(events)
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
