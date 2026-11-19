//go:build windows

package input

import "fmt"

func New(mode string) (Adapter, error) {
	switch mode {
	case "mock":
		return mockAdapter(), nil
	case "", "native":
		return newWindowsPointer()
	default:
		return nil, fmt.Errorf("unknown input mode %q", mode)
	}
}
