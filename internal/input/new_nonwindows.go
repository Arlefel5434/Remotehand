//go:build !windows

package input

import "fmt"

// New returns the only adapter supported outside Windows: the safe mock.
func New(mode string) (Adapter, error) {
	switch mode {
	case "", "mock":
		return mockAdapter(), nil
	default:
		return nil, fmt.Errorf("input mode %q is unavailable on this platform", mode)
	}
}
