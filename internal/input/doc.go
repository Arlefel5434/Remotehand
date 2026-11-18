// Package input contains operating-system adapters for remote input and volume
// control.
package input

type Button string

const (
	ButtonLeft  Button = "left"
	ButtonRight Button = "right"
)

// PointerController applies relative movement and complete button clicks.
type PointerController interface {
	Move(dx, dy int) error
	Click(button Button) error
}

type Key string

const (
	KeyEnter     Key = "enter"
	KeyBackspace Key = "backspace"
	KeyEscape    Key = "escape"
)

// KeyboardController injects Unicode text or one supported control key into
// the active Windows application.
type KeyboardController interface {
	SendText(text string) error
	PressKey(key Key) error
}

type Adapter interface {
	PointerController
	KeyboardController
	Name() string
}

type namedAdapter string

func (a namedAdapter) Name() string        { return string(a) }
func (namedAdapter) Move(int, int) error   { return nil }
func (namedAdapter) Click(Button) error    { return nil }
func (namedAdapter) SendText(string) error { return nil }
func (namedAdapter) PressKey(Key) error    { return nil }

func mockAdapter() Adapter { return namedAdapter("mock") }
