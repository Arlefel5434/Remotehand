//go:build windows

package input

import "fmt"

// New keeps the safe mock as the default until authenticated sessions exist.
func New(mode string) (Adapter, error) {
	switch mode {
	case "", "mock":
		return mockAdapter(), nil
	case "native":
		return newWindowsPointer()
	default:
		return nil, fmt.Errorf("unknown input mode %q", mode)
	}
}
