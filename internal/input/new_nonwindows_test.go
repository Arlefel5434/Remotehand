//go:build !windows

package input

import "testing"

func TestNewUsesMockOutsideWindows(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{"", "mock"} {
		adapter, err := New(mode)
		if err != nil {
			t.Fatalf("New(%q) returned an error: %v", mode, err)
		}
		if adapter.Name() != "mock" {
			t.Fatalf("New(%q) adapter = %q, want mock", mode, adapter.Name())
		}
	}
}

func TestNewRejectsNativeOutsideWindows(t *testing.T) {
	t.Parallel()

	if _, err := New("native"); err == nil {
		t.Fatal("New(native) returned no error")
	}
}
