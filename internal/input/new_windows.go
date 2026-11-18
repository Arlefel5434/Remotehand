//go:build windows

package input

import "fmt"

// New uses the native Windows adapter by default and permits an explicit safe
// mock mode for diagnostics.
func New(mode string) (Adapter, error) {
	switch mode {
	case "", "native":
		return namedAdapter("windows"), nil
	case "mock":
		return mockAdapter(), nil
	default:
		return nil, fmt.Errorf("unknown input mode %q", mode)
	}
}
